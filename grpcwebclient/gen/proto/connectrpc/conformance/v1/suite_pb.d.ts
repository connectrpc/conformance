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

import * as jspb from 'google-protobuf'

import * as connectrpc_conformance_v1_client_compat_pb from '../../../connectrpc/conformance/v1/client_compat_pb'; // proto import: "connectrpc/conformance/v1/client_compat.proto"
import * as connectrpc_conformance_v1_config_pb from '../../../connectrpc/conformance/v1/config_pb'; // proto import: "connectrpc/conformance/v1/config.proto"


export class TestSuite extends jspb.Message {
  getName(): string;
  setName(value: string): TestSuite;

  getMode(): TestSuite.TestMode;
  setMode(value: TestSuite.TestMode): TestSuite;

  getTestCasesList(): Array<TestCase>;
  setTestCasesList(value: Array<TestCase>): TestSuite;
  clearTestCasesList(): TestSuite;
  addTestCases(value?: TestCase, index?: number): TestCase;

  getRelevantProtocolsList(): Array<connectrpc_conformance_v1_config_pb.Protocol>;
  setRelevantProtocolsList(value: Array<connectrpc_conformance_v1_config_pb.Protocol>): TestSuite;
  clearRelevantProtocolsList(): TestSuite;
  addRelevantProtocols(value: connectrpc_conformance_v1_config_pb.Protocol, index?: number): TestSuite;

  getRelevantHttpVersionsList(): Array<connectrpc_conformance_v1_config_pb.HTTPVersion>;
  setRelevantHttpVersionsList(value: Array<connectrpc_conformance_v1_config_pb.HTTPVersion>): TestSuite;
  clearRelevantHttpVersionsList(): TestSuite;
  addRelevantHttpVersions(value: connectrpc_conformance_v1_config_pb.HTTPVersion, index?: number): TestSuite;

  getRelevantCodecsList(): Array<connectrpc_conformance_v1_config_pb.Codec>;
  setRelevantCodecsList(value: Array<connectrpc_conformance_v1_config_pb.Codec>): TestSuite;
  clearRelevantCodecsList(): TestSuite;
  addRelevantCodecs(value: connectrpc_conformance_v1_config_pb.Codec, index?: number): TestSuite;

  getRelevantCompressionsList(): Array<connectrpc_conformance_v1_config_pb.Compression>;
  setRelevantCompressionsList(value: Array<connectrpc_conformance_v1_config_pb.Compression>): TestSuite;
  clearRelevantCompressionsList(): TestSuite;
  addRelevantCompressions(value: connectrpc_conformance_v1_config_pb.Compression, index?: number): TestSuite;

  getConnectVersionMode(): TestSuite.ConnectVersionMode;
  setConnectVersionMode(value: TestSuite.ConnectVersionMode): TestSuite;

  getReliesOnTls(): boolean;
  setReliesOnTls(value: boolean): TestSuite;

  getReliesOnTlsClientCerts(): boolean;
  setReliesOnTlsClientCerts(value: boolean): TestSuite;

  getReliesOnConnectGet(): boolean;
  setReliesOnConnectGet(value: boolean): TestSuite;

  getReliesOnMessageReceiveLimit(): boolean;
  setReliesOnMessageReceiveLimit(value: boolean): TestSuite;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): TestSuite.AsObject;
  static toObject(includeInstance: boolean, msg: TestSuite): TestSuite.AsObject;
  static serializeBinaryToWriter(message: TestSuite, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): TestSuite;
  static deserializeBinaryFromReader(message: TestSuite, reader: jspb.BinaryReader): TestSuite;
}

export namespace TestSuite {
  export type AsObject = {
    name: string,
    mode: TestSuite.TestMode,
    testCasesList: Array<TestCase.AsObject>,
    relevantProtocolsList: Array<connectrpc_conformance_v1_config_pb.Protocol>,
    relevantHttpVersionsList: Array<connectrpc_conformance_v1_config_pb.HTTPVersion>,
    relevantCodecsList: Array<connectrpc_conformance_v1_config_pb.Codec>,
    relevantCompressionsList: Array<connectrpc_conformance_v1_config_pb.Compression>,
    connectVersionMode: TestSuite.ConnectVersionMode,
    reliesOnTls: boolean,
    reliesOnTlsClientCerts: boolean,
    reliesOnConnectGet: boolean,
    reliesOnMessageReceiveLimit: boolean,
  }

  export enum TestMode { 
    TEST_MODE_UNSPECIFIED = 0,
    TEST_MODE_CLIENT = 1,
    TEST_MODE_SERVER = 2,
  }

  export enum ConnectVersionMode { 
    CONNECT_VERSION_MODE_UNSPECIFIED = 0,
    CONNECT_VERSION_MODE_REQUIRE = 1,
    CONNECT_VERSION_MODE_IGNORE = 2,
  }
}

export class TestCase extends jspb.Message {
  getRequest(): connectrpc_conformance_v1_client_compat_pb.ClientCompatRequest | undefined;
  setRequest(value?: connectrpc_conformance_v1_client_compat_pb.ClientCompatRequest): TestCase;
  hasRequest(): boolean;
  clearRequest(): TestCase;

  getExpandRequestsList(): Array<TestCase.ExpandedSize>;
  setExpandRequestsList(value: Array<TestCase.ExpandedSize>): TestCase;
  clearExpandRequestsList(): TestCase;
  addExpandRequests(value?: TestCase.ExpandedSize, index?: number): TestCase.ExpandedSize;

  getExpectedResponse(): connectrpc_conformance_v1_client_compat_pb.ClientResponseResult | undefined;
  setExpectedResponse(value?: connectrpc_conformance_v1_client_compat_pb.ClientResponseResult): TestCase;
  hasExpectedResponse(): boolean;
  clearExpectedResponse(): TestCase;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): TestCase.AsObject;
  static toObject(includeInstance: boolean, msg: TestCase): TestCase.AsObject;
  static serializeBinaryToWriter(message: TestCase, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): TestCase;
  static deserializeBinaryFromReader(message: TestCase, reader: jspb.BinaryReader): TestCase;
}

export namespace TestCase {
  export type AsObject = {
    request?: connectrpc_conformance_v1_client_compat_pb.ClientCompatRequest.AsObject,
    expandRequestsList: Array<TestCase.ExpandedSize.AsObject>,
    expectedResponse?: connectrpc_conformance_v1_client_compat_pb.ClientResponseResult.AsObject,
  }

  export class ExpandedSize extends jspb.Message {
    getSizeRelativeToLimit(): number;
    setSizeRelativeToLimit(value: number): ExpandedSize;
    hasSizeRelativeToLimit(): boolean;
    clearSizeRelativeToLimit(): ExpandedSize;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ExpandedSize.AsObject;
    static toObject(includeInstance: boolean, msg: ExpandedSize): ExpandedSize.AsObject;
    static serializeBinaryToWriter(message: ExpandedSize, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ExpandedSize;
    static deserializeBinaryFromReader(message: ExpandedSize, reader: jspb.BinaryReader): ExpandedSize;
  }

  export namespace ExpandedSize {
    export type AsObject = {
      sizeRelativeToLimit?: number,
    }

    export enum SizeRelativeToLimitCase { 
      _SIZE_RELATIVE_TO_LIMIT_NOT_SET = 0,
      SIZE_RELATIVE_TO_LIMIT = 1,
    }
  }

}

