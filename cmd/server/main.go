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
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"connectrpc.com/conformance/internal/app/server"
)

func main() {
	modeArg := flag.String("mode", "", "mode in which the test suite is run. must be one of 'server' or 'client'.")
	runner := flag.String("runner", "", "the runner under test")

	flag.Parse()

	// var err error
	mode := *modeArg

	fmt.Println(mode)
	if mode == "client" {
		go server.Run(context.Background(), os.Args, os.Stdin, os.Stdout, os.Stderr)

		// w := bufio.NewWriter(os.Stdin)
		// io.WriteString(w, "fartsie")
		fmt.Println("writing to stdin")
		if _, err := io.WriteString(os.Stdin, "Hello World"); err != nil {
			fmt.Println("err is")
			fmt.Println(err)
		}
		fmt.Println("donezo")

		// writer := bufio.NewReadWriter(r *Reader, w *Writer)()

		time.Sleep(time.Duration(10000 * time.Millisecond))

		// 	os.S
	} else if mode == "server" {
		// Look up the program and start it somehow
		fmt.Println("starting ", runner)
	} else {
		fmt.Println("a mode of 'client' or 'server' is required")
	}
}
