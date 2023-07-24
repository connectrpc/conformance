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

#include <string>

#include "absl/flags/flag.h"
#include "absl/flags/parse.h"
#include "absl/flags/usage.h"
#include "connect/conformance/test.grpc.pb.h"
#include "connect/conformance/test.pb.h"
#include "grpcpp/grpcpp.h"

ABSL_FLAG(std::string, host, "127.0.0.1", "host to connect to");
ABSL_FLAG(std::string, port, "", "port to connect to");
ABSL_FLAG(bool, insecure, false, "use insecure credentials");
ABSL_FLAG(std::string, certFile, "", "the server's certificate file");
ABSL_FLAG(std::string, keyFile, "", "the server's key file");

int main(int argc, char** argv) {
  absl::SetProgramUsageMessage("A gRPC-base cross test conformance client");
  absl::ParseCommandLine(argc, argv);

  if (absl::GetFlag(FLAGS_port).empty()) {
    std::cerr << "port must be specified" << std::endl;
    return 1;
  }

  if (absl::GetFlag(FLAGS_insecure)) {
    if (!absl::GetFlag(FLAGS_certFile).empty() || !absl::GetFlag(FLAGS_keyFile).empty()) {
      std::cerr << "insecure cannot be used with certFile or keyFile" << std::endl;
      return 1;
    }
  } else {
    if (absl::GetFlag(FLAGS_certFile).empty() || absl::GetFlag(FLAGS_keyFile).empty()) {
      std::cerr << "insecure or certFile and keyFile must be specified" << std::endl;
      return 1;
    }
  }

  return 0;
}
