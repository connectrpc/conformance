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

package config

const (
	// DefaultHost is the default host to use for the server.
	DefaultHost = "127.0.0.1"
	// DefaultPort is the default port to use for the server. We choose 0 so that
	// an ephemeral port is selected by the OS if no port is specified.
	DefaultPort = "0"
)
