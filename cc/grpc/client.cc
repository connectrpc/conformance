#include "absl/flags/flag.h"
#include "absl/flags/parse.h"
#include "connect/conformance/test.grpc.pb.h"
#include "connect/conformance/test.pb.h"
#include "grpcpp/grpcpp.h"

int main(int argc, char** argv) {
  absl::ParseCommandLine(argc, argv);

  return 0;
}
