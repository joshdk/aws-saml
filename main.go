// Copyright Josh Komoroske. All rights reserved.
// Use of this source code is governed by the MIT license,
// a copy of which can be found in the LICENSE.txt file.
// SPDX-License-Identifier: MIT

// Package main contains the entry point for aws-saml.
package main

import (
	"fmt"
	"os"

	"github.com/joshdk/aws-saml/cmd"
)

func main() {
	if err := cmd.Command().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "aws-saml:", err)
		os.Exit(1)
	}
}
