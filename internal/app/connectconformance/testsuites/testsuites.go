// Copyright 2023-2024 The Connect Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package testsuites contains embedded test suite data used when running
// conformance tests. While it is possible to point the test runner at
// other files, by default, it will use the test cases embedded in this
// package. This package embeds all *.yaml files in this folder.
package testsuites

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

//go:embed *.yaml
var testSuiteFS embed.FS

// LoadTestSuites returns a map of file paths to their contents for the
// embedded corpus of test suites. The file names are the names of test
// suite YAML files.
func LoadTestSuites() (map[string][]byte, error) {
	testSuites := map[string][]byte{}
	err := fs.WalkDir(testSuiteFS, ".", func(currentPath string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || strings.ToLower(path.Ext(entry.Name())) != ".yaml" {
			return nil
		}
		data, err := testSuiteFS.ReadFile(currentPath)
		if err != nil {
			return fmt.Errorf("failed to load test suite data file %s: %w", currentPath, err)
		}
		testSuites[currentPath] = data
		return nil
	})
	if err != nil {
		return nil, err
	}
	return testSuites, nil
}

// LoadTestSuitesFromFiles loads the test suites specified in the given paths.
// If a provided path is not found, is a directory, or is not a YAML file, the
// function will return an error.
func LoadTestSuitesFromFiles(paths []string) (map[string][]byte, error) {
	testSuites := make(map[string][]byte, len(paths))
	for _, path := range paths {
		testFile, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		if filepath.Ext(path) != ".yaml" {
			return nil, fmt.Errorf("failed to load test data file: %s. file is not in YAML format", path)
		}
		testSuites[path] = testFile
	}
	return testSuites, nil
}
