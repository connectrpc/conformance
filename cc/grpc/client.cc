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
