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

#include <memory>
#include <string>

#include "absl/container/flat_hash_map.h"
#include "absl/strings/escaping.h"
#include "absl/strings/str_split.h"
#include "absl/time/clock.h"
#include "connectrpc/conformance/v1/test.grpc.pb.h"
#include "connectrpc/conformance/v1/test.pb.h"
#include "gmock/gmock.h"
#include "google/rpc/status.pb.h"
#include "grpcpp/channel.h"
#include "grpcpp/client_context.h"
#include "grpcpp/create_channel.h"
#include "grpcpp/grpcpp.h"
#include "grpcpp/security/credentials.h"
#include "gtest/gtest.h"

namespace connectrpc::conformance {
namespace {
using v1::ErrorDetail;
using v1::Payload;
using v1::PayloadType;
using v1::SimpleRequest;
using v1::SimpleResponse;
using v1::StreamingInputCallRequest;
using v1::StreamingInputCallResponse;
using v1::StreamingOutputCallRequest;
using v1::StreamingOutputCallResponse;
using v1::TestService;
using v1::UnimplementedService;

constexpr int kEightBytes = 8;

constexpr int kSixteenBytes = 16;
constexpr int kOneKiB = 1024;
constexpr int kTwoKiB = 2028;
constexpr int kThirtyTwoKiB = 32768;
constexpr int kSixtyFourKiB = 65536;
constexpr int kTwoFiftyKiB = 256000;
constexpr int kFiveHundredKiB = 512000;
constexpr int kLargeReqSize = kTwoFiftyKiB;
constexpr int kLargeRespSize = kFiveHundredKiB;
constexpr const auto kLeadingMetadataKey = "x-grpc-test-echo-initial";
constexpr const auto kTrailingMetadataKey = "x-grpc-test-echo-trailing-bin";

constexpr std::array<int, 4> kReqSizes = {kTwoFiftyKiB, kEightBytes, kOneKiB, kThirtyTwoKiB};
constexpr std::array<int, 4> kRespSizes = {kFiveHundredKiB, kSixteenBytes, kTwoKiB, kSixtyFourKiB};
using metadata_t = absl::flat_hash_map<std::string, std::vector<std::string>>;

class GrpcClientTest : public ::testing::Test {
 public:
  static void SetUpTestSuite() {
    // Get the port from the env
    std::string port = getEnvStr("PORT");
    if (port.empty()) {
      port = "8081";
    }
    // Get the host from the env
    std::string host = getEnvStr("HOST", "127.0.0.1");
    // Get the insecure cert from the env
    std::string certFile = getEnvStr("CERT_FILE");
    std::string keyFile = getEnvStr("KEY_FILE");
    if (!certFile.empty() || !keyFile.empty()) {
      channel = grpc::CreateChannel(
          host + ":" + port,
          grpc::SslCredentials(grpc::SslCredentialsOptions{
              .pem_root_certs = certFile,
              .pem_private_key = keyFile,
          }));
    } else {
      channel = grpc::CreateChannel(host + ":" + port, grpc::InsecureChannelCredentials());
    }
    stub = TestService::NewStub(channel);
  }

  static void TearDownTestSuite() {
    stub = nullptr;
    channel = nullptr;
  }

 protected:
  static std::string getEnvStr(const char* name, const char* defaultValue = "") {
    const char* value = std::getenv(name);
    if (value == nullptr) {
      return defaultValue;
    }
    return value;
  }

  static std::shared_ptr<grpc::ChannelInterface> channel;
  static std::unique_ptr<TestService::Stub> stub;

  void newPayload(PayloadType type, int size, Payload* payload) {
    payload->set_type(type);
    payload->set_body(std::string(size, '\0'));
  }

  void addMetadata(
      metadata_t customMetadataString,
      metadata_t customMetadataBinary,
      grpc::ClientContext* context) {
    for (const auto& [key, values] : customMetadataString) {
      for (const auto& value : values) {
        context->AddMetadata(key, value);
      }
    }
    for (const auto& [key, values] : customMetadataBinary) {
      for (const auto& value : values) {
        std::string encoded;
        absl::Base64Escape(value, &encoded);
        context->AddMetadata(key, encoded);
      }
    }
  }

  void customMetadataUnaryTest(metadata_t customMetadataString, metadata_t customMetadataBinary) {
    grpc::ClientContext context;
    addMetadata(customMetadataString, customMetadataBinary, &context);
    SimpleRequest req;
    req.set_response_type(PayloadType::COMPRESSABLE);
    req.set_response_size(1);
    newPayload(PayloadType::COMPRESSABLE, 1, req.mutable_payload());

    SimpleResponse res;
    auto result = stub->UnaryCall(&context, req, &res);
    EXPECT_TRUE(result.ok()) << result.error_code() << ": " << result.error_message();
    EXPECT_EQ(res.payload().type(), PayloadType::COMPRESSABLE);
    EXPECT_EQ(res.payload().body().size(), 1);
    validateMetadata(context, customMetadataString, customMetadataBinary);
  }

  void customMetadataServerStreamingTest(
      metadata_t customMetadataString, metadata_t customMetadataBinary) {
    grpc::ClientContext context;
    addMetadata(customMetadataString, customMetadataBinary, &context);
    StreamingOutputCallRequest req;
    req.set_response_type(PayloadType::COMPRESSABLE);
    req.add_response_parameters()->set_size(1);
    newPayload(PayloadType::COMPRESSABLE, 1, req.mutable_payload());
    auto stream = stub->StreamingOutputCall(&context, req);
    StreamingOutputCallResponse res;
    EXPECT_TRUE(stream->Read(&res));
    EXPECT_EQ(res.payload().type(), PayloadType::COMPRESSABLE);
    EXPECT_EQ(res.payload().body().size(), 1);
    auto result = stream->Finish();
    EXPECT_TRUE(result.ok()) << result.error_code() << ": " << result.error_message();
    validateMetadata(context, customMetadataString, customMetadataBinary);
  }

  void customMetadataFullDuplexTest(
      metadata_t customMetadataString, metadata_t customMetadataBinary) {
    grpc::ClientContext context;
    addMetadata(customMetadataString, customMetadataBinary, &context);
    auto stream = stub->FullDuplexCall(&context);
    StreamingOutputCallRequest req;
    req.set_response_type(PayloadType::COMPRESSABLE);
    req.add_response_parameters()->set_size(1);
    newPayload(PayloadType::COMPRESSABLE, 1, req.mutable_payload());
    EXPECT_TRUE(stream->Write(req));
    StreamingOutputCallResponse res;
    EXPECT_TRUE(stream->Read(&res));
    EXPECT_EQ(res.payload().type(), PayloadType::COMPRESSABLE);
    EXPECT_EQ(res.payload().body().size(), 1);
    stream->WritesDone();
    auto result = stream->Finish();
    EXPECT_TRUE(result.ok()) << result.error_code() << ": " << result.error_message();
    validateMetadata(context, customMetadataString, customMetadataBinary);
  }

  void validateMetadata(
      const grpc::ClientContext& context,
      metadata_t expectedMetadataString,
      metadata_t expectedMetadataBinary) {
    const auto header = context.GetServerInitialMetadata();
    const auto trailer = context.GetServerTrailingMetadata();
    for (const auto& [key, values] : expectedMetadataString) {
      auto range = header.equal_range(key);
      EXPECT_TRUE(range.first != range.second) << "key: " << key;
      if (range.first == range.second) {
        continue;
      }
      std::vector<std::string> actualValues;
      for (auto it = range.first; it != range.second; ++it) {
        actualValues.emplace_back(it->second.data(), it->second.size());
      }
      // The server may have combined multiple lines for a field to a single line, see
      // https://www.rfc-editor.org/rfc/rfc9110.html#section-5.3.
      if (values.size() != actualValues.size() && actualValues.size() == 1) {
        actualValues = absl::StrSplit(actualValues[0], ", ");
      }
      EXPECT_THAT(actualValues, testing::UnorderedElementsAreArray(values)) << "key: " << key;
    }
    for (const auto& [key, values] : expectedMetadataBinary) {
      auto range = trailer.equal_range(key);
      EXPECT_TRUE(range.first != range.second) << "key: " << key;
      if (range.first == range.second) {
        continue;
      }
      std::vector<std::string> actualValues;
      for (auto it = range.first; it != range.second; ++it) {
        actualValues.emplace_back(it->second.data(), it->second.size());
      }
      // The server may have combined multiple lines for a field to a single line, see
      // https://www.rfc-editor.org/rfc/rfc9110.html#section-5.3.
      if (values.size() != actualValues.size() && actualValues.size() == 1) {
        actualValues = absl::StrSplit(actualValues[0], ", ");
      }
      // Decode the base64 encoded values.
      for (auto& value : actualValues) {
        std::string decoded;
        EXPECT_TRUE(absl::Base64Unescape(value, &decoded));
        value = decoded;
      }
      EXPECT_THAT(actualValues, testing::UnorderedElementsAreArray(values)) << "key: " << key;
    }
  }
};

std::shared_ptr<grpc::ChannelInterface> GrpcClientTest::channel = nullptr;
std::unique_ptr<TestService::Stub> GrpcClientTest::stub = nullptr;

TEST_F(GrpcClientTest, EmptyUnary) {
  grpc::ClientContext context;
  google::protobuf::Empty res;
  auto result = stub->EmptyCall(&context, google::protobuf::Empty::default_instance(), &res);
  EXPECT_TRUE(result.ok()) << result.error_code() << ": " << result.error_message();
}

// Performs an idempotent unary RPC with empty request and response messages.
TEST_F(GrpcClientTest, CacheableUnaryCall) {
  grpc::ClientContext context;
  SimpleRequest req;
  req.set_response_type(PayloadType::COMPRESSABLE);
  req.set_response_size(1);
  newPayload(PayloadType::COMPRESSABLE, 1, req.mutable_payload());

  SimpleResponse res;
  auto result = stub->CacheableUnaryCall(&context, req, &res);
  EXPECT_TRUE(result.ok()) << result.error_code() << ": " << result.error_message();
  EXPECT_EQ(res.payload().type(), PayloadType::COMPRESSABLE);
  EXPECT_EQ(res.payload().body().size(), 1);
}

// Performs a unary RPC with large payload in the request and response.
TEST_F(GrpcClientTest, LargeUnaryCall) {
  grpc::ClientContext context;
  SimpleRequest req;
  req.set_response_type(PayloadType::COMPRESSABLE);
  req.set_response_size(kLargeRespSize);
  newPayload(PayloadType::COMPRESSABLE, kLargeReqSize, req.mutable_payload());

  SimpleResponse res;
  auto result = stub->UnaryCall(&context, req, &res);
  EXPECT_TRUE(result.ok()) << result.error_code() << ": " << result.error_message();
  EXPECT_EQ(res.payload().type(), PayloadType::COMPRESSABLE);
  EXPECT_EQ(res.payload().body().size(), kLargeRespSize);
}

// Performs a client streaming RPC.
TEST_F(GrpcClientTest, ClientStreaming) {
  grpc::ClientContext context;
  StreamingInputCallResponse res;
  auto stream = stub->StreamingInputCall(&context, &res);
  int sum = 0;
  for (int size : kReqSizes) {
    StreamingInputCallRequest req;
    newPayload(PayloadType::COMPRESSABLE, size, req.mutable_payload());
    EXPECT_TRUE(stream->Write(req));
    sum += size;
  }
  stream->WritesDone();
  auto result = stream->Finish();
  EXPECT_TRUE(result.ok()) << result.error_code() << ": " << result.error_message();
  EXPECT_EQ(res.aggregated_payload_size(), sum);
}

// Performs a server streaming RPC.
TEST_F(GrpcClientTest, ServerStreaming) {
  grpc::ClientContext context;
  StreamingOutputCallRequest req;
  req.set_response_type(PayloadType::COMPRESSABLE);
  for (int size : kRespSizes) {
    req.add_response_parameters()->set_size(size);
  }
  auto stream = stub->StreamingOutputCall(&context, req);
  StreamingOutputCallResponse res;
  int i = 0;
  while (stream->Read(&res)) {
    EXPECT_EQ(res.payload().type(), PayloadType::COMPRESSABLE);
    EXPECT_EQ(res.payload().body().size(), kRespSizes.at(i++));
  }
}

// Performs a ping-pong style bi-directional streaming RPC.
TEST_F(GrpcClientTest, PingPong) {
  grpc::ClientContext context;
  auto stream = stub->FullDuplexCall(&context);
  for (int reqSize : kReqSizes) {
    StreamingOutputCallRequest req;
    req.set_response_type(PayloadType::COMPRESSABLE);
    req.add_response_parameters()->set_size(reqSize);
    newPayload(PayloadType::COMPRESSABLE, reqSize, req.mutable_payload());
    EXPECT_TRUE(stream->Write(req));
    StreamingOutputCallResponse res;
    EXPECT_TRUE(stream->Read(&res));
    EXPECT_EQ(res.payload().type(), PayloadType::COMPRESSABLE);
    EXPECT_EQ(res.payload().body().size(), reqSize);
  }
  stream->WritesDone();
  auto result = stream->Finish();
  EXPECT_TRUE(result.ok()) << result.error_code() << ": " << result.error_message();
}

// Sets up a bidi streaming with zero messages.
TEST_F(GrpcClientTest, EmptyStream) {
  grpc::ClientContext context;
  auto stream = stub->FullDuplexCall(&context);
  stream->WritesDone();
  auto result = stream->Finish();
  EXPECT_TRUE(result.ok()) << result.error_code() << ": " << result.error_message();
}

// Performs an RPC on a sleep server which causes RPC timeout.
TEST_F(GrpcClientTest, TimeoutOnSleepingServer) {
  grpc::ClientContext context;
  context.set_deadline(std::chrono::system_clock::now() + std::chrono::milliseconds(500));
  auto stream = stub->FullDuplexCall(&context);
  StreamingOutputCallRequest req;
  newPayload(PayloadType::COMPRESSABLE, 27182, req.mutable_payload());
  stream->Write(req);
  absl::SleepFor(absl::Seconds(1));
  auto result = stream->Finish();
  EXPECT_EQ(result.error_code(), grpc::StatusCode::DEADLINE_EXCEEDED);
}

// Cancels the RPC after metadata has been sent but before payloads are sent.
TEST_F(GrpcClientTest, CancelAfterBegin) {
  grpc::ClientContext context;
  StreamingInputCallResponse res;
  auto stream = stub->StreamingInputCall(&context, &res);
  context.TryCancel();
  auto result = stream->Finish();
  EXPECT_EQ(result.error_code(), grpc::StatusCode::CANCELLED);
}

// Cancels the RPC after receiving the first message from the server.
TEST_F(GrpcClientTest, CancelAfterFirstResponse) {
  grpc::ClientContext context;
  auto stream = stub->FullDuplexCall(&context);
  StreamingOutputCallRequest req;
  req.set_response_type(PayloadType::COMPRESSABLE);
  req.add_response_parameters()->set_size(31415);
  newPayload(PayloadType::COMPRESSABLE, 27182, req.mutable_payload());
  EXPECT_TRUE(stream->Write(req));
  StreamingOutputCallResponse res;
  EXPECT_TRUE(stream->Read(&res));
  context.TryCancel();
  auto result = stream->Finish();
  EXPECT_EQ(result.error_code(), grpc::StatusCode::CANCELLED);
}

constexpr auto kLeadingMetadataValue = "test_initial_metadata_value";
constexpr auto kTrailingMetadataValue = "\x0a\x0b\x0a\x0b\x0a\x0b";

// Checks that metadata is echoed back to the client with unary call.
TEST_F(GrpcClientTest, CustomMetadataUnary) {
  metadata_t customMetadataString = {
      {kLeadingMetadataKey, {kLeadingMetadataValue}},
  };
  metadata_t customMetadataBinary = {
      {kTrailingMetadataKey, {kTrailingMetadataValue}},
  };
  customMetadataUnaryTest(customMetadataString, customMetadataBinary);
}

// Checks that metadata is echoed back to the client with server streaming call.
TEST_F(GrpcClientTest, CustomMetadataServerStreaming) {
  metadata_t customMetadataString = {
      {kLeadingMetadataKey, {kLeadingMetadataValue}},
  };
  metadata_t customMetadataBinary = {
      {kTrailingMetadataKey, {kTrailingMetadataValue}},
  };
  customMetadataServerStreamingTest(customMetadataString, customMetadataBinary);
}

// Checks that metadata is echoed back to the client with bidi streaming call.
TEST_F(GrpcClientTest, CustomMetadataFullDuplex) {
  metadata_t customMetadataString = {
      {kLeadingMetadataKey, {kLeadingMetadataValue}},
  };
  metadata_t customMetadataBinary = {
      {kTrailingMetadataKey, {kTrailingMetadataValue}},
  };
  customMetadataFullDuplexTest(customMetadataString, customMetadataBinary);
}

// Adds duplicated metadata keys and checks that the metadata is echoed back
// to the client with unary call.
TEST_F(GrpcClientTest, DuplicatedCustomMetadataUnary) {
  metadata_t customMetadataString = {
      {kLeadingMetadataKey,
       {kLeadingMetadataValue, kLeadingMetadataValue + std::string(";more_stuff")}},
  };
  metadata_t customMetadataBinary = {
      {kTrailingMetadataKey,
       {kTrailingMetadataValue, kTrailingMetadataValue + std::string("\x0a")}},
  };
  customMetadataUnaryTest(customMetadataString, customMetadataBinary);
}

TEST_F(GrpcClientTest, DuplicatedCustomMetadataServerStreaming) {
  metadata_t customMetadataString = {
      {kLeadingMetadataKey,
       {kLeadingMetadataValue, kLeadingMetadataValue + std::string(";more_stuff")}},
  };
  metadata_t customMetadataBinary = {
      {kTrailingMetadataKey,
       {kTrailingMetadataValue, kTrailingMetadataValue + std::string("\x0a")}},
  };
  customMetadataServerStreamingTest(customMetadataString, customMetadataBinary);
}

TEST_F(GrpcClientTest, DuplicatedCustomMetadataFullDuplex) {
  metadata_t customMetadataString = {
      {kLeadingMetadataKey,
       {kLeadingMetadataValue, kLeadingMetadataValue + std::string(";more_stuff")}},
  };
  metadata_t customMetadataBinary = {
      {kTrailingMetadataKey,
       {kTrailingMetadataValue, kTrailingMetadataValue + std::string("\x0a")}},
  };
  customMetadataFullDuplexTest(customMetadataString, customMetadataBinary);
}

// Checks that the status code is propagated back to the client with unary call.
TEST_F(GrpcClientTest, StatusCodeAndMessageUnary) {
  grpc::ClientContext context;
  SimpleRequest req;
  req.mutable_response_status()->set_code(grpc::StatusCode::UNKNOWN);
  req.mutable_response_status()->set_message("test status message");
  SimpleResponse res;
  auto result = stub->UnaryCall(&context, req, &res);
  EXPECT_EQ(result.error_code(), grpc::StatusCode::UNKNOWN);
  EXPECT_EQ(result.error_message(), "test status message");
}

TEST_F(GrpcClientTest, StatusCodeAndMessageFullDuplex) {
  grpc::ClientContext context;
  auto stream = stub->FullDuplexCall(&context);
  StreamingOutputCallRequest req;
  req.mutable_response_status()->set_code(grpc::StatusCode::UNKNOWN);
  req.mutable_response_status()->set_message("test status message");
  EXPECT_TRUE(stream->Write(req));
  stream->WritesDone();
  auto result = stream->Finish();
  EXPECT_EQ(result.error_code(), grpc::StatusCode::UNKNOWN);
  EXPECT_EQ(result.error_message(), "test status message");
}

// Verifies Unicode and whitespace is correctly processed in status message
TEST_F(GrpcClientTest, SpecialStatusMessage) {
  std::string msg = "\t\ntest with whitespace\r\nand Unicode BMP ☺ and non-BMP 😈\t\n";
  grpc::ClientContext context;
  SimpleRequest req;
  req.mutable_response_status()->set_code(grpc::StatusCode::UNKNOWN);
  req.mutable_response_status()->set_message(msg);
  SimpleResponse res;
  auto result = stub->UnaryCall(&context, req, &res);
  EXPECT_EQ(result.error_code(), grpc::StatusCode::UNKNOWN);
  EXPECT_EQ(result.error_message(), msg);
}

// Attempts to call an unimplemented method.
TEST_F(GrpcClientTest, UnimplementedMethod) {
  grpc::ClientContext context;
  google::protobuf::Empty req;
  google::protobuf::Empty res;
  auto result = stub->UnimplementedCall(&context, req, &res);
  EXPECT_EQ(result.error_code(), grpc::StatusCode::UNIMPLEMENTED);
}

// Performs a server streaming RPC that is unimplemented.
TEST_F(GrpcClientTest, UnimplementedServerStreamingMehod) {
  grpc::ClientContext context;
  google::protobuf::Empty req;
  auto stream = stub->UnimplementedStreamingOutputCall(&context, req);
  auto result = stream->Finish();
  EXPECT_EQ(result.error_code(), grpc::StatusCode::UNIMPLEMENTED);
}

// Attempts to call a method from an unimplemented service.
TEST_F(GrpcClientTest, UnimplementedService) {
  UnimplementedService::Stub ustub(channel);
  grpc::ClientContext context;
  google::protobuf::Empty req;
  google::protobuf::Empty res;
  auto result = ustub.UnimplementedCall(&context, req, &res);

  // TODO: 404 should always be UNIMPLEMENTED. Report the bug to gRPC.
  if (result.error_code() != grpc::StatusCode::UNKNOWN) {
    EXPECT_EQ(result.error_code(), grpc::StatusCode::UNIMPLEMENTED) << result.error_message();
  }
}

// Performs a server streaming RPC from an unimplemented service.
TEST_F(GrpcClientTest, UnimplementedServiceStreaming) {
  UnimplementedService::Stub ustub(channel);
  grpc::ClientContext context;
  google::protobuf::Empty req;
  auto stream = ustub.UnimplementedStreamingOutputCall(&context, req);
  auto result = stream->Finish();
  // TODO: 404 should always be UNIMPLEMENTED. Report the bug to gRPC.
  if (result.error_code() != grpc::StatusCode::UNKNOWN) {
    EXPECT_EQ(result.error_code(), grpc::StatusCode::UNIMPLEMENTED) << result.error_message();
  }
}

constexpr auto kNonAsciiError = "soirée 🎉";

void checkNonAsciiError(const grpc::Status& result) {
  EXPECT_EQ(result.error_code(), grpc::StatusCode::RESOURCE_EXHAUSTED);
  EXPECT_EQ(result.error_message(), kNonAsciiError);
  google::rpc::Status status;
  status.ParseFromString(result.error_details());
  EXPECT_EQ(status.code(), grpc::StatusCode::RESOURCE_EXHAUSTED);
  EXPECT_EQ(status.message(), kNonAsciiError);
  ASSERT_EQ(status.details_size(), 1);
  ErrorDetail error_details;
  // status.details().at(0).UnpackTo(&error_details);
  error_details.ParseFromString(status.details().at(0).value());
  EXPECT_EQ(error_details.reason(), kNonAsciiError);
  EXPECT_EQ(error_details.domain(), "connect-crosstest");
}

// Performs a unary RPC that always return a readable non-ASCII error.
TEST_F(GrpcClientTest, FailWithNonAsciiError) {
  grpc::ClientContext context;
  SimpleRequest req;
  req.set_response_type(PayloadType::COMPRESSABLE);
  SimpleResponse res;
  checkNonAsciiError(stub->FailUnaryCall(&context, req, &res));
}

// Performs a server streaming RPC that always return a readable non-ASCII error.
TEST_F(GrpcClientTest, FailServerStreamingWithNonASCIIError) {
  grpc::ClientContext context;
  StreamingOutputCallRequest req;
  req.set_response_type(PayloadType::COMPRESSABLE);
  auto stream = stub->FailStreamingOutputCall(&context, req);
  checkNonAsciiError(stream->Finish());
}

TEST_F(GrpcClientTest, FailServerStreamingAfterResponse) {
  StreamingOutputCallRequest req;
  req.set_response_type(PayloadType::COMPRESSABLE);
  for (int size : kRespSizes) {
    req.add_response_parameters()->set_size(size);
  }
  grpc::ClientContext context;
  auto stream = stub->FailStreamingOutputCall(&context, req);
  StreamingOutputCallResponse res;
  for (int size : kRespSizes) {
    EXPECT_TRUE(stream->Read(&res));
    EXPECT_EQ(res.payload().type(), PayloadType::COMPRESSABLE);
    EXPECT_EQ(res.payload().body().size(), size);
  }
  checkNonAsciiError(stream->Finish());
}

} // namespace
} // namespace connectrpc::conformance
