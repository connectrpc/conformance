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
	"embed"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"
)

//go:embed *.ts
//go:embed gen/proto/connectrpc/conformance/v1/*.ts
var grpcwebFS embed.FS

func main() {
	fs, err := loadFiles()
	if err != nil {
		return
	}
	fmt.Println(fs)

	_ = os.WriteFile("app.ts", fs["app.ts"], 0755)
	// out, _ := exec.Command("./foobar").Output()
	// fmt.Printf("Output: %s\n", out)
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, "./app.ts") //nolint:gosec
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Cancel = func() error {
		err := cmd.Process.Signal(syscall.SIGTERM)
		if err != nil {
			// Signals like above are not supported on Windows. So if signal fails, try killing.
			err = cmd.Process.Kill()
		}
		return err
	}
	if err := cmd.Start(); err != nil {
		cancel()
		return
	}
}

func loadFiles() (map[string][]byte, error) {
	grpcweb := map[string][]byte{}
	err := fs.WalkDir(grpcwebFS, ".", func(currentPath string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || strings.ToLower(path.Ext(entry.Name())) != ".ts" {
			return nil
		}
		data, err := grpcwebFS.ReadFile(currentPath)
		if err != nil {
			return fmt.Errorf("failed to load test suite data file %s: %w", currentPath, err)
		}
		grpcweb[currentPath] = data
		return nil
	})
	if err != nil {
		return nil, err
	}
	return grpcweb, nil
}
