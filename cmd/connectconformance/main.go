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

package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"connectrpc.com/conformance/internal"
	"connectrpc.com/conformance/internal/app/connectconformance"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	modeFlagName          = "mode"
	configFlagName        = "conf"
	testFileFlagName      = "test-file"
	knownFailingFlagName  = "known-failing"
	knownFlakyFlagName    = "known-flaky"
	runFlagName           = "run"
	skipFlagName          = "skip"
	verboseFlagName       = "verbose"
	verboseFlagShortName  = "v"
	veryVerboseFlagName   = "vv"
	versionFlagName       = "version"
	maxServersFlagName    = "max-servers"
	parallelFlagName      = "parallel"
	parallelFlagShortName = "p"
	tlsCertFlagName       = "cert"
	tlsKeyFlagName        = "key"
	portFlagName          = "port"
	bindFlagName          = "bind"
	traceFlagName         = "trace"
)

type flags struct {
	mode                 string
	configFile           string
	testFiles            []string
	runPatterns          []string
	skipPatterns         []string
	knownFailingPatterns []string
	knownFlakyPatterns   []string
	verbose              bool
	veryVerbose          bool
	version              bool
	maxServers           uint
	parallel             uint
	tlsCertFile          string
	tlsKeyFile           string
	port                 uint
	bind                 string
	trace                bool
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

Flags can also be specified to filter the list of test case permutations run
and change how results are interpreted. These are the --run, --skip,
--known-failing, and --known-flaky flags. The --run and --skip flags should
be used when running and troubleshooting specific test cases. For continuous
integration tests, the --known-failing and --known-flaky flags should be used
instead. With these, the tests are still run, but failing tests are interpreted
differently. With --known-failing, the test cases must fail. This is useful to
make sure that the list of known-failing test cases is updated if/when test
failures are fixed. All of these flags support reading the list of test case
patterns from a file using the "@" prefix. So a flag value with this prefix
should be the path to a text file, which contains names or patterns, one per
line.
`,
		Run: func(cmd *cobra.Command, args []string) {
			run(flagset, cmd.Flags(), args)
		},
	}
	bind(rootCmd, flagset)
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(2)
	}
}

func bind(cmd *cobra.Command, flags *flags) {
	cmd.Flags().StringVar(&flags.mode, modeFlagName, "",
		"required: the mode of the test to run; must be 'client', 'server', or 'both'")
	cmd.Flags().StringVar(&flags.configFile, configFlagName, "",
		"a config file in YAML format with supported features")
	cmd.Flags().StringArrayVar(&flags.testFiles, testFileFlagName, nil,
		"a file in YAML format containing tests to run, which will skip running the embedded tests; can be specified more than once")
	cmd.Flags().StringArrayVar(&flags.runPatterns, runFlagName, nil,
		"a pattern indicating the name of test cases to run; when absent, all tests are run (other than indicated by --skip); can be specified more than once")
	cmd.Flags().StringArrayVar(&flags.skipPatterns, skipFlagName, nil,
		"a pattern indicating the name of test cases to skip; when absent, no tests are skipped; can be specified more than once")
	cmd.Flags().StringArrayVar(&flags.knownFailingPatterns, knownFailingFlagName, nil,
		"a pattern indicating the name of test cases that are known to fail; these test cases will be required to fail for the run to be successful; can be specified more than once")
	cmd.Flags().StringArrayVar(&flags.knownFlakyPatterns, knownFlakyFlagName, nil,
		"a pattern indicating the name of test cases that are flaky; these test cases are allowed (but not required) to fail; can be specified more than once")
	cmd.Flags().BoolVarP(&flags.verbose, verboseFlagName, verboseFlagShortName, false,
		"enables verbose output")
	cmd.Flags().BoolVar(&flags.veryVerbose, veryVerboseFlagName, false,
		"enables even more verbose output")
	cmd.Flags().BoolVar(&flags.version, versionFlagName, false,
		"print version and exit")
	cmd.Flags().UintVar(&flags.maxServers, maxServersFlagName, 4,
		"the maximum number of server processes to be running in parallel")
	cmd.Flags().UintVarP(&flags.parallel, parallelFlagName, parallelFlagShortName, uint(runtime.GOMAXPROCS(0)*4),
		"in server mode, the level of parallelism used when issuing RPCs")
	cmd.Flags().StringVar(&flags.tlsCertFile, tlsCertFlagName, "",
		"in client mode, the path to a PEM-encoded TLS certificate file that the reference server should use")
	cmd.Flags().StringVar(&flags.tlsKeyFile, tlsKeyFlagName, "",
		"in client mode, the path to a PEM-encoded TLS key file that the reference server should use")
	cmd.Flags().UintVar(&flags.port, portFlagName, internal.DefaultPort,
		"in client mode, the port number on which the reference server should listen (implies --max-servers=1)")
	cmd.Flags().StringVar(&flags.bind, bindFlagName, internal.DefaultHost,
		"in client mode, the bind address on which the reference server should listen (0.0.0.0 means listen on all interfaces)")
	cmd.Flags().BoolVar(&flags.trace, traceFlagName, false,
		"if true, full HTTP traces will be captured and shown alongside failing test cases")
}

func run(flags *flags, cobraFlags *pflag.FlagSet, command []string) { //nolint:gocyclo
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

	if flags.maxServers == 0 {
		fatal(`Invalid max servers: must be greater than zero`)
	}
	if flags.port != 0 {
		if flags.maxServers > 1 && cobraFlags.Changed(maxServersFlagName) {
			fatal(`Invalid max servers: cannot be greater than one when non-zero --port is specified`)
		}
		// Can only run a single server process at a time with a specific port.
		flags.maxServers = 1
	}
	if flags.parallel == 0 {
		fatal(`Invalid parallelism: must be greater than zero`)
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
		fatal(`Invalid mode: expecting "client", "server", or "both"; got %q`, flags.mode)
	}

	if flags.mode != "client" {
		if cobraFlags.Changed(tlsCertFlagName) {
			fatal(fmt.Sprintf("Cannot specify --%s flag when mode is %s", tlsCertFlagName, flags.mode))
		}
		if cobraFlags.Changed(tlsKeyFlagName) {
			fatal(fmt.Sprintf("Cannot specify --%s flag when mode is %s", tlsKeyFlagName, flags.mode))
		}
		if cobraFlags.Changed(portFlagName) {
			fatal(fmt.Sprintf("Cannot specify --%s flag when mode is %s", portFlagName, flags.mode))
		}
		if cobraFlags.Changed(bindFlagName) {
			fatal(fmt.Sprintf("Cannot specify --%s flag when mode is %s", bindFlagName, flags.mode))
		}
	}
	if flags.mode != "server" {
		if cobraFlags.Changed(parallelFlagName) {
			fatal(fmt.Sprintf("Cannot specify --%s/-%s flag when mode is %s", parallelFlagName, parallelFlagShortName, flags.mode))
		}
	}

	switch {
	case flags.tlsCertFile != "" && flags.tlsKeyFile == "":
		fatal(fmt.Sprintf("Missing TLS key: --%s flag must be specified when --%s is used", tlsKeyFlagName, tlsCertFlagName))
	case flags.tlsCertFile == "" && flags.tlsKeyFile != "":
		fatal(fmt.Sprintf("Missing TLS certificate: --%s flag must be specified when --%s is used", tlsCertFlagName, tlsKeyFlagName))
	case flags.tlsCertFile != "":
		// Make sure the given TLS files are present and readable
		file, err := os.Open(flags.tlsCertFile)
		if err != nil {
			fatal(errWithFilename(err, flags.tlsCertFile).Error())
		}
		_ = file.Close()
		file, err = os.Open(flags.tlsKeyFile)
		if err != nil {
			fatal(errWithFilename(err, flags.tlsKeyFile).Error())
		}
		_ = file.Close()
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

	runPatterns, err := argsToPatterns(flags.runPatterns)
	if err != nil {
		fatal("%s", err)
	}
	skipPatterns, err := argsToPatterns(flags.skipPatterns)
	if err != nil {
		fatal("%s", err)
	}
	knownFailingPatterns, err := argsToPatterns(flags.knownFailingPatterns)
	if err != nil {
		fatal("%s", err)
	}
	knownFlakyPatterns, err := argsToPatterns(flags.knownFlakyPatterns)
	if err != nil {
		fatal("%s", err)
	}

	ok, err := connectconformance.Run(
		&connectconformance.Flags{
			ConfigFile:           flags.configFile,
			RunPatterns:          runPatterns,
			SkipPatterns:         skipPatterns,
			KnownFailingPatterns: knownFailingPatterns,
			KnownFlakyPatterns:   knownFlakyPatterns,
			TestFiles:            flags.testFiles,
			Verbose:              flags.verbose || flags.veryVerbose,
			VeryVerbose:          flags.veryVerbose,
			ClientCommand:        clientCommand,
			ServerCommand:        serverCommand,
			MaxServers:           flags.maxServers,
			Parallelism:          flags.parallel,
			TLSCertFile:          flags.tlsCertFile,
			TLSKeyFile:           flags.tlsKeyFile,
			ServerPort:           flags.port,
			ServerBind:           flags.bind,
			HTTPTrace:            flags.trace,
		},
		internal.NewPrinter(os.Stdout),
		internal.NewPrinter(os.Stderr),
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

func argsToPatterns(args []string) ([]string, error) {
	patterns := make([]string, 0, len(args))
	for _, pattern := range args {
		filename := strings.TrimPrefix(pattern, "@")
		if filename == pattern {
			// no prefix stripped? not a path reference
			patterns = append(patterns, pattern)
			continue
		}
		var data []byte
		if filename != "" {
			var err error
			if data, err = os.ReadFile(filename); err != nil {
				return nil, internal.EnsureFileName(err, filename)
			}
		}
		return parsePatternFile(data), nil
	}
	return patterns, nil
}

func parsePatternFile(data []byte) []string {
	lines := bytes.Split(data, []byte{'\n'})
	patterns := make([]string, 0, len(lines))
	for _, line := range lines {
		line := bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		if line[0] == '#' {
			// comment line
			continue
		}
		patterns = append(patterns, string(line))
	}
	return patterns
}
