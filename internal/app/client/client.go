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

package client

import (
	"context"
	"io"
)

// Run runs the reference client process. It will read details describing all requests
// to send from stdin and then write the results of each operation to stdout. It
// exits after all data has been read from stdin and all described RPCs have completed.
func Run(ctx context.Context, args []string, stdin io.ReadCloser, stdout, stderr io.WriteCloser) error {
	// TODO: move everything out of cmd/client/main.go into here. Update that main func to call this.
	_ = ctx
	_ = args
	_ = stdin
	_ = stdout
	_ = stderr
	return nil
}
