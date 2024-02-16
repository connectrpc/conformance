// Copyright 2023-2024 The Connect Authors
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

import * as connectrpc_conformance_v1_config_pb from '../../../connectrpc/conformance/v1/config_pb'; // proto import: "connectrpc/conformance/v1/config.proto"


export class ServerCompatRequest extends jspb.Message {
  getProtocol(): connectrpc_conformance_v1_config_pb.Protocol;
  setProtocol(value: connectrpc_conformance_v1_config_pb.Protocol): ServerCompatRequest;

  getHttpVersion(): connectrpc_conformance_v1_config_pb.HTTPVersion;
  setHttpVersion(value: connectrpc_conformance_v1_config_pb.HTTPVersion): ServerCompatRequest;

  getUseTls(): boolean;
  setUseTls(value: boolean): ServerCompatRequest;

  getClientTlsCert(): Uint8Array | string;
  getClientTlsCert_asU8(): Uint8Array;
  getClientTlsCert_asB64(): string;
  setClientTlsCert(value: Uint8Array | string): ServerCompatRequest;

  getMessageReceiveLimit(): number;
  setMessageReceiveLimit(value: number): ServerCompatRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ServerCompatRequest.AsObject;
  static toObject(includeInstance: boolean, msg: ServerCompatRequest): ServerCompatRequest.AsObject;
  static serializeBinaryToWriter(message: ServerCompatRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ServerCompatRequest;
  static deserializeBinaryFromReader(message: ServerCompatRequest, reader: jspb.BinaryReader): ServerCompatRequest;
}

export namespace ServerCompatRequest {
  export type AsObject = {
    protocol: connectrpc_conformance_v1_config_pb.Protocol,
    httpVersion: connectrpc_conformance_v1_config_pb.HTTPVersion,
    useTls: boolean,
    clientTlsCert: Uint8Array | string,
    messageReceiveLimit: number,
  }
}

export class ServerCompatResponse extends jspb.Message {
  getHost(): string;
  setHost(value: string): ServerCompatResponse;

  getPort(): number;
  setPort(value: number): ServerCompatResponse;

  getPemCert(): Uint8Array | string;
  getPemCert_asU8(): Uint8Array;
  getPemCert_asB64(): string;
  setPemCert(value: Uint8Array | string): ServerCompatResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ServerCompatResponse.AsObject;
  static toObject(includeInstance: boolean, msg: ServerCompatResponse): ServerCompatResponse.AsObject;
  static serializeBinaryToWriter(message: ServerCompatResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ServerCompatResponse;
  static deserializeBinaryFromReader(message: ServerCompatResponse, reader: jspb.BinaryReader): ServerCompatResponse;
}

export namespace ServerCompatResponse {
  export type AsObject = {
    host: string,
    port: number,
    pemCert: Uint8Array | string,
  }
}

