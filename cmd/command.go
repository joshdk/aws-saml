// Copyright Josh Komoroske. All rights reserved.
// Use of this source code is governed by the MIT license,
// a copy of which can be found in the LICENSE.txt file.
// SPDX-License-Identifier: MIT

// Package cmd contains functionality for supporting the aws-saml cli.
package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"jdk.sh/meta"
)

// Command returns a complete handler for the aws-saml cli.
func Command() *cobra.Command {
	cmd := &cobra.Command{ //nolint:exhaustruct
		Use:     "aws-saml",
		Long:    "aws-saml - Generate AWS credentials from a SAML IdP login",
		Version: "-",

		SilenceUsage:  true,
		SilenceErrors: true,

		RunE: func(*cobra.Command, []string) error {
			return nil
		},
	}

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
