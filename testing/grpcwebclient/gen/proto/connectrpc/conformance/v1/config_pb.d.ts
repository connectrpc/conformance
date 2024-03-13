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



export class Config extends jspb.Message {
  getFeatures(): Features | undefined;
  setFeatures(value?: Features): Config;
  hasFeatures(): boolean;
  clearFeatures(): Config;

  getIncludeCasesList(): Array<ConfigCase>;
  setIncludeCasesList(value: Array<ConfigCase>): Config;
  clearIncludeCasesList(): Config;
  addIncludeCases(value?: ConfigCase, index?: number): ConfigCase;

  getExcludeCasesList(): Array<ConfigCase>;
  setExcludeCasesList(value: Array<ConfigCase>): Config;
  clearExcludeCasesList(): Config;
  addExcludeCases(value?: ConfigCase, index?: number): ConfigCase;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Config.AsObject;
  static toObject(includeInstance: boolean, msg: Config): Config.AsObject;
  static serializeBinaryToWriter(message: Config, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Config;
  static deserializeBinaryFromReader(message: Config, reader: jspb.BinaryReader): Config;
}

export namespace Config {
  export type AsObject = {
    features?: Features.AsObject,
    includeCasesList: Array<ConfigCase.AsObject>,
    excludeCasesList: Array<ConfigCase.AsObject>,
  }
}

export class Features extends jspb.Message {
  getVersionsList(): Array<HTTPVersion>;
  setVersionsList(value: Array<HTTPVersion>): Features;
  clearVersionsList(): Features;
  addVersions(value: HTTPVersion, index?: number): Features;

  getProtocolsList(): Array<Protocol>;
  setProtocolsList(value: Array<Protocol>): Features;
  clearProtocolsList(): Features;
  addProtocols(value: Protocol, index?: number): Features;

  getCodecsList(): Array<Codec>;
  setCodecsList(value: Array<Codec>): Features;
  clearCodecsList(): Features;
  addCodecs(value: Codec, index?: number): Features;

  getCompressionsList(): Array<Compression>;
  setCompressionsList(value: Array<Compression>): Features;
  clearCompressionsList(): Features;
  addCompressions(value: Compression, index?: number): Features;

  getStreamTypesList(): Array<StreamType>;
  setStreamTypesList(value: Array<StreamType>): Features;
  clearStreamTypesList(): Features;
  addStreamTypes(value: StreamType, index?: number): Features;

  getSupportsH2c(): boolean;
  setSupportsH2c(value: boolean): Features;
  hasSupportsH2c(): boolean;
  clearSupportsH2c(): Features;

  getSupportsTls(): boolean;
  setSupportsTls(value: boolean): Features;
  hasSupportsTls(): boolean;
  clearSupportsTls(): Features;

  getSupportsTlsClientCerts(): boolean;
  setSupportsTlsClientCerts(value: boolean): Features;
  hasSupportsTlsClientCerts(): boolean;
  clearSupportsTlsClientCerts(): Features;

  getSupportsTrailers(): boolean;
  setSupportsTrailers(value: boolean): Features;
  hasSupportsTrailers(): boolean;
  clearSupportsTrailers(): Features;

  getSupportsHalfDuplexBidiOverHttp1(): boolean;
  setSupportsHalfDuplexBidiOverHttp1(value: boolean): Features;
  hasSupportsHalfDuplexBidiOverHttp1(): boolean;
  clearSupportsHalfDuplexBidiOverHttp1(): Features;

  getSupportsConnectGet(): boolean;
  setSupportsConnectGet(value: boolean): Features;
  hasSupportsConnectGet(): boolean;
  clearSupportsConnectGet(): Features;

  getSupportsMessageReceiveLimit(): boolean;
  setSupportsMessageReceiveLimit(value: boolean): Features;
  hasSupportsMessageReceiveLimit(): boolean;
  clearSupportsMessageReceiveLimit(): Features;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Features.AsObject;
  static toObject(includeInstance: boolean, msg: Features): Features.AsObject;
  static serializeBinaryToWriter(message: Features, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Features;
  static deserializeBinaryFromReader(message: Features, reader: jspb.BinaryReader): Features;
}

export namespace Features {
  export type AsObject = {
    versionsList: Array<HTTPVersion>,
    protocolsList: Array<Protocol>,
    codecsList: Array<Codec>,
    compressionsList: Array<Compression>,
    streamTypesList: Array<StreamType>,
    supportsH2c?: boolean,
    supportsTls?: boolean,
    supportsTlsClientCerts?: boolean,
    supportsTrailers?: boolean,
    supportsHalfDuplexBidiOverHttp1?: boolean,
    supportsConnectGet?: boolean,
    supportsMessageReceiveLimit?: boolean,
  }

  export enum SupportsH2cCase { 
    _SUPPORTS_H2C_NOT_SET = 0,
    SUPPORTS_H2C = 6,
  }

  export enum SupportsTlsCase { 
    _SUPPORTS_TLS_NOT_SET = 0,
    SUPPORTS_TLS = 7,
  }

  export enum SupportsTlsClientCertsCase { 
    _SUPPORTS_TLS_CLIENT_CERTS_NOT_SET = 0,
    SUPPORTS_TLS_CLIENT_CERTS = 8,
  }

  export enum SupportsTrailersCase { 
    _SUPPORTS_TRAILERS_NOT_SET = 0,
    SUPPORTS_TRAILERS = 9,
  }

  export enum SupportsHalfDuplexBidiOverHttp1Case { 
    _SUPPORTS_HALF_DUPLEX_BIDI_OVER_HTTP1_NOT_SET = 0,
    SUPPORTS_HALF_DUPLEX_BIDI_OVER_HTTP1 = 10,
  }

  export enum SupportsConnectGetCase { 
    _SUPPORTS_CONNECT_GET_NOT_SET = 0,
    SUPPORTS_CONNECT_GET = 11,
  }

  export enum SupportsMessageReceiveLimitCase { 
    _SUPPORTS_MESSAGE_RECEIVE_LIMIT_NOT_SET = 0,
    SUPPORTS_MESSAGE_RECEIVE_LIMIT = 12,
  }
}

export class ConfigCase extends jspb.Message {
  getVersion(): HTTPVersion;
  setVersion(value: HTTPVersion): ConfigCase;

  getProtocol(): Protocol;
  setProtocol(value: Protocol): ConfigCase;

  getCodec(): Codec;
  setCodec(value: Codec): ConfigCase;

  getCompression(): Compression;
  setCompression(value: Compression): ConfigCase;

  getStreamType(): StreamType;
  setStreamType(value: StreamType): ConfigCase;

  getUseTls(): boolean;
  setUseTls(value: boolean): ConfigCase;
  hasUseTls(): boolean;
  clearUseTls(): ConfigCase;

  getUseTlsClientCerts(): boolean;
  setUseTlsClientCerts(value: boolean): ConfigCase;
  hasUseTlsClientCerts(): boolean;
  clearUseTlsClientCerts(): ConfigCase;

  getUseMessageReceiveLimit(): boolean;
  setUseMessageReceiveLimit(value: boolean): ConfigCase;
  hasUseMessageReceiveLimit(): boolean;
  clearUseMessageReceiveLimit(): ConfigCase;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ConfigCase.AsObject;
  static toObject(includeInstance: boolean, msg: ConfigCase): ConfigCase.AsObject;
  static serializeBinaryToWriter(message: ConfigCase, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ConfigCase;
  static deserializeBinaryFromReader(message: ConfigCase, reader: jspb.BinaryReader): ConfigCase;
}

export namespace ConfigCase {
  export type AsObject = {
    version: HTTPVersion,
    protocol: Protocol,
    codec: Codec,
    compression: Compression,
    streamType: StreamType,
    useTls?: boolean,
    useTlsClientCerts?: boolean,
    useMessageReceiveLimit?: boolean,
  }

  export enum UseTlsCase { 
    _USE_TLS_NOT_SET = 0,
    USE_TLS = 6,
  }

  export enum UseTlsClientCertsCase { 
    _USE_TLS_CLIENT_CERTS_NOT_SET = 0,
    USE_TLS_CLIENT_CERTS = 7,
  }

  export enum UseMessageReceiveLimitCase { 
    _USE_MESSAGE_RECEIVE_LIMIT_NOT_SET = 0,
    USE_MESSAGE_RECEIVE_LIMIT = 8,
  }
}

export class TLSCreds extends jspb.Message {
  getCert(): Uint8Array | string;
  getCert_asU8(): Uint8Array;
  getCert_asB64(): string;
  setCert(value: Uint8Array | string): TLSCreds;

  getKey(): Uint8Array | string;
  getKey_asU8(): Uint8Array;
  getKey_asB64(): string;
  setKey(value: Uint8Array | string): TLSCreds;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): TLSCreds.AsObject;
  static toObject(includeInstance: boolean, msg: TLSCreds): TLSCreds.AsObject;
  static serializeBinaryToWriter(message: TLSCreds, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): TLSCreds;
  static deserializeBinaryFromReader(message: TLSCreds, reader: jspb.BinaryReader): TLSCreds;
}

export namespace TLSCreds {
  export type AsObject = {
    cert: Uint8Array | string,
    key: Uint8Array | string,
  }
}

export enum HTTPVersion { 
  HTTP_VERSION_UNSPECIFIED = 0,
  HTTP_VERSION_1 = 1,
  HTTP_VERSION_2 = 2,
  HTTP_VERSION_3 = 3,
}
export enum Protocol { 
  PROTOCOL_UNSPECIFIED = 0,
  PROTOCOL_CONNECT = 1,
  PROTOCOL_GRPC = 2,
  PROTOCOL_GRPC_WEB = 3,
}
export enum Codec { 
  CODEC_UNSPECIFIED = 0,
  CODEC_PROTO = 1,
  CODEC_JSON = 2,
  CODEC_TEXT = 3,
}
export enum Compression { 
  COMPRESSION_UNSPECIFIED = 0,
  COMPRESSION_IDENTITY = 1,
  COMPRESSION_GZIP = 2,
  COMPRESSION_BR = 3,
  COMPRESSION_ZSTD = 4,
  COMPRESSION_DEFLATE = 5,
  COMPRESSION_SNAPPY = 6,
}
export enum StreamType { 
  STREAM_TYPE_UNSPECIFIED = 0,
  STREAM_TYPE_UNARY = 1,
  STREAM_TYPE_CLIENT_STREAM = 2,
  STREAM_TYPE_SERVER_STREAM = 3,
  STREAM_TYPE_HALF_DUPLEX_BIDI_STREAM = 4,
  STREAM_TYPE_FULL_DUPLEX_BIDI_STREAM = 5,
}
export enum Code { 
  CODE_UNSPECIFIED = 0,
  CODE_CANCELED = 1,
  CODE_UNKNOWN = 2,
  CODE_INVALID_ARGUMENT = 3,
  CODE_DEADLINE_EXCEEDED = 4,
  CODE_NOT_FOUND = 5,
  CODE_ALREADY_EXISTS = 6,
  CODE_PERMISSION_DENIED = 7,
  CODE_RESOURCE_EXHAUSTED = 8,
  CODE_FAILED_PRECONDITION = 9,
  CODE_ABORTED = 10,
  CODE_OUT_OF_RANGE = 11,
  CODE_UNIMPLEMENTED = 12,
  CODE_INTERNAL = 13,
  CODE_UNAVAILABLE = 14,
  CODE_DATA_LOSS = 15,
  CODE_UNAUTHENTICATED = 16,
}
