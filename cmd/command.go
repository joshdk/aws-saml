// Copyright Josh Komoroske. All rights reserved.
// Use of this source code is governed by the MIT license,
// a copy of which can be found in the LICENSE.txt file.
// SPDX-License-Identifier: MIT

// Package cmd contains functionality for supporting the aws-saml cli.
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/processcreds"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/joshdk/aws-saml/awscache"
	"github.com/joshdk/aws-saml/server"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
	"jdk.sh/meta"
)

type flags struct {
	// duration is how long the role session should last before expiring.
	duration time.Duration

	// idp is the URL used to begin idp initiated login. A browser will be
	// opened to this URL and the user will be prompted to login via the SAML
	// IdP.
	idp string

	// listen is the local address to start an HTTP server on to listen for the
	// SAML assertion POST. Once the SAML assertion is POSTed (or if the given
	// timeout is exceeded) the server will shut down automatically.
	listen string

	// principal is the arn of the SAML provider in IAM that describes the IdP.
	principal string

	// role is the arn of the role that the caller is assuming.
	role string

	// timeout is how long to wait for SAML assertion redirect. If the user
	// takes too long to login via the SAML IdP, then the command will
	// automatically fail after this duration.
	timeout time.Duration

	// userAgent is the user agent to use when making API calls.
	userAgent string
}

// Command returns a complete handler for the aws-saml cli.
func Command() *cobra.Command { //nolint:funlen
	var flags flags

	cmd := &cobra.Command{ //nolint:exhaustruct
		Use:     "aws-saml",
		Long:    "aws-saml - Generate AWS credentials from a SAML IdP login",
		Version: "-",

		SilenceUsage:  true,
		SilenceErrors: true,

		RunE: func(cmd *cobra.Command, _ []string) error {
			// Create a context that is cancelled after the given timeout.
			ctx, cancel := context.WithTimeout(cmd.Context(), flags.timeout)
			defer cancel()

			// Attempt to load cached credentials so that we don't need to
			// login to the SAML IdP repeatedly.
			properties := map[string]string{
				"idp":       flags.idp,
				"principal": flags.principal,
				"duration":  flags.duration.String(),
				"role":      flags.role,
			}
			result := awscache.LoadAssumeRoleWithSAMLOutput("aws-saml", "common", properties)

			// If no credentials were retrieved, then we need to go through the
			// whole SAML IdP login flow.
			if result == nil {
				// Start the local server which will handle the assertion callback
				// from the SAML IdP.
				loginURL, waitForSAMLResponse := server.Start(ctx, flags.listen, flags.idp)

				// Launch a browser to the login url to initiate the SAML login flow.
				if err := browser.OpenURL(loginURL); err != nil {
					return err
				}

				// Wait for the user to complete the login flow.
				samlResponse, err := waitForSAMLResponse()
				if err != nil {
					return err
				}

				// Create a new default session.
				sess, err := session.NewSession()
				if err != nil {
					return err
				}

				// Configure a new STS client.
				client := sts.New(sess)
				client.Handlers.Build.PushBack(request.WithSetRequestHeaders(map[string]string{"User-Agent": flags.userAgent}))

				// Actually assume the given role with our SAML response.
				input := sts.AssumeRoleWithSAMLInput{ //nolint:exhaustruct
					DurationSeconds: aws.Int64(int64(flags.duration.Seconds())),
					PrincipalArn:    aws.String(flags.principal),
					RoleArn:         aws.String(flags.role),
					SAMLAssertion:   aws.String(samlResponse),
				}
				result, err = client.AssumeRoleWithSAML(&input)
				if err != nil {
					return err
				}

				// Attempt to store credentials into the cache.
				awscache.StoreAssumeRoleWithSAMLOutput("aws-saml", "common", properties, result)
			}

			// Configure a credential process response with that the aws cli
			// can consume our new credentials.
			output := processcreds.CredentialProcessResponse{
				Version:         1,
				AccessKeyID:     aws.StringValue(result.Credentials.AccessKeyId),
				SecretAccessKey: aws.StringValue(result.Credentials.SecretAccessKey),
				SessionToken:    aws.StringValue(result.Credentials.SessionToken),
				Expiration:      result.Credentials.Expiration,
			}

			// Marshal the credential process response to stdout so that the
			// aws cli can read it.
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")

			return enc.Encode(output)
		},
	}

	// Define -d/--duration flag.
	cmd.Flags().DurationVarP(&flags.duration, "duration", "d",
		12*time.Hour,
		"duration of the role session")

	// Define -i/--idp flag.
	cmd.Flags().StringVarP(&flags.idp, "idp", "i",
		"",
		"url to use for idp initiated login")
	cmd.MarkFlagRequired("idp") //nolint

	// Define -l/--listen flag.
	cmd.Flags().StringVarP(&flags.listen, "listen", "l",
		"",
		"local address to listen for SAML assertion POST")
	cmd.MarkFlagRequired("listen") //nolint

	// Define -p/--principal flag.
	cmd.Flags().StringVarP(&flags.principal, "principal", "p",
		"",
		"arn of the SAML provider in IAM that describes the IdP")
	cmd.MarkFlagRequired("principal") //nolint

	// Define -r/--role flag.
	cmd.Flags().StringVarP(&flags.role, "role", "r",
		"",
		"arn of the role that the caller is assuming")
	cmd.MarkFlagRequired("role") //nolint

	// Define -t/--timeout flag.
	cmd.Flags().DurationVarP(&flags.timeout, "timeout", "t",
		5*time.Minute,
		"duration to wait for SAML assertion")

	// Define -A/--user-agent flag.
	cmd.Flags().StringVarP(&flags.userAgent, "user-agent", "A",
		versionFmt("joshdk/aws-saml", " %s (%s)", meta.Version(), meta.ShortSHA()),
		"user agent to use for http requests")

	// Add a custom usage footer template.
	cmd.SetUsageTemplate(cmd.UsageTemplate() + versionFmt(
		"\nInfo:\n"+
			"  https://github.com/joshdk/aws-saml\n",
		"  %s (%s) built on %v\n",
		meta.Version(), meta.ShortSHA(), meta.DateFormat(time.RFC3339),
	))

	// Set a custom version template.
	cmd.SetVersionTemplate(versionFmt(
		"homepage: https://github.com/joshdk/aws-saml\n"+
			"author:   Josh Komoroske\n"+
			"license:  MIT\n",
		"version:  %s\n"+
			"sha:      %s\n"+
			"date:     %s\n",
		meta.Version(), meta.ShortSHA(), meta.DateFormat(time.RFC3339),
	))

	return cmd
}

// versionFmt returns the given literal, as well as a formatted string if
// version metadata is set.
func versionFmt(literal, format string, a ...interface{}) string {
	if meta.Version() == "" {
		return literal
	}

	return literal + fmt.Sprintf(format, a...)
}
