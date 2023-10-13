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

	"connectrpc.com/conformance/internal/app/server"
)

const (
	h1PortFlagName   = "h1port"
	h2PortFlagName   = "h2port"
	insecureFlagName = "insecure"
)

func main() {
	h1Port := flag.String("h1Port", "8080", "port for HTTP/1.1 traffic")
	h2Port := flag.String("h2Port", "8081", "port for HTTP/2 traffic")

	flag.Parse()

	args := []string{*h1Port, *h2Port}

	err := server.Run(context.Background(), args, nil, nil, nil)
	if err != nil {
		fmt.Println("an error occurred running the server ", err)
	}
}
