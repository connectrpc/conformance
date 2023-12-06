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
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"sync"

	"connectrpc.com/conformance/internal"
	"connectrpc.com/conformance/internal/app/connectconformance/testsuites"
	"connectrpc.com/conformance/internal/app/grpcclient"
	"connectrpc.com/conformance/internal/app/grpcserver"
	"connectrpc.com/conformance/internal/app/referenceclient"
	"connectrpc.com/conformance/internal/app/referenceserver"
	conformancev1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1"
	"golang.org/x/sync/semaphore"
	"google.golang.org/protobuf/proto"
)

// Flags are the config values for the test runner that may be provided via
// command-line flags and arguments.
type Flags struct {
	ConfigFile       string
	KnownFailingFile string
	Verbose          bool
	ClientCommand    []string
	ServerCommand    []string
	TestFile         string
	MaxServers       uint
	Parallelism      uint
}

func Run(flags *Flags, logOut io.Writer) (bool, error) { //nolint:gocyclo
	var configData []byte
	if flags.ConfigFile != "" {
		var err error
		if configData, err = os.ReadFile(flags.ConfigFile); err != nil {
			return false, ensureFileName(err, flags.ConfigFile)
		}
	} else if flags.Verbose {
		_, _ = fmt.Fprintf(logOut, "No config file provided. Using defaults.\n")
	}
	configCases, err := parseConfig(flags.ConfigFile, configData)
	if err != nil {
		return false, err
	}
	if flags.Verbose {
		_, _ = fmt.Fprintf(logOut, "Computed %d config case permutations.\n", len(configCases))
	}

	var knownFailingData []byte
	if flags.KnownFailingFile != "" {
		var err error
		if knownFailingData, err = os.ReadFile(flags.KnownFailingFile); err != nil {
			return false, ensureFileName(err, flags.KnownFailingFile)
		}
	}
	knownFailing := parseKnownFailing(knownFailingData)
	if err != nil {
		return false, err
	}
	if flags.Verbose {
		_, _ = fmt.Fprintf(logOut, "Loaded %d known failing test cases/patterns.\n", knownFailing.length())
	}

	var testSuiteData map[string][]byte
	if flags.TestFile != "" {
		testSuiteData, err = testsuites.LoadTestSuitesFromFile(flags.TestFile)
		if err != nil {
			return false, fmt.Errorf("failed to load test suite data from file: %w", err)
		}
	} else {
		testSuiteData, err = testsuites.LoadTestSuites()
		if err != nil {
			return false, fmt.Errorf("failed to load embedded test suite data: %w", err)
		}
	}
	allSuites, err := parseTestSuites(testSuiteData)
	if err != nil {
		return false, fmt.Errorf("embedded test suite: %w", err)
	}

	if flags.Verbose {
		var numCases int
		for _, suite := range allSuites {
			numCases += len(suite.TestCases)
		}
		_, _ = fmt.Fprintf(logOut, "Loaded %d test suites, %d test case templates.\n", len(allSuites), numCases)
	}
	mode := conformancev1.TestSuite_TEST_MODE_UNSPECIFIED
	var useReferenceClient, useReferenceServer bool
	switch {
	case len(flags.ClientCommand) > 0 && len(flags.ServerCommand) == 0:
		// Client mode uses a reference server to test a client
		mode = conformancev1.TestSuite_TEST_MODE_CLIENT
		useReferenceServer = true
	case len(flags.ClientCommand) == 0 && len(flags.ServerCommand) > 0:
		// Server mode uses a reference client to test a server
		mode = conformancev1.TestSuite_TEST_MODE_SERVER
		useReferenceClient = true
	default:
		// Otherwise, no reference server or client is used, so
		// leave mode as "unspecified" (so we'll include neither
		// client-specific nor server-specific cases).
	}
	testCaseLib, err := newTestCaseLibrary(allSuites, configCases, mode)
	if err != nil {
		return false, err
	}
	svrInstances := serverInstancesSlice(testCaseLib, flags.Verbose)
	if flags.Verbose {
		numPermutations := len(testCaseLib.testCases)
		if useReferenceClient || useReferenceServer {
			// Count permutations used against grpc-go reference impl.
			testCaseSlice := make([]*conformancev1.TestCase, 0, len(testCaseLib.testCases))
			for _, testCase := range testCaseLib.testCases {
				testCaseSlice = append(testCaseSlice, testCase)
			}
			grpcTestCases := filterGRPCImplTestCases(testCaseSlice, useReferenceClient)
			numPermutations += len(grpcTestCases)
		}
		_, _ = fmt.Fprintf(logOut, "Computed %d test case permutations across %d server configurations.\n", numPermutations, len(testCaseLib.casesByServer))
	}

	// Validate keys in knownFailing, to make sure they match actual test names
	// (to prevent accidental typos and inadvertently ignored entries)
	for name := range testCaseLib.testCases {
		knownFailing.match(strings.Split(name, "/"))
	}
	unmatched := map[string]struct{}{}
	knownFailing.findUnmatched("", unmatched)
	if len(unmatched) > 0 {
		unmatchedSlice := make([]string, 0, len(unmatched))
		for name := range unmatched {
			unmatchedSlice = append(unmatchedSlice, name)
		}
		sort.Strings(unmatchedSlice)
		return false, fmt.Errorf("file %s contains unmatched and possibly invalid patterns:\n%v",
			flags.KnownFailingFile, strings.Join(unmatchedSlice, "\n"))
	}

	var clientCreds *conformancev1.ClientCompatRequest_TLSCreds
	for svrInstance := range testCaseLib.casesByServer {
		if svrInstance.useTLSClientCerts {
			clientCertBytes, clientKeyBytes, err := internal.NewClientCert()
			if err != nil {
				return false, fmt.Errorf("failed to generate client certificate: %w", err)
			}
			clientCreds = &conformancev1.ClientCompatRequest_TLSCreds{
				Cert: clientCertBytes,
				Key:  clientKeyBytes,
			}
			break
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var clients []processInfo
	if useReferenceClient {
		clients = []processInfo{
			{
				name:            "reference client",
				start:           runInProcess([]string{"reference-client", "-p", strconv.Itoa(int(flags.Parallelism))}, referenceclient.Run),
				isReferenceImpl: true,
			},
			{
				name:       "reference client (grpc)",
				start:      runInProcess([]string{"grpc-reference-client", "-p", strconv.Itoa(int(flags.Parallelism))}, grpcclient.Run),
				isGrpcImpl: true,
			},
		}
	} else {
		clients = []processInfo{
			{
				start: runCommand(flags.ClientCommand),
			},
		}
	}

	results := newResults(knownFailing)

	for _, clientInfo := range clients {
		clientProcess, err := runClient(ctx, clientInfo.start)
		if err != nil {
			return false, fmt.Errorf("error starting client: %w", err)
		}
		defer clientProcess.stop()

		var servers []processInfo
		if useReferenceServer {
			servers = []processInfo{
				{
					name:            "reference server",
					start:           runInProcess([]string{"reference-server"}, referenceserver.RunInReferenceMode),
					isReferenceImpl: true,
				},
				{
					name:       "reference server (grpc)",
					start:      runInProcess([]string{"grpc-reference-server"}, grpcserver.Run),
					isGrpcImpl: true,
				},
			}
		} else {
			servers = []processInfo{
				{
					start: runCommand(flags.ServerCommand),
				},
			}
		}

		err = func() error {
			var wg sync.WaitGroup
			defer wg.Wait()
			sema := semaphore.NewWeighted(int64(flags.MaxServers))

			for _, serverInfo := range servers {
				for _, svrInstance := range svrInstances {
					testCases := testCaseLib.casesByServer[svrInstance]
					if clientInfo.isGrpcImpl || serverInfo.isGrpcImpl {
						testCases = filterGRPCImplTestCases(testCases, clientInfo.isGrpcImpl)
						if len(testCases) == 0 {
							continue
						}
					}

					if err := sema.Acquire(ctx, 1); err != nil {
						return err
					}

					if flags.Verbose {
						with := serverInfo.name
						if with == "" {
							with = clientInfo.name
						}
						logTestCaseInfo(with, svrInstance, len(testCases), logOut)
					}

					// Double-check that client is still running before spawning a server process.
					if !clientProcess.isRunning() {
						err := clientProcess.waitForResponses()
						if err == nil {
							err = errors.New("client process unexpectedly stopped")
						} else {
							err = fmt.Errorf("client process unexpectedly stopped: %w", err)
						}
						return err
					}

					wg.Add(1)
					go func(ctx context.Context, clientInfo processInfo, serverInfo processInfo, svrInstance serverInstance) {
						defer wg.Done()
						defer sema.Release(1)
						runTestCasesForServer(ctx, clientInfo.isReferenceImpl, serverInfo.isReferenceImpl, svrInstance, testCases, clientCreds, serverInfo.start, results, clientProcess)
					}(ctx, clientInfo, serverInfo, svrInstance)
				}
			}
			return nil
		}()

		if err != nil {
			return false, err
		}

		clientProcess.closeSend()
		if err := clientProcess.waitForResponses(); err != nil {
			return false, err
		}
	}

	return results.report(logOut)
}

func serverInstancesSlice(testCaseLib *testCaseLibrary, sorted bool) []serverInstance {
	svrInstances := make([]serverInstance, 0, len(testCaseLib.casesByServer))
	for svrInstance := range testCaseLib.casesByServer {
		svrInstances = append(svrInstances, svrInstance)
	}
	if !sorted {
		return svrInstances
	}
	sort.Slice(svrInstances, func(i, j int) bool { //nolint:varnamelen
		if svrInstances[i].httpVersion != svrInstances[j].httpVersion {
			return svrInstances[i].httpVersion < svrInstances[j].httpVersion
		}
		if svrInstances[i].protocol != svrInstances[j].protocol {
			return svrInstances[i].protocol < svrInstances[j].protocol
		}
		if svrInstances[i].useTLS != svrInstances[j].useTLS {
			return !svrInstances[i].useTLS
		}
		return !svrInstances[i].useTLSClientCerts || svrInstances[j].useTLSClientCerts
	})
	return svrInstances
}

func logTestCaseInfo(with string, svrInstance serverInstance, numCases int, logOut io.Writer) {
	var tlsMode string
	switch {
	case !svrInstance.useTLS:
		tlsMode = "false"
	case svrInstance.useTLS && svrInstance.useTLSClientCerts:
		tlsMode = "true (with client certs)"
	default:
		tlsMode = "true"
	}
	_, _ = fmt.Fprintf(logOut, "Running %d tests with %s for server config {%s, %s, TLS:%s}...\n",
		numCases, with, svrInstance.httpVersion, svrInstance.protocol, tlsMode)
}

type processInfo struct {
	name            string
	start           processStarter
	isReferenceImpl bool
	isGrpcImpl      bool
}

func filterGRPCImplTestCases(testCases []*conformancev1.TestCase, clientIsGRPC bool) []*conformancev1.TestCase {
	// The gRPC reference impl does not support everything that the main reference impl does. So
	// we must filter away any test cases that aren't applicable to the gRPC impls.
	filtered := make([]*conformancev1.TestCase, 0, len(testCases))
	for _, testCase := range testCases {
		// Client only supports gRPC protocol. Server also supports gRPC-Web.
		if clientIsGRPC && testCase.Request.Protocol != conformancev1.Protocol_PROTOCOL_GRPC ||
			testCase.Request.Protocol == conformancev1.Protocol_PROTOCOL_CONNECT {
			continue
		}

		if testCase.Request.Protocol == conformancev1.Protocol_PROTOCOL_GRPC_WEB {
			// grpc-web supports HTTP/1 and HTTP/2
			switch testCase.Request.HttpVersion {
			case conformancev1.HTTPVersion_HTTP_VERSION_1, conformancev1.HTTPVersion_HTTP_VERSION_2:
			default:
				continue
			}
		} else if testCase.Request.HttpVersion != conformancev1.HTTPVersion_HTTP_VERSION_2 {
			// but grpc only supports HTTP/2
			continue
		}

		if testCase.Request.Codec != conformancev1.Codec_CODEC_PROTO {
			continue
		}
		if testCase.Request.Compression != conformancev1.Compression_COMPRESSION_IDENTITY &&
			testCase.Request.Compression != conformancev1.Compression_COMPRESSION_GZIP {
			continue
		}

		if len(testCase.Request.ServerTlsCert) > 0 {
			continue
		}

		filteredCase := proto.Clone(testCase).(*conformancev1.TestCase) //nolint:errcheck,forcetypeassert
		// Insert a path in the test name to indicate that this is against the gRPC impl.
		dir, base := path.Dir(filteredCase.Request.TestName), path.Base(filteredCase.Request.TestName)
		filteredCase.Request.TestName = path.Join(dir, "(grpc impl)", base)
		filtered = append(filtered, filteredCase)
	}
	return filtered
}
