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
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"connectrpc.com/conformance/internal/app/client"
	"connectrpc.com/conformance/internal/app/connectconformance/testsuites"
	"connectrpc.com/conformance/internal/app/server"
	conformancev1alpha1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1alpha1"
)

type Flags struct {
	Mode             conformancev1alpha1.TestSuite_TestMode
	ConfigFile       string
	KnownFailingFile string
}

func Run(flags *Flags, command []string, logOut io.Writer) error {
	var configData []byte
	if flags.ConfigFile != "" {
		var err error
		if configData, err = os.ReadFile(flags.ConfigFile); err != nil {
			return ensureFileName(err, flags.ConfigFile)
		}
	}
	configCases, err := parseConfig(flags.ConfigFile, configData)
	if err != nil {
		return err
	}

	var knownFailingData []byte
	if flags.KnownFailingFile != "" {
		var err error
		if knownFailingData, err = os.ReadFile(flags.KnownFailingFile); err != nil {
			return ensureFileName(err, flags.KnownFailingFile)
		}
	}
	knownFailing := parseKnownFailing(knownFailingData)
	if err != nil {
		return err
	}

	// TODO: allow test suite files to indicate on command-line to override use
	//       of built-in, embedded test suite data
	testSuiteData, err := testsuites.LoadTestSuites()
	if err != nil {
		return fmt.Errorf("failed to load embedded test suite data: %w", err)
	}
	allSuites, err := parseTestSuites(testSuiteData)
	if err != nil {
		return fmt.Errorf("embedded test suite: %w", err)
	}
	testCaseLib, err := newTestCaseLibrary(allSuites, configCases, flags.Mode)
	if err != nil {
		return err
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
		return fmt.Errorf("file %s contains unmatched and possibly invalid patterns:\n%v",
			flags.KnownFailingFile, strings.Join(unmatchedSlice, "\n"))
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	isClient := flags.Mode == conformancev1alpha1.TestSuite_TEST_MODE_CLIENT
	var startClient processStarter
	if isClient {
		startClient = runCommand(command)
	} else {
		startClient = runInProcess("reference-client", client.Run)
	}
	clientProcess, err := runClient(ctx, startClient)
	if err != nil {
		return fmt.Errorf("error starting client: %w", err)
	}
	defer clientProcess.stop()

	results := newResults(knownFailing)

	var startServer processStarter
	if isClient {
		startServer = runInProcess("reference-server", server.Run)
	} else {
		startServer = runCommand(command)
	}
	// TODO: start servers in parallel (up to a limit) to allow parallelism and faster test execution
	for svrInstance, testCases := range testCaseLib.casesByServer {
		runTestCasesForServer(ctx, isClient, svrInstance, testCases, startServer, results, clientProcess)
		if !clientProcess.isRunning() {
			return clientProcess.waitForResponses()
		}
	}
	clientProcess.closeSend()
	if err := clientProcess.waitForResponses(); err != nil {
		return err
	}

	return results.report(logOut)
}
