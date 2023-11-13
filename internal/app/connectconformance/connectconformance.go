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
	"sort"
	"strings"

	"connectrpc.com/conformance/internal"
	"connectrpc.com/conformance/internal/app/connectconformance/testsuites"
	"connectrpc.com/conformance/internal/app/referenceclient"
	"connectrpc.com/conformance/internal/app/referenceserver"
	conformancev1alpha1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1alpha1"
)

type Flags struct {
	Mode             conformancev1alpha1.TestSuite_TestMode
	ConfigFile       string
	KnownFailingFile string
	Verbose          bool
}

func Run(flags *Flags, command []string, logOut io.Writer) (bool, error) {
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
		_, _ = fmt.Fprintf(logOut, "Computed %d active config case permutations.\n", len(configCases))
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

	// TODO: allow test suite files to indicate on command-line to override use
	//       of built-in, embedded test suite data
	testSuiteData, err := testsuites.LoadTestSuites()
	if err != nil {
		return false, fmt.Errorf("failed to load embedded test suite data: %w", err)
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
	testCaseLib, err := newTestCaseLibrary(allSuites, configCases, flags.Mode)
	if err != nil {
		return false, err
	}
	if flags.Verbose {
		logTestCaseInfo(testCaseLib, logOut)
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

	var clientCreds *conformancev1alpha1.ClientCompatRequest_TLSCreds
	for svrInstance := range testCaseLib.casesByServer {
		if svrInstance.useTLSClientCerts {
			clientCertBytes, clientKeyBytes, err := internal.NewClientCert()
			if err != nil {
				return false, fmt.Errorf("failed to generate client certificate: %w", err)
			}
			clientCreds = &conformancev1alpha1.ClientCompatRequest_TLSCreds{
				Cert: clientCertBytes,
				Key:  clientKeyBytes,
			}
			break
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	isClient := flags.Mode == conformancev1alpha1.TestSuite_TEST_MODE_CLIENT
	var startClient processStarter
	if isClient {
		startClient = runCommand(command)
	} else {
		startClient = runInProcess("reference-client", referenceclient.Run)
	}
	clientProcess, err := runClient(ctx, startClient)
	if err != nil {
		return false, fmt.Errorf("error starting client: %w", err)
	}
	defer clientProcess.stop()

	results := newResults(knownFailing)

	var startServer processStarter
	if isClient {
		startServer = runInProcess("reference-server", referenceserver.Run)
	} else {
		startServer = runCommand(command)
	}
	// TODO: start servers in parallel (up to a limit) to allow parallelism and faster test execution
	for svrInstance, testCases := range testCaseLib.casesByServer {
		runTestCasesForServer(ctx, !isClient, isClient, svrInstance, testCases, clientCreds, startServer, results, clientProcess)
		if !clientProcess.isRunning() {
			err := clientProcess.waitForResponses()
			if err == nil {
				err = errors.New("client process unexpectedly stopped")
			} else {
				err = fmt.Errorf("client process unexpectedly stopped: %w", err)
			}
			return false, err
		}
	}
	clientProcess.closeSend()
	if err := clientProcess.waitForResponses(); err != nil {
		return false, err
	}

	return results.report(logOut)
}

func logTestCaseInfo(testCaseLib *testCaseLibrary, logOut io.Writer) {
	svrInstances := make([]serverInstance, 0, len(testCaseLib.casesByServer))
	for svrInstance := range testCaseLib.casesByServer {
		svrInstances = append(svrInstances, svrInstance)
	}
	sort.Slice(svrInstances, func(i, j int) bool { //nolint:varnamelen
		if svrInstances[i].httpVersion != svrInstances[j].httpVersion {
			return svrInstances[i].httpVersion < svrInstances[j].httpVersion
		}
		if svrInstances[i].protocol != svrInstances[j].protocol {
			return svrInstances[i].protocol < svrInstances[j].protocol
		}
		return !svrInstances[i].useTLS || svrInstances[j].useTLS
	})
	for _, svrInstance := range svrInstances {
		testCases := testCaseLib.casesByServer[svrInstance]
		_, _ = fmt.Fprintf(logOut, "Running %d tests for server config {%s, %s, TLS:%v}...\n",
			len(testCases), svrInstance.httpVersion, svrInstance.protocol, svrInstance.useTLS)
	}
}
