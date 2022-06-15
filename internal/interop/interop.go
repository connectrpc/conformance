// Copyright 2022 Buf Technologies, Inc.
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

package interop

import testpb "github.com/bufbuild/connect-crosstest/internal/gen/proto/go/grpc/testing"

// NonASCIIErrMsg is a non-ASCII error message.
const NonASCIIErrMsg = "soirée 🎉" // readable non-ASCII

// ErrorDetail is an error detail to be included in an error.
var ErrorDetail = &testpb.ErrorDetail{
	Reason: NonASCIIErrMsg,
	Domain: "connect-crosstest",
}
