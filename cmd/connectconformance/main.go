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
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"connectrpc.com/conformance/internal"
	"connectrpc.com/conformance/internal/app/connectconformance"
	"github.com/spf13/cobra"
)

const (
	modeFlagName         = "mode"
	configFlagName       = "conf"
	knownFailingFlagName = "known-failing"
	verboseFlagName      = "verbose"
	verboseFlagShortName = "v"
	versionFlagName      = "version"
)

type flags struct {
	mode             string
	configFile       string
	knownFailingFile string
	verbose          bool
	version          bool
}

func main() {
	flagset := &flags{}
	rootCmd := &cobra.Command{
		Use: `connectconformance --mode [client|server] -- command...
  connectconformance --mode both -- client-command... ---- server-command...`,
		Short: "Runs conformance tests against the given command.",
		Long: `Runs conformance tests against a Connect implementation. Depending on the mode,
the given command must be either a conformance client or a conformance server.
When mode is both, two commands are given, separated by a quadruple-slash ("----"),
with the client command being first and the server command second.

A conformance client tests a client implementation: the command reads test cases
from stdin. Each test case describes an RPC to make. The command then records
the result of each operation to stdout. The input is a sequence of binary-encoded
Protobuf messages of type connectrpc.conformance.v1.ClientCompatRequest,
each prefixed with a fixed-32-bit length. The output is expected to be similar:
a sequence of fixed-32-bit-length-prefixed messages, but the results are of
type connectrpc.conformance.v1.ClientCompatResponse. The command should exit
when it has read all test cases (i.e reached EOF of stdin) and then issued RPCs
and recorded all results to stdout. The command should also exit and abort any
in-progress RPCs if it receives a SIGTERM signal.

A conformance server tests a server implementation: the command reads the required
server properties from stdin. This comes in the form of a binary-encoded Protobuf
message of type connectrpc.conformance.v1.ServerCompatRequest, prefixed with a
fixed-32-bit length. The command should then start a server process and write its
properties to stdout in the same form as the input, but using a
connectrpc.conformance.v1.ServerCompatResponse message. The server process should
provide an implementation of the test service defined by
connectrpc.conformance.v1.ConformanceService. The command should exit
upon receiving a SIGTERM signal. The command maybe invoked repeatedly, to start
and test servers with different properties.

A configuration file may be provided which specifies what features the client
or server under test supports. This is used to filter the set of test cases
that will be executed. If no config file is indicated, default configuration
will be used.

A file with a list of known-failing test cases may also be provided. For these
cases, the test runner reverses its assertions: it considers the test case
failing to be successful; if the test case succeeds, it is considered a failure.
(The latter aspect makes sure that the file is up-to-date and does not include
test cases which have actually been fixed.)
`,
		Run: func(cmd *cobra.Command, args []string) {
			run(flagset, args)
		},
	}
	bind(rootCmd, flagset)
	_ = rootCmd.Execute()
}

func bind(cmd *cobra.Command, flags *flags) {
	cmd.Flags().StringVar(&flags.mode, modeFlagName, "", "required: the mode of the test to run; must be 'client', 'server', or 'both'")
	cmd.Flags().StringVar(&flags.configFile, configFlagName, "", "a config file in YAML format with supported features")
	cmd.Flags().StringVar(&flags.knownFailingFile, knownFailingFlagName, "", "a file with a list of known-failing test cases")
	cmd.Flags().BoolVarP(&flags.verbose, verboseFlagName, verboseFlagShortName, false, "enables verbose output")
	cmd.Flags().BoolVar(&flags.version, versionFlagName, false, "print version and exit")
}

func run(flags *flags, command []string) {
	if flags.version {
		fmt.Printf("%s %s\n", filepath.Base(os.Args[0]), internal.Version)
		return
	}

	fatal := func(format string, args ...any) {
		_, _ = fmt.Fprintf(os.Stderr, format+"\n", args...)
		os.Exit(1)
	}

	if len(command) == 0 {
		fatal(`Positional arguments are required to configure the command line of the client or server under test.`)
	}

	var clientCommand, serverCommand []string
	switch flags.mode {
	case "client":
		clientCommand = command
	case "server":
		serverCommand = command
	case "both":
		pos := positionOf(command, "----")
		if pos < 0 {
			fatal(`Command is missing "----" separator. In mode "both", positional args should include client command, "----", then server command.`)
		}
		clientCommand = command[:pos]
		serverCommand = command[pos+1:]
		if len(clientCommand) == 0 {
			fatal(`Client command (before the "----") is empty.`)
		}
		if len(serverCommand) == 0 {
			fatal(`Server command (after the "----") is empty.`)
		}
	default:
		// TODO: support mode "both", which would allow the caller to supply both the
		//       client and server commands, instead of using a reference impl?
		fatal(`Invalid mode: expecting "client", "server", or "both"; got %q`, flags.mode)
	}

	for _, cmd := range [][]string{clientCommand, serverCommand} {
		if len(cmd) == 0 {
			continue
		}
		// Resolve command name, using PATH if need be.
		resolvedCommand, err := exec.LookPath(cmd[0])
		if err != nil {
			fatal("%s", errWithFilename(err, cmd[0]))
		}
		cmd[0] = resolvedCommand
	}

	ok, err := connectconformance.Run(
		&connectconformance.Flags{
			ConfigFile:       flags.configFile,
			KnownFailingFile: flags.knownFailingFile,
			Verbose:          flags.verbose,
			ClientCommand:    clientCommand,
			ServerCommand:    serverCommand,
		},
		os.Stdout,
	)
	if err != nil {
		fatal("%s", err)
	}
	if !ok {
		os.Exit(1)
	}
}

func positionOf(slice []string, item string) int {
	for i, str := range slice {
		if str == item {
			return i
		}
	}
	return -1
}

func errWithFilename(err error, filename string) error {
	if strings.Contains(err.Error(), filename) {
		return err
	}
	// make sure error message includes file name
	return fmt.Errorf("%s: %w", filename, err)
}
