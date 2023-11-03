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

package main

import (
	"log"
	"os"

	"connectrpc.com/conformance/internal/app/connectconformance"
	conformancev1alpha1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1alpha1"
	"github.com/spf13/cobra"
)

const (
	modeFlagName         = "mode"
	configFlagName       = "conf"
	knownFailingFlagName = "known-failing"
)

type flags struct {
	mode             string
	configFile       string
	knownFailingFile string
}

func main() {
	flagset := &flags{}
	rootCmd := &cobra.Command{
		Use:   "connectconformance --mode [client|server] command...",
		Short: "Runs conformance tests against the given command.",
		Long: `Runs conformance tests against a Connect implementation. Depending on the mode,
the given command must be either a conformance client or a conformance server.

A conformance client tests a client implementation: the command reads test cases
from stdin. Each test case describes an RPC to make. The command then records
the result of each operation to stdout. The input is a sequence of binary-encoded
Protobuf messages of type connectrpc.conformance.v1alpha1.ClientCompatRequest,
each prefixed with a varint-encoded length. The output is expected to be similar:
a sequence of varint-encoded-length-prefixed messages, but the results are of
type connectrpc.conformance.v1alpha1.ClientCompatResponse. The command should exit
when it has read all test cases (i.e reached EOF of stdin) and then issued RPCs
and recorded all results to stdout. The command should also exit and abort any
in-progress RPCs if it receives a SIGTERM signal.

A conformance server tests a server implementation: the command reads the required
server properties from stdin. This comes in the form of a binary-encoded Protobuf
message of type connectrpc.conformance.v1alpha1.ServerCompatRequest. The command
should then start a server process and write its properties to stdout in the form
of a binary-encoded connectrpc.conformance.v1alpha1.ServerCompatResponse message.
The server process should provide an implementation of the test service defined
by connectrpc.conformance.v1alpha1.ConformanceService. The command should exit
upon receiving a SIGTERM signal. The command maybe invoked repeatedly, to start
and test servers with different properties.

A configuration file may be provided which specifies what features the client
or server under test supports. This is used to filter the set of test cases
that will be executed. If no config file is indicated, all tests will be run.

A file with a list of known-failing test cases may also be provided. For these
cases, the test runner reverses its assertions: it considers the test case
failing to be successful; if the test case succeeds, it is considered a failure.
(The latter aspect makes sure that the file doesn't is up-to-date and does not
include test cases which have actually been fixed.)
`,
		Run: func(cmd *cobra.Command, args []string) {
			run(flagset, args)
		},
	}
	bind(rootCmd, flagset)
	_ = rootCmd.Execute()
}

func bind(cmd *cobra.Command, flags *flags) {
	cmd.Flags().StringVar(&flags.mode, modeFlagName, "", "required: the mode of the test to run; must be 'client' or 'server'")
	cmd.Flags().StringVar(&flags.configFile, configFlagName, "", "a config file in YAML format with supported features")
	cmd.Flags().StringVar(&flags.knownFailingFile, knownFailingFlagName, "", "a file with a list of known-failing test cases")
}

func run(flags *flags, command []string) {
	var mode conformancev1alpha1.TestSuite_TestMode
	switch flags.mode {
	case "client":
		mode = conformancev1alpha1.TestSuite_TEST_MODE_CLIENT
	case "server":
		mode = conformancev1alpha1.TestSuite_TEST_MODE_SERVER
	default:
		// TODO: support mode "both", which would allow the caller to supply both the
		//       client and server commands, instead of using a reference impl?
		log.Fatalf(`invalid mode: expecting "client" or "server""; got %q`, flags.mode)
	}
	err := connectconformance.Run(
		&connectconformance.Flags{
			Mode:             mode,
			ConfigFile:       flags.configFile,
			KnownFailingFile: flags.knownFailingFile,
		},
		command,
		os.Stdout,
	)
	if err != nil {
		log.Fatal(err)
	}
}
