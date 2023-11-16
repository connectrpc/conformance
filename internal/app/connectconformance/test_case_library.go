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
	"strings"

	conformancev2 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v2"
	"github.com/bufbuild/protoyaml-go"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

//nolint:gochecknoglobals
var (
	allProtocols    = allValues[conformancev2.Protocol](conformancev2.Protocol_name)
	allHTTPVersions = allValues[conformancev2.HTTPVersion](conformancev2.HTTPVersion_name)
	allCodecs       = allValues[conformancev2.Codec](conformancev2.Codec_name)
	allCompressions = allValues[conformancev2.Compression](conformancev2.Compression_name)
	allStreamTypes  = allValues[conformancev2.StreamType](conformancev2.StreamType_name)
)

// testCaseLibrary is the set of all applicable test cases for a run
// of the conformance tests.
type testCaseLibrary struct {
	testCases     map[string]*conformancev2.TestCase
	casesByServer map[serverInstance][]*conformancev2.TestCase
}

// newTestCaseLibrary creates a new resolved set of test cases by applying
// the given test suite configuration to the given config cases that are
// applicable to the current run of conformance tests.
func newTestCaseLibrary(
	allSuites map[string]*conformancev2.TestSuite,
	configCases []configCase,
	mode conformancev2.TestSuite_TestMode,
) (*testCaseLibrary, error) {
	configCaseSet := make(map[configCase]struct{}, len(configCases))
	for _, c := range configCases {
		configCaseSet[c] = struct{}{}
	}
	lib := &testCaseLibrary{
		testCases: map[string]*conformancev2.TestCase{},
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
		if suite.Mode != conformancev2.TestSuite_TEST_MODE_UNSPECIFIED && suite.Mode != mode {
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

func (lib *testCaseLibrary) expandSuite(suite *conformancev2.TestSuite, configCases map[configCase]struct{}) error {
	if suite.ReliesOnTlsClientCerts && !suite.ReliesOnTls {
		return fmt.Errorf("suite %q is misconfigured: it relies on TLS client certs but not TLS", suite.Name)
	}
	if suite.ReliesOnConnectGet && !only(suite.RelevantProtocols, conformancev2.Protocol_PROTOCOL_CONNECT) {
		return fmt.Errorf("suite %q is misconfigured: it relies on Connect GET support, but has unexpected relevant protocols: %v", suite.Name, suite.RelevantProtocols)
	}
	if suite.ConnectVersionMode == conformancev2.TestSuite_CONNECT_VERSION_MODE_IGNORE && !only(suite.RelevantProtocols, conformancev2.Protocol_PROTOCOL_CONNECT) {
		return fmt.Errorf("suite %q is misconfigured: it ignores Connect Version headers, but has unexpected relevant protocols: %v", suite.Name, suite.RelevantProtocols)
	}
	if suite.ConnectVersionMode == conformancev2.TestSuite_CONNECT_VERSION_MODE_REQUIRE && !only(suite.RelevantProtocols, conformancev2.Protocol_PROTOCOL_CONNECT) {
		return fmt.Errorf("suite %q is misconfigured: it requires Connect Version headers, but has unexpected relevant protocols: %v", suite.Name, suite.RelevantProtocols)
	}
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
							Version:                httpVersion,
							Protocol:               protocol,
							Codec:                  codec,
							Compression:            compression,
							StreamType:             streamType,
							UseTLS:                 suite.ReliesOnTls,
							UseTLSClientCerts:      suite.ReliesOnTlsClientCerts,
							UseConnectGET:          suite.ReliesOnConnectGet,
							ConnectVersionMode:     suite.ConnectVersionMode,
							UseMessageReceiveLimit: suite.ReliesOnMessageReceiveLimit,
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

func (lib *testCaseLibrary) expandCases(cfgCase configCase, namePrefix []string, testCases []*conformancev2.TestCase) error {
	for _, testCase := range testCases {
		if testCase.Request.StreamType != cfgCase.StreamType {
			continue
		}
		name := path.Join(append(namePrefix, testCase.Request.TestName)...)
		if _, exists := lib.testCases[name]; exists {
			return fmt.Errorf("test case library includes duplicate definition for %v", name)
		}
		testCase := proto.Clone(testCase).(*conformancev2.TestCase) //nolint:errcheck,forcetypeassert
		testCase.Request.TestName = name
		if cfgCase.UseTLS {
			// to be replaced with actual cert provided by server
			testCase.Request.ServerTlsCert = []byte("PLACEHOLDER")
			if cfgCase.UseTLSClientCerts {
				testCase.Request.ClientTlsCreds = &conformancev2.ClientCompatRequest_TLSCreds{
					Key:  []byte("PLACEHOLDER"),
					Cert: []byte("PLACEHOLDER"),
				}
			} else {
				testCase.Request.ClientTlsCreds = nil
			}
		} else {
			testCase.Request.ServerTlsCert = nil
			testCase.Request.ClientTlsCreds = nil
		}
		testCase.Request.HttpVersion = cfgCase.Version
		testCase.Request.Protocol = cfgCase.Protocol
		testCase.Request.Codec = cfgCase.Codec
		testCase.Request.Compression = cfgCase.Compression
		// We always set this. If client-under-test does not support it, we just
		// won't run the test cases that verify that it's enforced.
		// Note that we always use a larger limit on the client so that when
		// we test the server limit, even when close to the server's limit, the
		// response (which echoes back the request data) won't exceed client limit.
		testCase.Request.MessageReceiveLimit = 1024 * 1024 // 1 MB
		lib.testCases[name] = testCase
	}
	return nil
}

func (lib *testCaseLibrary) groupTestCases() {
	lib.casesByServer = map[serverInstance][]*conformancev2.TestCase{}
	for _, testCase := range lib.testCases {
		svr := serverInstanceForCase(testCase)
		lib.casesByServer[svr] = append(lib.casesByServer[svr], testCase)
	}
}

// serverInstance identifies the properties of a server process, so tests
// can be grouped by target server process.
type serverInstance struct {
	protocol          conformancev2.Protocol
	httpVersion       conformancev2.HTTPVersion
	useTLS            bool
	useTLSClientCerts bool
}

func serverInstanceForCase(testCase *conformancev2.TestCase) serverInstance {
	return serverInstance{
		protocol:          testCase.Request.Protocol,
		httpVersion:       testCase.Request.HttpVersion,
		useTLS:            len(testCase.Request.ServerTlsCert) > 0,
		useTLSClientCerts: testCase.Request.ClientTlsCreds != nil,
	}
}

// parseTestSuites processes the given file contents. The given map is keyed
// by test file name. Each entry's value is the contents of the named file.
// The given argument often represents the embedded test suite data. Also
// see testsuites.LoadTestSuites.
func parseTestSuites(testFileData map[string][]byte) (map[string]*conformancev2.TestSuite, error) {
	allSuites := make(map[string]*conformancev2.TestSuite, len(testFileData))
	for testFilePath, data := range testFileData {
		opts := protoyaml.UnmarshalOptions{
			Path: testFilePath,
		}
		suite := &conformancev2.TestSuite{}
		if err := opts.Unmarshal(data, suite); err != nil {
			return nil, ensureFileName(err, testFilePath)
		}
		for _, testCase := range suite.TestCases {
			err := populateExpectedResponse(testCase)
			if err != nil {
				return nil, fmt.Errorf("%s: failed to compute expected response for test case %q: %w",
					testFilePath, testCase.Request.TestName, err)
			}
		}
		allSuites[testFilePath] = suite
	}
	return allSuites, nil
}

// populateExpectedResponse populates the response we expected to get back from the server
// by examining the requests we sent.
func populateExpectedResponse(testCase *conformancev2.TestCase) error {
	// If an expected response was already provided, return and use that.
	// This allows for overriding this function with explicit values in the yaml file.
	if testCase.ExpectedResponse != nil {
		return nil
	}
	// TODO - This is just a temporary constraint to protect against panics for now.
	// Eventually, we want to be able to test client and bidi streams where there are no request messages.
	// The potential plan is for server impls to produce (and the code below to expect) a single response
	// message in this situation, where the response data value is some fixed string (such as "no response definition")
	// and whose request info will still be present, but we expect it to indicate zero request messages.
	if len(testCase.Request.RequestMessages) == 0 {
		return errors.New("at least one request is required")
	}

	switch testCase.Request.StreamType {
	case conformancev2.StreamType_STREAM_TYPE_FULL_DUPLEX_BIDI_STREAM,
		conformancev2.StreamType_STREAM_TYPE_HALF_DUPLEX_BIDI_STREAM,
		conformancev2.StreamType_STREAM_TYPE_SERVER_STREAM:
		return populateExpectedStreamResponse(testCase)

	case conformancev2.StreamType_STREAM_TYPE_UNARY,
		conformancev2.StreamType_STREAM_TYPE_CLIENT_STREAM:
		return populateExpectedUnaryResponse(testCase)

	case conformancev2.StreamType_STREAM_TYPE_UNSPECIFIED:
		return errors.New("stream type is required")
	default:
		return fmt.Errorf("stream type %s is not supported", testCase.Request.StreamType)
	}
}

// populates the expected response for a unary test case.
func populateExpectedUnaryResponse(testCase *conformancev2.TestCase) error {
	req := testCase.Request.RequestMessages[0]
	// First, find the response definition that the client instructed the server to return
	concreteReq, err := req.UnmarshalNew()
	if err != nil {
		return err
	}
	type unaryResponseDefiner interface {
		GetResponseDefinition() *conformancev2.UnaryResponseDefinition
	}

	definer, ok := concreteReq.(unaryResponseDefiner)
	if !ok {
		return fmt.Errorf("%T is not a unary test case", concreteReq)
	}

	// TODO - Need to define this better in the protos and tests as to how services should
	// behave if no responses are specified. The behavior right now differs for unary vs. streaming
	// If no responses are specified for unary, the service will still return a response with the
	// request information inside (but none of the response information since it wasn't provided)
	// But streaming endpoints don't return a single response and instead return responses via sending
	// on a stream. But if no responses are specified in the request, the streams don't send anything outbound
	// so there's no way to relay this to a client. So right now, streaming endpoints simply expect an empty
	// ClientResponseResult if no response definition is provided
	def := definer.GetResponseDefinition()
	if def == nil {
		testCase.ExpectedResponse = &conformancev2.ClientResponseResult{
			Payloads: []*conformancev2.ConformancePayload{
				{
					RequestInfo: &conformancev2.ConformancePayload_RequestInfo{
						RequestHeaders: testCase.Request.RequestHeaders,
						Requests:       testCase.Request.RequestMessages,
					},
				},
			},
		}
		return nil
	}

	// Server should have echoed back all specified headers and trailers
	expected := &conformancev2.ClientResponseResult{
		ResponseHeaders:  def.ResponseHeaders,
		ResponseTrailers: def.ResponseTrailers,
	}

	switch respType := def.Response.(type) {
	case *conformancev2.UnaryResponseDefinition_Error:
		// If an error was specified, it should be returned in the response
		expected.Error = respType.Error
	case *conformancev2.UnaryResponseDefinition_ResponseData, nil:
		// If response data was specified for the response (or nothing at all),
		// the server should echo back the request message and headers in the response
		payload := &conformancev2.ConformancePayload{
			RequestInfo: &conformancev2.ConformancePayload_RequestInfo{
				RequestHeaders: testCase.Request.RequestHeaders,
				Requests:       testCase.Request.RequestMessages,
			},
		}
		// If response data was specified for the response, it should be returned
		if respType, ok := respType.(*conformancev2.UnaryResponseDefinition_ResponseData); ok {
			payload.Data = respType.ResponseData
		}
		expected.Payloads = []*conformancev2.ConformancePayload{payload}
	default:
		return fmt.Errorf("provided UnaryRequest.Response has an unexpected type %T", respType)
	}

	testCase.ExpectedResponse = expected
	return nil
}

// populates the expected response for a streaming test case.
func populateExpectedStreamResponse(testCase *conformancev2.TestCase) error {
	req := testCase.Request.RequestMessages[0]
	// First, find the response definition that the client instructed the
	// server to return
	concreteReq, err := req.UnmarshalNew()
	if err != nil {
		return err
	}
	type streamResponseDefiner interface {
		GetResponseDefinition() *conformancev2.StreamResponseDefinition
	}

	definer, ok := concreteReq.(streamResponseDefiner)
	if !ok {
		return fmt.Errorf(
			"TestCase %s contains a request message of type %T, which is not a streaming request",
			testCase.Request.TestName,
			concreteReq,
		)
	}

	def := definer.GetResponseDefinition()
	if def == nil {
		testCase.ExpectedResponse = &conformancev2.ClientResponseResult{}
		return nil
	}

	// Server should have echoed back all specified headers, trailers, and errors
	expected := &conformancev2.ClientResponseResult{
		ResponseHeaders:  def.ResponseHeaders,
		ResponseTrailers: def.ResponseTrailers,
		Error:            def.Error,
	}

	// There should be one payload for every ResponseData the client specified
	expected.Payloads = make([]*conformancev2.ConformancePayload, len(def.ResponseData))

	for idx, data := range def.ResponseData {
		expected.Payloads[idx] = &conformancev2.ConformancePayload{
			Data: data,
		}
		switch testCase.Request.StreamType { //nolint:exhaustive
		case conformancev2.StreamType_STREAM_TYPE_SERVER_STREAM,
			conformancev2.StreamType_STREAM_TYPE_HALF_DUPLEX_BIDI_STREAM:
			// For server streams and half duplex bidi streams, all request information
			// specified should only be echoed back in the first response
			if idx == 0 {
				expected.Payloads[idx].RequestInfo = &conformancev2.ConformancePayload_RequestInfo{
					RequestHeaders: testCase.Request.RequestHeaders,
					Requests:       testCase.Request.RequestMessages,
				}
			}
		case conformancev2.StreamType_STREAM_TYPE_FULL_DUPLEX_BIDI_STREAM:
			// For a full duplex stream, the first request should be echoed back in the first
			// payload. The second should be echoed back in the second payload, etc. (i.e. a ping pong interaction)
			expected.Payloads[idx].RequestInfo = &conformancev2.ConformancePayload_RequestInfo{
				// RequestHeaders: testCase.Request.RequestHeaders,
				Requests: []*anypb.Any{testCase.Request.RequestMessages[idx]},
			}
			if idx == 0 {
				expected.Payloads[idx].RequestInfo.RequestHeaders = testCase.Request.RequestHeaders
			}
		}
	}
	testCase.ExpectedResponse = expected
	return nil
}

func generateTestCasePrefix(suite *conformancev2.TestSuite, cfgCase configCase) []string {
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

func ensureFileName(err error, filename string) error {
	if strings.Contains(err.Error(), filename) {
		return err // already contains filename, nothing else to do
	}
	return fmt.Errorf("%s: %w", filename, err)
}
