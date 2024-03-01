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

package connectconformance

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
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
	"connectrpc.com/conformance/internal/tracer"
	"golang.org/x/sync/semaphore"
)

// Flags are the config values for the test runner that may be provided via
// command-line flags and arguments.
type Flags struct {
	ConfigFile           string
	RunPatterns          []string
	SkipPatterns         []string
	KnownFailingPatterns []string
	KnownFlakyPatterns   []string
	Verbose              bool
	VeryVerbose          bool
	ClientCommand        []string
	ServerCommand        []string
	TestFiles            []string
	MaxServers           uint
	Parallelism          uint
	TLSCertFile          string
	TLSKeyFile           string
	ServerPort           uint
	ServerBind           string
	HTTPTrace            bool
}

func Run(flags *Flags, logPrinter internal.Printer, errPrinter internal.Printer) (bool, error) {
	var configData []byte
	if flags.ConfigFile != "" {
		var err error
		if configData, err = os.ReadFile(flags.ConfigFile); err != nil {
			return false, internal.EnsureFileName(err, flags.ConfigFile)
		}
	} else if flags.Verbose {
		logPrinter.Printf("No config file provided. Using defaults.")
	}
	configCases, err := parseConfig(flags.ConfigFile, configData)
	if err != nil {
		return false, err
	}
	if flags.Verbose {
		logPrinter.Printf("Computed %d config case permutations.", len(configCases))
	}

	knownFailing := parsePatterns(flags.KnownFailingPatterns)
	if knownFailing == nil {
		// treat as empty
		knownFailing = &testTrie{}
	}
	knownFlaky := parsePatterns(flags.KnownFlakyPatterns)
	if knownFlaky == nil {
		// treat as empty
		knownFlaky = &testTrie{}
	}

	runPatterns := parsePatterns(flags.RunPatterns)
	skipPatterns := parsePatterns(flags.SkipPatterns)

	var testSuiteData map[string][]byte
	if len(flags.TestFiles) > 0 {
		testSuiteData, err = testsuites.LoadTestSuitesFromFiles(flags.TestFiles)
		if err != nil {
			return false, fmt.Errorf("failed to load test suite data: %w", err)
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
		logPrinter.Printf("Loaded %d test suite(s), %d test case template(s).", len(allSuites), numCases)
	}

	results, err := run(configCases, knownFailing, knownFlaky, runPatterns, skipPatterns, allSuites, logPrinter, errPrinter, flags)
	if err != nil {
		return false, err
	}
	return results.report(logPrinter), nil
}

func run( //nolint:gocyclo
	configCases []configCase,
	knownFailing *testTrie,
	knownFlaky *testTrie,
	run *testTrie,
	skip *testTrie,
	allSuites map[string]*conformancev1.TestSuite,
	logPrinter internal.Printer,
	errPrinter internal.Printer,
	flags *Flags,
) (*testResults, error) {
	mode := conformancev1.TestSuite_TEST_MODE_UNSPECIFIED
	useReferenceClient := len(flags.ClientCommand) == 0
	useReferenceServer := len(flags.ServerCommand) == 0
	switch {
	case useReferenceServer && !useReferenceClient:
		// Client mode uses a reference server to test a given client
		mode = conformancev1.TestSuite_TEST_MODE_CLIENT
	case useReferenceClient && !useReferenceServer:
		// Server mode uses a reference client to test a given server
		mode = conformancev1.TestSuite_TEST_MODE_SERVER
	default:
		// Otherwise, leave mode as "unspecified" so we'll include
		// neither client-specific nor server-specific cases.
	}
	testCaseLib, err := newTestCaseLibrary(allSuites, configCases, mode)
	if err != nil {
		return nil, err
	}
	svrInstances := serverInstancesSlice(testCaseLib, flags.Verbose)

	// Calculate all permutations of test cases that will be run, including gRPC tests
	allPermutations := testCaseLib.allPermutations(useReferenceClient, useReferenceServer)

	// Validate keys in knownFailing, runPatterns, and noRunPatterns, to
	// make sure they match actual test names (to prevent accidental typos
	// and inadvertently ignored entries)
	if knownFailing.length() > 0 {
		matched, err := tryMatchPatterns("known failing", knownFailing, allPermutations)
		if err != nil {
			return nil, err
		}
		if flags.Verbose {
			logPrinter.Printf("Loaded %d known failing test case pattern(s) that match %d test case permutation(s).",
				knownFailing.length(), matched)
		}
	}
	if knownFlaky.length() > 0 {
		matched, err := tryMatchPatterns("known flaky", knownFlaky, allPermutations)
		if err != nil {
			return nil, err
		}
		if flags.Verbose {
			logPrinter.Printf("Loaded %d known flaky test case pattern(s) that match %d test case permutation(s).",
				knownFlaky.length(), matched)
		}
	}
	if run != nil {
		if _, err := tryMatchPatterns("run patterns", run, allPermutations); err != nil {
			return nil, err
		}
	}
	if skip != nil {
		if _, err := tryMatchPatterns("no-run patterns", skip, allPermutations); err != nil {
			return nil, err
		}
	}
	// we don't allow ambiguity whether a file is known to fail vs known to be flaky
	if knownFailing.length() > 0 && knownFlaky.length() > 0 {
		var conflicts []string
		for _, testCase := range allPermutations {
			name := testCase.Request.TestName
			if knownFailing.matchPattern(name) && knownFlaky.matchPattern(name) {
				conflicts = append(conflicts, name)
			}
		}
		if len(conflicts) > 0 {
			sort.Strings(conflicts)
			return nil, fmt.Errorf("known failing and known flaky configs are ambiguous as some test cases are matched as both\n:%v", strings.Join(conflicts, "\n"))
		}
	}

	filter := newFilter(run, skip)
	if flags.Verbose {
		logPrinter.Printf("Computed %d test case permutation(s) across %d server configuration(s).",
			len(allPermutations), len(testCaseLib.casesByServer))
		if filter != nil {
			var count int
			filteredServerInstances := map[serverInstance]struct{}{}
			for _, tc := range allPermutations {
				if filter.accept(tc) {
					count++
					filteredServerInstances[serverInstance{
						protocol:          tc.Request.Protocol,
						httpVersion:       tc.Request.HttpVersion,
						useTLS:            len(tc.Request.ServerTlsCert) > 0,
						useTLSClientCerts: tc.Request.ClientTlsCreds != nil,
					}] = struct{}{}
				}
			}
			if count != len(allPermutations) {
				logPrinter.Printf("Filtered tests to %d test case permutation(s) across %d server configuration(s).",
					count, len(filteredServerInstances))
			}
		}
	}

	var clientCreds *conformancev1.ClientCompatRequest_TLSCreds
	for svrInstance := range testCaseLib.casesByServer {
		if svrInstance.useTLSClientCerts {
			clientCertBytes, clientKeyBytes, err := internal.NewClientCert()
			if err != nil {
				return nil, fmt.Errorf("failed to generate client certificate: %w", err)
			}
			clientCreds = &conformancev1.ClientCompatRequest_TLSCreds{
				Cert: clientCertBytes,
				Key:  clientKeyBytes,
			}
			break
		}
	}

	var trace *tracer.Tracer
	if flags.HTTPTrace {
		trace = &tracer.Tracer{}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var clients []processInfo
	if useReferenceClient {
		clients = []processInfo{
			{
				name: "reference client",
				start: runInProcess([]string{
					"reference-client",
					"-p", strconv.Itoa(int(flags.Parallelism)),
				}, func(ctx context.Context, args []string, inReader io.ReadCloser, outWriter, errWriter io.WriteCloser) error {
					return referenceclient.RunInReferenceMode(ctx, args, inReader, outWriter, errWriter, trace)
				}),
				isReferenceImpl: true,
			},
			{
				name: "reference client (grpc)",
				start: runInProcess([]string{
					"grpc-reference-client",
					"-p", strconv.Itoa(int(flags.Parallelism)),
				}, grpcclient.Run),
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

	results := newResults(knownFailing, knownFlaky, trace)

	for _, clientInfo := range clients {
		clientProcess, err := runClient(ctx, clientInfo.start)
		if err != nil {
			return nil, fmt.Errorf("error starting client: %w", err)
		}
		defer clientProcess.stop()

		var servers []processInfo
		if useReferenceServer {
			servers = []processInfo{
				{
					name: "reference server",
					start: runInProcess([]string{
						"reference-server",
						"-port", strconv.FormatUint(uint64(flags.ServerPort), 10),
						"-bind", flags.ServerBind,
						"-cert", flags.TLSCertFile,
						"-key", flags.TLSKeyFile,
					}, func(ctx context.Context, args []string, inReader io.ReadCloser, outWriter, errWriter io.WriteCloser) error {
						return referenceserver.RunInReferenceMode(ctx, args, inReader, outWriter, errWriter, trace)
					}),
					isReferenceImpl: true,
				},
				{
					name: "reference server (grpc)",
					start: runInProcess([]string{
						"grpc-reference-server",
						"-port", strconv.FormatUint(uint64(flags.ServerPort), 10),
						"-bind", flags.ServerBind,
					}, grpcserver.Run),
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
					testCases = testCaseLib.filterGRPCImplTestCases(testCases, clientInfo.isGrpcImpl, serverInfo.isGrpcImpl)
					testCases = filter.apply(testCases)
					if len(testCases) == 0 {
						continue
					}

					if err := sema.Acquire(ctx, 1); err != nil {
						return err
					}

					if flags.Verbose {
						var with string
						switch {
						case clientInfo.name != "" && serverInfo.name != "":
							with = clientInfo.name + " and " + serverInfo.name
						case clientInfo.name != "":
							with = clientInfo.name
						case serverInfo.name != "":
							with = serverInfo.name
						}
						logTestCaseInfo(with, svrInstance, len(testCases), logPrinter)
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
						runTestCasesForServer(
							ctx,
							clientInfo.isReferenceImpl,
							serverInfo.isReferenceImpl,
							svrInstance,
							testCases,
							clientCreds,
							serverInfo.start,
							logPrinter,
							errPrinter,
							results,
							clientProcess,
							trace,
							flags.VeryVerbose,
						)
					}(ctx, clientInfo, serverInfo, svrInstance)
				}
			}
			return nil
		}()

		if err != nil {
			return nil, err
		}

		clientProcess.closeSend()
		if err := clientProcess.waitForResponses(); err != nil {
			return nil, err
		}
	}

	return results, nil
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

func logTestCaseInfo(with string, svrInstance serverInstance, numCases int, logPrinter internal.Printer) {
	var tlsMode string
	switch {
	case !svrInstance.useTLS:
		tlsMode = "false"
	case svrInstance.useTLS && svrInstance.useTLSClientCerts:
		tlsMode = "true (with client certs)"
	default:
		tlsMode = "true"
	}
	logPrinter.Printf("Running %d tests with %s for server config {%s, %s, TLS:%s}...",
		numCases, with, svrInstance.httpVersion, svrInstance.protocol, tlsMode)
}

func tryMatchPatterns(what string, patterns *testTrie, testCases []*conformancev1.TestCase) (int, error) {
	var matchCount int
	for _, tc := range testCases {
		if patterns.matchPattern(tc.Request.TestName) {
			matchCount++
		}
	}
	unmatched := patterns.allUnmatched()
	if len(unmatched) == 0 {
		return matchCount, nil
	}
	unmatchedSlice := make([]string, 0, len(unmatched))
	for name := range unmatched {
		unmatchedSlice = append(unmatchedSlice, name)
	}
	sort.Strings(unmatchedSlice)
	return matchCount, fmt.Errorf("%s: unmatched and possibly invalid patterns:\n%v", what, strings.Join(unmatchedSlice, "\n"))
}
