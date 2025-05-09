// Copyright Josh Komoroske. All rights reserved.
// Use of this source code is governed by the MIT license,
// a copy of which can be found in the LICENSE.txt file.
// SPDX-License-Identifier: MIT

package server

import (
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"io"
)

// formatSAMLResponse takes a base64 encoded SAML assertion body, base64
// decodes it, then pretty-prints the contained XML document.
func formatSAMLResponse(raw string, writer io.Writer) error {
	decoder := xml.NewDecoder(base64.NewDecoder(base64.StdEncoding, bytes.NewReader([]byte(raw))))

	encoder := xml.NewEncoder(writer)
	encoder.Indent("", "  ")

	// We have to round-trip through decoding+(re)encoding every token since it
	// isn't possible to unmarshal into an interface{}.
	for {
		token, err := decoder.RawToken()

		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		if err := encoder.EncodeToken(token); err != nil {
			return err
		}
	}

	if err := encoder.Flush(); err != nil {
		return err
	}

	// Append a single newline character to improve with copying the formatted
	// XML body from the console.
	_, err := writer.Write([]byte("\n"))

	return err
}
