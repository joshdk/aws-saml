// Copyright Josh Komoroske. All rights reserved.
// Use of this source code is governed by the MIT license,
// a copy of which can be found in the LICENSE.txt file.
// SPDX-License-Identifier: MIT

// Package awscache implements a cache for various types of AWS credentials.
package awscache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go/service/sts"
)

type standard struct {
	Credentials struct {
		Expiration time.Time `json:"Expiration"`
	} `json:"Credentials"`
}

// Store attempts to insert the given object into the cache. The given name and
// properties are used as a lookup key into the cache and must match what will
// be provided to the following Load call.
func Store(namespace string, name string, properties map[string]string, obj interface{}) {
	dirname, err := cacheDirname(namespace)
	if err != nil {
		return
	}

	filename, err := cacheFilename(name, properties)
	if err != nil {
		return
	}

	// Create the cache directory.
	if err := os.MkdirAll(dirname, 0o755); err != nil { //nolint:gosec
		return
	}

	cacheFilename := filepath.Join(dirname, filename)

	file, err := os.Create(cacheFilename) //nolint:gosec
	if err != nil {
		return
	}

	// Marshal the given object into the cache file.
	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	enc.Encode(obj) //nolint
}

// Load attempts to fetch the given object from the cache. The given name and
// properties are used as a lookup key into the cache and must match what was
// provided to the preceding Store call.
func Load(namespace string, name string, properties map[string]string, obj interface{}) bool {
	dirname, err := cacheDirname(namespace)
	if err != nil {
		return false
	}

	filename, err := cacheFilename(name, properties)
	if err != nil {
		return false
	}

	cacheFilename := filepath.Join(dirname, filename)

	data, err := os.ReadFile(cacheFilename) //nolint:gosec
	if err != nil {
		return false
	}

	// Unmarshal the cache file into a format that can be used to check if the
	// credentials are expired.
	var std standard
	if err := json.Unmarshal(data, &std); err != nil {
		defer os.RemoveAll(cacheFilename) //nolint:errcheck

		return false
	}

	// If the credentials are expired, then quit out.
	if std.Credentials.Expiration.Before(time.Now()) {
		defer os.RemoveAll(cacheFilename) //nolint:errcheck

		return false
	}

	if err := json.Unmarshal(data, &obj); err != nil {
		return false
	}

	return true
}

// StoreAssumeRoleWithSAMLOutput stores a sts.AssumeRoleWithSAMLOutput into the cache.
func StoreAssumeRoleWithSAMLOutput(namespace string, name string, properties map[string]string, obj *sts.AssumeRoleWithSAMLOutput) { //nolint:lll
	Store(namespace, name, properties, obj)
}

// LoadAssumeRoleWithSAMLOutput loads a sts.AssumeRoleWithSAMLOutput from the cache.
func LoadAssumeRoleWithSAMLOutput(namespace string, name string, properties map[string]string) *sts.AssumeRoleWithSAMLOutput { //nolint:lll
	var obj *sts.AssumeRoleWithSAMLOutput
	if Load(namespace, name, properties, &obj) {
		return obj
	}

	return nil
}
