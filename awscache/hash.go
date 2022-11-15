// Copyright Josh Komoroske. All rights reserved.
// Use of this source code is governed by the MIT license,
// a copy of which can be found in the LICENSE.txt file.
// SPDX-License-Identifier: MIT

package awscache

import (
	"crypto/sha1" //nolint:gosec
	"encoding/hex"
	"encoding/json"
)

func hash(name string, properties map[string]string) (string, error) {
	// key is an object that will be marshaled to json and hashed inorder to
	// derive a consistent lookup hash.
	key := struct {
		Name       string            `json:"name"`
		Properties map[string]string `json:"properties"`
	}{
		Name:       name,
		Properties: properties,
	}

	// Hash the above key in order to derive a consistent identifier for the
	// given input. This hash is used only for lookups and is not intended to
	// be cryptographically secure.
	sha := sha1.New() //nolint:gosec

	if err := json.NewEncoder(sha).Encode(key); err != nil {
		return "", err
	}

	return hex.EncodeToString(sha.Sum(nil)), nil
}
