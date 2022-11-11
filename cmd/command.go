// Copyright Josh Komoroske. All rights reserved.
// Use of this source code is governed by the MIT license,
// a copy of which can be found in the LICENSE.txt file.
// SPDX-License-Identifier: MIT

// Package cmd contains functionality for supporting the aws-saml cli.
package cmd

import (
	"github.com/spf13/cobra"
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

	return cmd
}
