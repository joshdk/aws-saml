// Copyright Josh Komoroske. All rights reserved.
// Use of this source code is governed by the MIT license,
// a copy of which can be found in the LICENSE.txt file.
// SPDX-License-Identifier: MIT

package server

import (
	"bytes"
	_ "embed"
	"text/template"
	"time"

	"jdk.sh/meta"
)

//go:embed templates/index.html
var html string

// index contains the result of templating index.html.
var index []byte //nolint:gochecknoglobals

func init() { //nolint:gochecknoinits
	// properties is a set of named values used when executing a template.
	properties := map[string]string{
		"Date":    meta.DateFormat(time.RFC3339),
		"SHA":     meta.ShortSHA(),
		"Version": meta.Version(),
	}

	tpl, err := template.New("templates/index.html").Parse(html)
	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, properties); err != nil {
		panic(err)
	}

	index = buf.Bytes()
}
