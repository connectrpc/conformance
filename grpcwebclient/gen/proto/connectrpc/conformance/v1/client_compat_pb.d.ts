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
import * as connectrpc_conformance_v1_service_pb from '../../../connectrpc/conformance/v1/service_pb'; // proto import: "connectrpc/conformance/v1/service.proto"
import * as google_protobuf_any_pb from 'google-protobuf/google/protobuf/any_pb'; // proto import: "google/protobuf/any.proto"
import * as google_protobuf_empty_pb from 'google-protobuf/google/protobuf/empty_pb'; // proto import: "google/protobuf/empty.proto"
import * as google_protobuf_struct_pb from 'google-protobuf/google/protobuf/struct_pb'; // proto import: "google/protobuf/struct.proto"


export class ClientCompatRequest extends jspb.Message {
  getTestName(): string;
  setTestName(value: string): ClientCompatRequest;

  getHttpVersion(): connectrpc_conformance_v1_config_pb.HTTPVersion;
  setHttpVersion(value: connectrpc_conformance_v1_config_pb.HTTPVersion): ClientCompatRequest;

  getProtocol(): connectrpc_conformance_v1_config_pb.Protocol;
  setProtocol(value: connectrpc_conformance_v1_config_pb.Protocol): ClientCompatRequest;

  getCodec(): connectrpc_conformance_v1_config_pb.Codec;
  setCodec(value: connectrpc_conformance_v1_config_pb.Codec): ClientCompatRequest;

  getCompression(): connectrpc_conformance_v1_config_pb.Compression;
  setCompression(value: connectrpc_conformance_v1_config_pb.Compression): ClientCompatRequest;

  getHost(): string;
  setHost(value: string): ClientCompatRequest;

  getPort(): number;
  setPort(value: number): ClientCompatRequest;

  getServerTlsCert(): Uint8Array | string;
  getServerTlsCert_asU8(): Uint8Array;
  getServerTlsCert_asB64(): string;
  setServerTlsCert(value: Uint8Array | string): ClientCompatRequest;

  getClientTlsCreds(): ClientCompatRequest.TLSCreds | undefined;
  setClientTlsCreds(value?: ClientCompatRequest.TLSCreds): ClientCompatRequest;
  hasClientTlsCreds(): boolean;
  clearClientTlsCreds(): ClientCompatRequest;

  getMessageReceiveLimit(): number;
  setMessageReceiveLimit(value: number): ClientCompatRequest;

  getService(): string;
  setService(value: string): ClientCompatRequest;

  getMethod(): string;
  setMethod(value: string): ClientCompatRequest;

  getStreamType(): connectrpc_conformance_v1_config_pb.StreamType;
  setStreamType(value: connectrpc_conformance_v1_config_pb.StreamType): ClientCompatRequest;

  getUseGetHttpMethod(): boolean;
  setUseGetHttpMethod(value: boolean): ClientCompatRequest;

  getRequestHeadersList(): Array<connectrpc_conformance_v1_service_pb.Header>;
  setRequestHeadersList(value: Array<connectrpc_conformance_v1_service_pb.Header>): ClientCompatRequest;
  clearRequestHeadersList(): ClientCompatRequest;
  addRequestHeaders(value?: connectrpc_conformance_v1_service_pb.Header, index?: number): connectrpc_conformance_v1_service_pb.Header;

  getRequestMessagesList(): Array<google_protobuf_any_pb.Any>;
  setRequestMessagesList(value: Array<google_protobuf_any_pb.Any>): ClientCompatRequest;
  clearRequestMessagesList(): ClientCompatRequest;
  addRequestMessages(value?: google_protobuf_any_pb.Any, index?: number): google_protobuf_any_pb.Any;

  getTimeoutMs(): number;
  setTimeoutMs(value: number): ClientCompatRequest;
  hasTimeoutMs(): boolean;
  clearTimeoutMs(): ClientCompatRequest;

  getRequestDelayMs(): number;
  setRequestDelayMs(value: number): ClientCompatRequest;

  getCancel(): ClientCompatRequest.Cancel | undefined;
  setCancel(value?: ClientCompatRequest.Cancel): ClientCompatRequest;
  hasCancel(): boolean;
  clearCancel(): ClientCompatRequest;

  getRawRequest(): connectrpc_conformance_v1_service_pb.RawHTTPRequest | undefined;
  setRawRequest(value?: connectrpc_conformance_v1_service_pb.RawHTTPRequest): ClientCompatRequest;
  hasRawRequest(): boolean;
  clearRawRequest(): ClientCompatRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ClientCompatRequest.AsObject;
  static toObject(includeInstance: boolean, msg: ClientCompatRequest): ClientCompatRequest.AsObject;
  static serializeBinaryToWriter(message: ClientCompatRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ClientCompatRequest;
  static deserializeBinaryFromReader(message: ClientCompatRequest, reader: jspb.BinaryReader): ClientCompatRequest;
}

export namespace ClientCompatRequest {
  export type AsObject = {
    testName: string,
    httpVersion: connectrpc_conformance_v1_config_pb.HTTPVersion,
    protocol: connectrpc_conformance_v1_config_pb.Protocol,
    codec: connectrpc_conformance_v1_config_pb.Codec,
    compression: connectrpc_conformance_v1_config_pb.Compression,
    host: string,
    port: number,
    serverTlsCert: Uint8Array | string,
    clientTlsCreds?: ClientCompatRequest.TLSCreds.AsObject,
    messageReceiveLimit: number,
    service: string,
    method: string,
    streamType: connectrpc_conformance_v1_config_pb.StreamType,
    useGetHttpMethod: boolean,
    requestHeadersList: Array<connectrpc_conformance_v1_service_pb.Header.AsObject>,
    requestMessagesList: Array<google_protobuf_any_pb.Any.AsObject>,
    timeoutMs?: number,
    requestDelayMs: number,
    cancel?: ClientCompatRequest.Cancel.AsObject,
    rawRequest?: connectrpc_conformance_v1_service_pb.RawHTTPRequest.AsObject,
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


  export class Cancel extends jspb.Message {
    getBeforeCloseSend(): google_protobuf_empty_pb.Empty | undefined;
    setBeforeCloseSend(value?: google_protobuf_empty_pb.Empty): Cancel;
    hasBeforeCloseSend(): boolean;
    clearBeforeCloseSend(): Cancel;

    getAfterCloseSendMs(): number;
    setAfterCloseSendMs(value: number): Cancel;

    getAfterNumResponses(): number;
    setAfterNumResponses(value: number): Cancel;

    getCancelTimingCase(): Cancel.CancelTimingCase;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Cancel.AsObject;
    static toObject(includeInstance: boolean, msg: Cancel): Cancel.AsObject;
    static serializeBinaryToWriter(message: Cancel, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Cancel;
    static deserializeBinaryFromReader(message: Cancel, reader: jspb.BinaryReader): Cancel;
  }

  export namespace Cancel {
    export type AsObject = {
      beforeCloseSend?: google_protobuf_empty_pb.Empty.AsObject,
      afterCloseSendMs: number,
      afterNumResponses: number,
    }

    export enum CancelTimingCase { 
      CANCEL_TIMING_NOT_SET = 0,
      BEFORE_CLOSE_SEND = 1,
      AFTER_CLOSE_SEND_MS = 2,
      AFTER_NUM_RESPONSES = 3,
    }
  }


  export enum TimeoutMsCase { 
    _TIMEOUT_MS_NOT_SET = 0,
    TIMEOUT_MS = 17,
  }
}

export class ClientCompatResponse extends jspb.Message {
  getTestName(): string;
  setTestName(value: string): ClientCompatResponse;

  getResponse(): ClientResponseResult | undefined;
  setResponse(value?: ClientResponseResult): ClientCompatResponse;
  hasResponse(): boolean;
  clearResponse(): ClientCompatResponse;

  getError(): ClientErrorResult | undefined;
  setError(value?: ClientErrorResult): ClientCompatResponse;
  hasError(): boolean;
  clearError(): ClientCompatResponse;

  getFeedbackList(): Array<string>;
  setFeedbackList(value: Array<string>): ClientCompatResponse;
  clearFeedbackList(): ClientCompatResponse;
  addFeedback(value: string, index?: number): ClientCompatResponse;

  getResultCase(): ClientCompatResponse.ResultCase;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ClientCompatResponse.AsObject;
  static toObject(includeInstance: boolean, msg: ClientCompatResponse): ClientCompatResponse.AsObject;
  static serializeBinaryToWriter(message: ClientCompatResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ClientCompatResponse;
  static deserializeBinaryFromReader(message: ClientCompatResponse, reader: jspb.BinaryReader): ClientCompatResponse;
}

export namespace ClientCompatResponse {
  export type AsObject = {
    testName: string,
    response?: ClientResponseResult.AsObject,
    error?: ClientErrorResult.AsObject,
    feedbackList: Array<string>,
  }

  export enum ResultCase { 
    RESULT_NOT_SET = 0,
    RESPONSE = 2,
    ERROR = 3,
  }
}

export class ClientResponseResult extends jspb.Message {
  getResponseHeadersList(): Array<connectrpc_conformance_v1_service_pb.Header>;
  setResponseHeadersList(value: Array<connectrpc_conformance_v1_service_pb.Header>): ClientResponseResult;
  clearResponseHeadersList(): ClientResponseResult;
  addResponseHeaders(value?: connectrpc_conformance_v1_service_pb.Header, index?: number): connectrpc_conformance_v1_service_pb.Header;

  getPayloadsList(): Array<connectrpc_conformance_v1_service_pb.ConformancePayload>;
  setPayloadsList(value: Array<connectrpc_conformance_v1_service_pb.ConformancePayload>): ClientResponseResult;
  clearPayloadsList(): ClientResponseResult;
  addPayloads(value?: connectrpc_conformance_v1_service_pb.ConformancePayload, index?: number): connectrpc_conformance_v1_service_pb.ConformancePayload;

  getError(): connectrpc_conformance_v1_service_pb.Error | undefined;
  setError(value?: connectrpc_conformance_v1_service_pb.Error): ClientResponseResult;
  hasError(): boolean;
  clearError(): ClientResponseResult;

  getResponseTrailersList(): Array<connectrpc_conformance_v1_service_pb.Header>;
  setResponseTrailersList(value: Array<connectrpc_conformance_v1_service_pb.Header>): ClientResponseResult;
  clearResponseTrailersList(): ClientResponseResult;
  addResponseTrailers(value?: connectrpc_conformance_v1_service_pb.Header, index?: number): connectrpc_conformance_v1_service_pb.Header;

  getNumUnsentRequests(): number;
  setNumUnsentRequests(value: number): ClientResponseResult;

  getWireDetails(): WireDetails | undefined;
  setWireDetails(value?: WireDetails): ClientResponseResult;
  hasWireDetails(): boolean;
  clearWireDetails(): ClientResponseResult;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ClientResponseResult.AsObject;
  static toObject(includeInstance: boolean, msg: ClientResponseResult): ClientResponseResult.AsObject;
  static serializeBinaryToWriter(message: ClientResponseResult, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ClientResponseResult;
  static deserializeBinaryFromReader(message: ClientResponseResult, reader: jspb.BinaryReader): ClientResponseResult;
}

export namespace ClientResponseResult {
  export type AsObject = {
    responseHeadersList: Array<connectrpc_conformance_v1_service_pb.Header.AsObject>,
    payloadsList: Array<connectrpc_conformance_v1_service_pb.ConformancePayload.AsObject>,
    error?: connectrpc_conformance_v1_service_pb.Error.AsObject,
    responseTrailersList: Array<connectrpc_conformance_v1_service_pb.Header.AsObject>,
    numUnsentRequests: number,
    wireDetails?: WireDetails.AsObject,
  }
}

export class ClientErrorResult extends jspb.Message {
  getMessage(): string;
  setMessage(value: string): ClientErrorResult;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ClientErrorResult.AsObject;
  static toObject(includeInstance: boolean, msg: ClientErrorResult): ClientErrorResult.AsObject;
  static serializeBinaryToWriter(message: ClientErrorResult, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ClientErrorResult;
  static deserializeBinaryFromReader(message: ClientErrorResult, reader: jspb.BinaryReader): ClientErrorResult;
}

export namespace ClientErrorResult {
  export type AsObject = {
    message: string,
  }
}

export class WireDetails extends jspb.Message {
  getActualStatusCode(): number;
  setActualStatusCode(value: number): WireDetails;

  getConnectErrorRaw(): google_protobuf_struct_pb.Struct | undefined;
  setConnectErrorRaw(value?: google_protobuf_struct_pb.Struct): WireDetails;
  hasConnectErrorRaw(): boolean;
  clearConnectErrorRaw(): WireDetails;

  getActualHttpTrailersList(): Array<connectrpc_conformance_v1_service_pb.Header>;
  setActualHttpTrailersList(value: Array<connectrpc_conformance_v1_service_pb.Header>): WireDetails;
  clearActualHttpTrailersList(): WireDetails;
  addActualHttpTrailers(value?: connectrpc_conformance_v1_service_pb.Header, index?: number): connectrpc_conformance_v1_service_pb.Header;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): WireDetails.AsObject;
  static toObject(includeInstance: boolean, msg: WireDetails): WireDetails.AsObject;
  static serializeBinaryToWriter(message: WireDetails, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): WireDetails;
  static deserializeBinaryFromReader(message: WireDetails, reader: jspb.BinaryReader): WireDetails;
}

export namespace WireDetails {
  export type AsObject = {
    actualStatusCode: number,
    connectErrorRaw?: google_protobuf_struct_pb.Struct.AsObject,
    actualHttpTrailersList: Array<connectrpc_conformance_v1_service_pb.Header.AsObject>,
  }
}

