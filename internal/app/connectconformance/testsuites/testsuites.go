// Copyright 2023 The Connect Authors
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

package testsuites

import (
	"embed"
	"fmt"
	"path"

	conformancev1alpha1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1alpha1"
	"github.com/bufbuild/protoyaml-go"
)

//go:embed *.yaml
var testSuiteFiles embed.FS

// LoadTestSuites returns a map of file names to contents for all
// test suite YAML files that are embedded into this package.
func LoadTestSuites() (map[string]*conformancev1alpha1.TestSuite, error) {
	testSuites := map[string]*conformancev1alpha1.TestSuite{}
	if err := scanDir(".", testSuites); err != nil {
		return nil, err
	}
	return testSuites, nil
}

func scanDir(dir string, testSuites map[string]*conformancev1alpha1.TestSuite) error {
	entries, err := testSuiteFiles.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to scan test suite data files in %s: %w", dir, err)
	}
	for _, entry := range entries {
		entryPath := path.Join(dir, entry.Name())
		if entry.IsDir() {
			if err := scanDir(entryPath, testSuites); err != nil {
				return err
			}
			continue
		}
		data, err := testSuiteFiles.ReadFile(entryPath)
		if err != nil {
			return fmt.Errorf("failed to load test suite date file %s: %w", entryPath, err)
		}
		opts := protoyaml.UnmarshalOptions{
			Path: entryPath,
		}
		suite := &conformancev1alpha1.TestSuite{}
		if err := opts.Unmarshal(data, suite); err != nil {
			return err
		}
		testSuites[entryPath] = suite
	}
	return nil
}
