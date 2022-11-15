// Copyright Josh Komoroske. All rights reserved.
// Use of this source code is governed by the MIT license,
// a copy of which can be found in the LICENSE.txt file.
// SPDX-License-Identifier: MIT

package awscache

import (
	"fmt"
	"os"
	"path/filepath"
)

func cacheDirname(namespace string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".aws", "cli", "cache", namespace), nil
}

func cacheFilename(name string, properties map[string]string) (string, error) {
	prefix, err := hash(name, properties)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s.json", prefix), nil
}
