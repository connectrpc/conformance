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

import * as jspb from 'google-protobuf'



export class ServerMetadata extends jspb.Message {
  getHost(): string;
  setHost(value: string): ServerMetadata;

  getProtocolsList(): Array<ProtocolSupport>;
  setProtocolsList(value: Array<ProtocolSupport>): ServerMetadata;
  clearProtocolsList(): ServerMetadata;
  addProtocols(value?: ProtocolSupport, index?: number): ProtocolSupport;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ServerMetadata.AsObject;
  static toObject(includeInstance: boolean, msg: ServerMetadata): ServerMetadata.AsObject;
  static serializeBinaryToWriter(message: ServerMetadata, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ServerMetadata;
  static deserializeBinaryFromReader(message: ServerMetadata, reader: jspb.BinaryReader): ServerMetadata;
}

export namespace ServerMetadata {
  export type AsObject = {
    host: string,
    protocolsList: Array<ProtocolSupport.AsObject>,
  }
}

export class ProtocolSupport extends jspb.Message {
  getProtocol(): Protocol;
  setProtocol(value: Protocol): ProtocolSupport;

  getHttpVersionsList(): Array<HTTPVersion>;
  setHttpVersionsList(value: Array<HTTPVersion>): ProtocolSupport;
  clearHttpVersionsList(): ProtocolSupport;
  addHttpVersions(value?: HTTPVersion, index?: number): HTTPVersion;

  getPort(): string;
  setPort(value: string): ProtocolSupport;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ProtocolSupport.AsObject;
  static toObject(includeInstance: boolean, msg: ProtocolSupport): ProtocolSupport.AsObject;
  static serializeBinaryToWriter(message: ProtocolSupport, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ProtocolSupport;
  static deserializeBinaryFromReader(message: ProtocolSupport, reader: jspb.BinaryReader): ProtocolSupport;
}

export namespace ProtocolSupport {
  export type AsObject = {
    protocol: Protocol,
    httpVersionsList: Array<HTTPVersion.AsObject>,
    port: string,
  }
}

export class HTTPVersion extends jspb.Message {
  getMajor(): number;
  setMajor(value: number): HTTPVersion;

  getMinor(): number;
  setMinor(value: number): HTTPVersion;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): HTTPVersion.AsObject;
  static toObject(includeInstance: boolean, msg: HTTPVersion): HTTPVersion.AsObject;
  static serializeBinaryToWriter(message: HTTPVersion, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): HTTPVersion;
  static deserializeBinaryFromReader(message: HTTPVersion, reader: jspb.BinaryReader): HTTPVersion;
}

export namespace HTTPVersion {
  export type AsObject = {
    major: number,
    minor: number,
  }
}

export enum Protocol { 
  PROTOCOL_UNSPECIFIED = 0,
  PROTOCOL_GRPC = 1,
  PROTOCOL_GRPC_WEB = 2,
}
