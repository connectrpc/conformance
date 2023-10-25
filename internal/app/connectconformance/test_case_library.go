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

package connectconformance

import (
	"errors"
	"fmt"
	"path"
	"sort"

	"connectrpc.com/conformance/internal/app/connectconformance/testsuites"
	conformancev1alpha1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1alpha1"
	"google.golang.org/protobuf/proto"
)

//nolint:gochecknoglobals
var (
	allProtocols    = allValues[conformancev1alpha1.Protocol](conformancev1alpha1.Protocol_name)
	allHTTPVersions = allValues[conformancev1alpha1.HTTPVersion](conformancev1alpha1.HTTPVersion_name)
	allCodecs       = allValues[conformancev1alpha1.Codec](conformancev1alpha1.Codec_name)
	allCompressions = allValues[conformancev1alpha1.Compression](conformancev1alpha1.Compression_name)
	allStreamTypes  = allValues[conformancev1alpha1.StreamType](conformancev1alpha1.StreamType_name)
)

type testCaseLibrary struct {
	testCases     map[string]*conformancev1alpha1.TestCase
	casesByServer map[serverInstance][]*conformancev1alpha1.TestCase
}

type serverInstance struct {
	protocol    conformancev1alpha1.Protocol
	httpVersion conformancev1alpha1.HTTPVersion
	useTLS      bool
}

func loadTestCases(configCases []configCase, mode conformancev1alpha1.TestSuite_TestMode) (*testCaseLibrary, error) {
	configCaseSet := make(map[configCase]struct{}, len(configCases))
	for _, c := range configCases {
		configCaseSet[c] = struct{}{}
	}

	allSuites, err := testsuites.LoadTestSuites()
	if err != nil {
		return nil, err
	}
	lib := &testCaseLibrary{
		testCases: map[string]*conformancev1alpha1.TestCase{},
	}
	suitesIndex := make(map[string]string, len(allSuites))
	for file, suite := range allSuites {
		if suite.Name == "" {
			return nil, fmt.Errorf("%s defines a suite with no name", file)
		}
		if len(suite.TestCases) == 0 {
			return nil, fmt.Errorf("%s defines a suite %s that has no test cases", file, suite.Name)
		}
		if existingFile, exists := suitesIndex[suite.Name]; exists {
			return nil, fmt.Errorf("both %s and %s define a suite named %s", file, existingFile, suite.Name)
		}
		suitesIndex[suite.Name] = file
		if suite.Mode != conformancev1alpha1.TestSuite_TEST_MODE_UNSPECIFIED && suite.Mode != mode {
			continue // skip it
		}
		if err := lib.expandSuite(suite, configCaseSet); err != nil {
			return nil, err
		}
	}

	if len(lib.testCases) == 0 {
		return nil, errors.New("no test cases apply to current configuration")
	}
	lib.groupTestCases()
	return lib, nil
}

func (lib *testCaseLibrary) expandSuite(suite *conformancev1alpha1.TestSuite, configCases map[configCase]struct{}) error {
	protocols := suite.RelevantProtocols
	if len(protocols) == 0 {
		protocols = allProtocols
	}
	for _, protocol := range protocols {
		httpVersions := suite.RelevantHttpVersions
		if len(httpVersions) == 0 {
			httpVersions = allHTTPVersions
		}
		for _, httpVersion := range httpVersions {
			codecs := suite.RelevantCodecs
			if len(codecs) == 0 {
				codecs = allCodecs
			}
			for _, codec := range codecs {
				compressions := suite.RelevantCompressions
				if len(compressions) == 0 {
					compressions = allCompressions
				}
				for _, compression := range compressions {
					for _, streamType := range allStreamTypes {
						cfgCase := configCase{
							Version:            httpVersion,
							Protocol:           protocol,
							Codec:              codec,
							Compression:        compression,
							StreamType:         streamType,
							UseTLS:             suite.ReliesOnTls,
							UseConnectGET:      suite.ReliesOnConnectGet,
							ConnectVersionMode: suite.ConnectVersionMode,
						}
						if _, ok := configCases[cfgCase]; ok {
							namePrefix := generateTestCasePrefix(suite, cfgCase)
							if err := lib.expandCases(cfgCase, namePrefix, suite.TestCases); err != nil {
								return err
							}
						}
					}
				}
			}
		}
	}
	return nil
}

func (lib *testCaseLibrary) expandCases(cfgCase configCase, namePrefix []string, testCases []*conformancev1alpha1.TestCase) error {
	for _, testCase := range testCases {
		if testCase.Request.StreamType != cfgCase.StreamType {
			continue
		}
		name := path.Join(append(namePrefix, testCase.Request.TestName)...)
		if _, exists := lib.testCases[name]; exists {
			return fmt.Errorf("test case library includes duplicate definition for %v", name)
		}
		testCase := proto.Clone(testCase).(*conformancev1alpha1.TestCase) //nolint:errcheck,forcetypeassert
		testCase.Request.TestName = name
		if cfgCase.UseTLS {
			// to be replaced with actual cert provided by server
			testCase.Request.ServerTlsCert = []byte("PLACEHOLDER")
		}
		testCase.Request.HttpVersion = cfgCase.Version
		testCase.Request.Protocol = cfgCase.Protocol
		testCase.Request.Codec = cfgCase.Codec
		testCase.Request.Compression = cfgCase.Compression
		lib.testCases[name] = testCase
	}
	return nil
}

func (lib *testCaseLibrary) groupTestCases() {
	lib.casesByServer = map[serverInstance][]*conformancev1alpha1.TestCase{}
	for _, testCase := range lib.testCases {
		svr := serverInstanceForCase(testCase)
		lib.casesByServer[svr] = append(lib.casesByServer[svr], testCase)
	}
}

func generateTestCasePrefix(suite *conformancev1alpha1.TestSuite, cfgCase configCase) []string {
	components := make([]string, 1, 5)
	components = append(components, suite.Name)
	if len(suite.RelevantHttpVersions) != 1 {
		components = append(components, fmt.Sprintf("HTTPVersion:%d", cfgCase.Version))
	}
	if len(suite.RelevantProtocols) != 1 {
		components = append(components, fmt.Sprintf("Protocol:%s", cfgCase.Protocol))
	}
	if len(suite.RelevantCodecs) != 1 {
		components = append(components, fmt.Sprintf("Codec:%s", cfgCase.Codec))
	}
	if len(suite.RelevantCompressions) != 1 {
		components = append(components, fmt.Sprintf("Compression:%s", cfgCase.Compression))
	}
	return components
}

func serverInstanceForCase(testCase *conformancev1alpha1.TestCase) serverInstance {
	return serverInstance{
		protocol:    testCase.Request.Protocol,
		httpVersion: testCase.Request.HttpVersion,
		useTLS:      len(testCase.Request.ServerTlsCert) > 0,
	}
}

func allValues[T ~int32](m map[int32]string) []T {
	vals := make([]T, 0, len(m))
	for k := range m {
		if k == 0 {
			continue
		}
		vals = append(vals, T(k))
	}
	sort.Slice(vals, func(i, j int) bool {
		return vals[i] < vals[j]
	})
	return vals
}
