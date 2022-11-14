// Copyright Josh Komoroske. All rights reserved.
// Use of this source code is governed by the MIT license,
// a copy of which can be found in the LICENSE.txt file.
// SPDX-License-Identifier: MIT

package server

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed static
var staticFS embed.FS

// static is a http handler that can be used to serve static web app content.
var static http.Handler //nolint:gochecknoglobals

func init() { //nolint:gochecknoinits
	// Re-root the filesystem embedded from the static directory.
	root, err := fs.Sub(staticFS, "static")
	if err != nil {
		panic(err)
	}

	static = http.FileServer(http.FS(root))
}
