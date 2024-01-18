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
import * as google_protobuf_any_pb from 'google-protobuf/google/protobuf/any_pb'; // proto import: "google/protobuf/any.proto"


export class UnaryResponseDefinition extends jspb.Message {
  getResponseHeadersList(): Array<Header>;
  setResponseHeadersList(value: Array<Header>): UnaryResponseDefinition;
  clearResponseHeadersList(): UnaryResponseDefinition;
  addResponseHeaders(value?: Header, index?: number): Header;

  getResponseData(): Uint8Array | string;
  getResponseData_asU8(): Uint8Array;
  getResponseData_asB64(): string;
  setResponseData(value: Uint8Array | string): UnaryResponseDefinition;

  getError(): Error | undefined;
  setError(value?: Error): UnaryResponseDefinition;
  hasError(): boolean;
  clearError(): UnaryResponseDefinition;

  getResponseTrailersList(): Array<Header>;
  setResponseTrailersList(value: Array<Header>): UnaryResponseDefinition;
  clearResponseTrailersList(): UnaryResponseDefinition;
  addResponseTrailers(value?: Header, index?: number): Header;

  getResponseDelayMs(): number;
  setResponseDelayMs(value: number): UnaryResponseDefinition;

  getRawResponse(): RawHTTPResponse | undefined;
  setRawResponse(value?: RawHTTPResponse): UnaryResponseDefinition;
  hasRawResponse(): boolean;
  clearRawResponse(): UnaryResponseDefinition;

  getResponseCase(): UnaryResponseDefinition.ResponseCase;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): UnaryResponseDefinition.AsObject;
  static toObject(includeInstance: boolean, msg: UnaryResponseDefinition): UnaryResponseDefinition.AsObject;
  static serializeBinaryToWriter(message: UnaryResponseDefinition, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): UnaryResponseDefinition;
  static deserializeBinaryFromReader(message: UnaryResponseDefinition, reader: jspb.BinaryReader): UnaryResponseDefinition;
}

export namespace UnaryResponseDefinition {
  export type AsObject = {
    responseHeadersList: Array<Header.AsObject>,
    responseData: Uint8Array | string,
    error?: Error.AsObject,
    responseTrailersList: Array<Header.AsObject>,
    responseDelayMs: number,
    rawResponse?: RawHTTPResponse.AsObject,
  }

  export enum ResponseCase { 
    RESPONSE_NOT_SET = 0,
    RESPONSE_DATA = 2,
    ERROR = 3,
  }
}

export class StreamResponseDefinition extends jspb.Message {
  getResponseHeadersList(): Array<Header>;
  setResponseHeadersList(value: Array<Header>): StreamResponseDefinition;
  clearResponseHeadersList(): StreamResponseDefinition;
  addResponseHeaders(value?: Header, index?: number): Header;

  getResponseDataList(): Array<Uint8Array | string>;
  setResponseDataList(value: Array<Uint8Array | string>): StreamResponseDefinition;
  clearResponseDataList(): StreamResponseDefinition;
  addResponseData(value: Uint8Array | string, index?: number): StreamResponseDefinition;

  getResponseDelayMs(): number;
  setResponseDelayMs(value: number): StreamResponseDefinition;

  getError(): Error | undefined;
  setError(value?: Error): StreamResponseDefinition;
  hasError(): boolean;
  clearError(): StreamResponseDefinition;

  getResponseTrailersList(): Array<Header>;
  setResponseTrailersList(value: Array<Header>): StreamResponseDefinition;
  clearResponseTrailersList(): StreamResponseDefinition;
  addResponseTrailers(value?: Header, index?: number): Header;

  getRawResponse(): RawHTTPResponse | undefined;
  setRawResponse(value?: RawHTTPResponse): StreamResponseDefinition;
  hasRawResponse(): boolean;
  clearRawResponse(): StreamResponseDefinition;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): StreamResponseDefinition.AsObject;
  static toObject(includeInstance: boolean, msg: StreamResponseDefinition): StreamResponseDefinition.AsObject;
  static serializeBinaryToWriter(message: StreamResponseDefinition, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): StreamResponseDefinition;
  static deserializeBinaryFromReader(message: StreamResponseDefinition, reader: jspb.BinaryReader): StreamResponseDefinition;
}

export namespace StreamResponseDefinition {
  export type AsObject = {
    responseHeadersList: Array<Header.AsObject>,
    responseDataList: Array<Uint8Array | string>,
    responseDelayMs: number,
    error?: Error.AsObject,
    responseTrailersList: Array<Header.AsObject>,
    rawResponse?: RawHTTPResponse.AsObject,
  }
}

export class UnaryRequest extends jspb.Message {
  getResponseDefinition(): UnaryResponseDefinition | undefined;
  setResponseDefinition(value?: UnaryResponseDefinition): UnaryRequest;
  hasResponseDefinition(): boolean;
  clearResponseDefinition(): UnaryRequest;

  getRequestData(): Uint8Array | string;
  getRequestData_asU8(): Uint8Array;
  getRequestData_asB64(): string;
  setRequestData(value: Uint8Array | string): UnaryRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): UnaryRequest.AsObject;
  static toObject(includeInstance: boolean, msg: UnaryRequest): UnaryRequest.AsObject;
  static serializeBinaryToWriter(message: UnaryRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): UnaryRequest;
  static deserializeBinaryFromReader(message: UnaryRequest, reader: jspb.BinaryReader): UnaryRequest;
}

export namespace UnaryRequest {
  export type AsObject = {
    responseDefinition?: UnaryResponseDefinition.AsObject,
    requestData: Uint8Array | string,
  }
}

export class UnaryResponse extends jspb.Message {
  getPayload(): ConformancePayload | undefined;
  setPayload(value?: ConformancePayload): UnaryResponse;
  hasPayload(): boolean;
  clearPayload(): UnaryResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): UnaryResponse.AsObject;
  static toObject(includeInstance: boolean, msg: UnaryResponse): UnaryResponse.AsObject;
  static serializeBinaryToWriter(message: UnaryResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): UnaryResponse;
  static deserializeBinaryFromReader(message: UnaryResponse, reader: jspb.BinaryReader): UnaryResponse;
}

export namespace UnaryResponse {
  export type AsObject = {
    payload?: ConformancePayload.AsObject,
  }
}

export class IdempotentUnaryRequest extends jspb.Message {
  getResponseDefinition(): UnaryResponseDefinition | undefined;
  setResponseDefinition(value?: UnaryResponseDefinition): IdempotentUnaryRequest;
  hasResponseDefinition(): boolean;
  clearResponseDefinition(): IdempotentUnaryRequest;

  getRequestData(): Uint8Array | string;
  getRequestData_asU8(): Uint8Array;
  getRequestData_asB64(): string;
  setRequestData(value: Uint8Array | string): IdempotentUnaryRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): IdempotentUnaryRequest.AsObject;
  static toObject(includeInstance: boolean, msg: IdempotentUnaryRequest): IdempotentUnaryRequest.AsObject;
  static serializeBinaryToWriter(message: IdempotentUnaryRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): IdempotentUnaryRequest;
  static deserializeBinaryFromReader(message: IdempotentUnaryRequest, reader: jspb.BinaryReader): IdempotentUnaryRequest;
}

export namespace IdempotentUnaryRequest {
  export type AsObject = {
    responseDefinition?: UnaryResponseDefinition.AsObject,
    requestData: Uint8Array | string,
  }
}

export class IdempotentUnaryResponse extends jspb.Message {
  getPayload(): ConformancePayload | undefined;
  setPayload(value?: ConformancePayload): IdempotentUnaryResponse;
  hasPayload(): boolean;
  clearPayload(): IdempotentUnaryResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): IdempotentUnaryResponse.AsObject;
  static toObject(includeInstance: boolean, msg: IdempotentUnaryResponse): IdempotentUnaryResponse.AsObject;
  static serializeBinaryToWriter(message: IdempotentUnaryResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): IdempotentUnaryResponse;
  static deserializeBinaryFromReader(message: IdempotentUnaryResponse, reader: jspb.BinaryReader): IdempotentUnaryResponse;
}

export namespace IdempotentUnaryResponse {
  export type AsObject = {
    payload?: ConformancePayload.AsObject,
  }
}

export class ServerStreamRequest extends jspb.Message {
  getResponseDefinition(): StreamResponseDefinition | undefined;
  setResponseDefinition(value?: StreamResponseDefinition): ServerStreamRequest;
  hasResponseDefinition(): boolean;
  clearResponseDefinition(): ServerStreamRequest;

  getRequestData(): Uint8Array | string;
  getRequestData_asU8(): Uint8Array;
  getRequestData_asB64(): string;
  setRequestData(value: Uint8Array | string): ServerStreamRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ServerStreamRequest.AsObject;
  static toObject(includeInstance: boolean, msg: ServerStreamRequest): ServerStreamRequest.AsObject;
  static serializeBinaryToWriter(message: ServerStreamRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ServerStreamRequest;
  static deserializeBinaryFromReader(message: ServerStreamRequest, reader: jspb.BinaryReader): ServerStreamRequest;
}

export namespace ServerStreamRequest {
  export type AsObject = {
    responseDefinition?: StreamResponseDefinition.AsObject,
    requestData: Uint8Array | string,
  }
}

export class ServerStreamResponse extends jspb.Message {
  getPayload(): ConformancePayload | undefined;
  setPayload(value?: ConformancePayload): ServerStreamResponse;
  hasPayload(): boolean;
  clearPayload(): ServerStreamResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ServerStreamResponse.AsObject;
  static toObject(includeInstance: boolean, msg: ServerStreamResponse): ServerStreamResponse.AsObject;
  static serializeBinaryToWriter(message: ServerStreamResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ServerStreamResponse;
  static deserializeBinaryFromReader(message: ServerStreamResponse, reader: jspb.BinaryReader): ServerStreamResponse;
}

export namespace ServerStreamResponse {
  export type AsObject = {
    payload?: ConformancePayload.AsObject,
  }
}

export class ClientStreamRequest extends jspb.Message {
  getResponseDefinition(): UnaryResponseDefinition | undefined;
  setResponseDefinition(value?: UnaryResponseDefinition): ClientStreamRequest;
  hasResponseDefinition(): boolean;
  clearResponseDefinition(): ClientStreamRequest;

  getRequestData(): Uint8Array | string;
  getRequestData_asU8(): Uint8Array;
  getRequestData_asB64(): string;
  setRequestData(value: Uint8Array | string): ClientStreamRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ClientStreamRequest.AsObject;
  static toObject(includeInstance: boolean, msg: ClientStreamRequest): ClientStreamRequest.AsObject;
  static serializeBinaryToWriter(message: ClientStreamRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ClientStreamRequest;
  static deserializeBinaryFromReader(message: ClientStreamRequest, reader: jspb.BinaryReader): ClientStreamRequest;
}

export namespace ClientStreamRequest {
  export type AsObject = {
    responseDefinition?: UnaryResponseDefinition.AsObject,
    requestData: Uint8Array | string,
  }
}

export class ClientStreamResponse extends jspb.Message {
  getPayload(): ConformancePayload | undefined;
  setPayload(value?: ConformancePayload): ClientStreamResponse;
  hasPayload(): boolean;
  clearPayload(): ClientStreamResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ClientStreamResponse.AsObject;
  static toObject(includeInstance: boolean, msg: ClientStreamResponse): ClientStreamResponse.AsObject;
  static serializeBinaryToWriter(message: ClientStreamResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ClientStreamResponse;
  static deserializeBinaryFromReader(message: ClientStreamResponse, reader: jspb.BinaryReader): ClientStreamResponse;
}

export namespace ClientStreamResponse {
  export type AsObject = {
    payload?: ConformancePayload.AsObject,
  }
}

export class BidiStreamRequest extends jspb.Message {
  getResponseDefinition(): StreamResponseDefinition | undefined;
  setResponseDefinition(value?: StreamResponseDefinition): BidiStreamRequest;
  hasResponseDefinition(): boolean;
  clearResponseDefinition(): BidiStreamRequest;

  getFullDuplex(): boolean;
  setFullDuplex(value: boolean): BidiStreamRequest;

  getRequestData(): Uint8Array | string;
  getRequestData_asU8(): Uint8Array;
  getRequestData_asB64(): string;
  setRequestData(value: Uint8Array | string): BidiStreamRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): BidiStreamRequest.AsObject;
  static toObject(includeInstance: boolean, msg: BidiStreamRequest): BidiStreamRequest.AsObject;
  static serializeBinaryToWriter(message: BidiStreamRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): BidiStreamRequest;
  static deserializeBinaryFromReader(message: BidiStreamRequest, reader: jspb.BinaryReader): BidiStreamRequest;
}

export namespace BidiStreamRequest {
  export type AsObject = {
    responseDefinition?: StreamResponseDefinition.AsObject,
    fullDuplex: boolean,
    requestData: Uint8Array | string,
  }
}

export class BidiStreamResponse extends jspb.Message {
  getPayload(): ConformancePayload | undefined;
  setPayload(value?: ConformancePayload): BidiStreamResponse;
  hasPayload(): boolean;
  clearPayload(): BidiStreamResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): BidiStreamResponse.AsObject;
  static toObject(includeInstance: boolean, msg: BidiStreamResponse): BidiStreamResponse.AsObject;
  static serializeBinaryToWriter(message: BidiStreamResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): BidiStreamResponse;
  static deserializeBinaryFromReader(message: BidiStreamResponse, reader: jspb.BinaryReader): BidiStreamResponse;
}

export namespace BidiStreamResponse {
  export type AsObject = {
    payload?: ConformancePayload.AsObject,
  }
}

export class UnimplementedRequest extends jspb.Message {
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): UnimplementedRequest.AsObject;
  static toObject(includeInstance: boolean, msg: UnimplementedRequest): UnimplementedRequest.AsObject;
  static serializeBinaryToWriter(message: UnimplementedRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): UnimplementedRequest;
  static deserializeBinaryFromReader(message: UnimplementedRequest, reader: jspb.BinaryReader): UnimplementedRequest;
}

export namespace UnimplementedRequest {
  export type AsObject = {
  }
}

export class UnimplementedResponse extends jspb.Message {
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): UnimplementedResponse.AsObject;
  static toObject(includeInstance: boolean, msg: UnimplementedResponse): UnimplementedResponse.AsObject;
  static serializeBinaryToWriter(message: UnimplementedResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): UnimplementedResponse;
  static deserializeBinaryFromReader(message: UnimplementedResponse, reader: jspb.BinaryReader): UnimplementedResponse;
}

export namespace UnimplementedResponse {
  export type AsObject = {
  }
}

export class ConformancePayload extends jspb.Message {
  getData(): Uint8Array | string;
  getData_asU8(): Uint8Array;
  getData_asB64(): string;
  setData(value: Uint8Array | string): ConformancePayload;

  getRequestInfo(): ConformancePayload.RequestInfo | undefined;
  setRequestInfo(value?: ConformancePayload.RequestInfo): ConformancePayload;
  hasRequestInfo(): boolean;
  clearRequestInfo(): ConformancePayload;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ConformancePayload.AsObject;
  static toObject(includeInstance: boolean, msg: ConformancePayload): ConformancePayload.AsObject;
  static serializeBinaryToWriter(message: ConformancePayload, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ConformancePayload;
  static deserializeBinaryFromReader(message: ConformancePayload, reader: jspb.BinaryReader): ConformancePayload;
}

export namespace ConformancePayload {
  export type AsObject = {
    data: Uint8Array | string,
    requestInfo?: ConformancePayload.RequestInfo.AsObject,
  }

  export class RequestInfo extends jspb.Message {
    getRequestHeadersList(): Array<Header>;
    setRequestHeadersList(value: Array<Header>): RequestInfo;
    clearRequestHeadersList(): RequestInfo;
    addRequestHeaders(value?: Header, index?: number): Header;

    getTimeoutMs(): number;
    setTimeoutMs(value: number): RequestInfo;
    hasTimeoutMs(): boolean;
    clearTimeoutMs(): RequestInfo;

    getRequestsList(): Array<google_protobuf_any_pb.Any>;
    setRequestsList(value: Array<google_protobuf_any_pb.Any>): RequestInfo;
    clearRequestsList(): RequestInfo;
    addRequests(value?: google_protobuf_any_pb.Any, index?: number): google_protobuf_any_pb.Any;

    getConnectGetInfo(): ConformancePayload.ConnectGetInfo | undefined;
    setConnectGetInfo(value?: ConformancePayload.ConnectGetInfo): RequestInfo;
    hasConnectGetInfo(): boolean;
    clearConnectGetInfo(): RequestInfo;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): RequestInfo.AsObject;
    static toObject(includeInstance: boolean, msg: RequestInfo): RequestInfo.AsObject;
    static serializeBinaryToWriter(message: RequestInfo, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): RequestInfo;
    static deserializeBinaryFromReader(message: RequestInfo, reader: jspb.BinaryReader): RequestInfo;
  }

  export namespace RequestInfo {
    export type AsObject = {
      requestHeadersList: Array<Header.AsObject>,
      timeoutMs?: number,
      requestsList: Array<google_protobuf_any_pb.Any.AsObject>,
      connectGetInfo?: ConformancePayload.ConnectGetInfo.AsObject,
    }

    export enum TimeoutMsCase { 
      _TIMEOUT_MS_NOT_SET = 0,
      TIMEOUT_MS = 2,
    }
  }


  export class ConnectGetInfo extends jspb.Message {
    getQueryParamsList(): Array<Header>;
    setQueryParamsList(value: Array<Header>): ConnectGetInfo;
    clearQueryParamsList(): ConnectGetInfo;
    addQueryParams(value?: Header, index?: number): Header;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ConnectGetInfo.AsObject;
    static toObject(includeInstance: boolean, msg: ConnectGetInfo): ConnectGetInfo.AsObject;
    static serializeBinaryToWriter(message: ConnectGetInfo, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ConnectGetInfo;
    static deserializeBinaryFromReader(message: ConnectGetInfo, reader: jspb.BinaryReader): ConnectGetInfo;
  }

  export namespace ConnectGetInfo {
    export type AsObject = {
      queryParamsList: Array<Header.AsObject>,
    }
  }

}

export class Error extends jspb.Message {
  getCode(): number;
  setCode(value: number): Error;

  getMessage(): string;
  setMessage(value: string): Error;
  hasMessage(): boolean;
  clearMessage(): Error;

  getDetailsList(): Array<google_protobuf_any_pb.Any>;
  setDetailsList(value: Array<google_protobuf_any_pb.Any>): Error;
  clearDetailsList(): Error;
  addDetails(value?: google_protobuf_any_pb.Any, index?: number): google_protobuf_any_pb.Any;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Error.AsObject;
  static toObject(includeInstance: boolean, msg: Error): Error.AsObject;
  static serializeBinaryToWriter(message: Error, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Error;
  static deserializeBinaryFromReader(message: Error, reader: jspb.BinaryReader): Error;
}

export namespace Error {
  export type AsObject = {
    code: number,
    message?: string,
    detailsList: Array<google_protobuf_any_pb.Any.AsObject>,
  }

  export enum MessageCase { 
    _MESSAGE_NOT_SET = 0,
    MESSAGE = 2,
  }
}

export class Header extends jspb.Message {
  getName(): string;
  setName(value: string): Header;

  getValueList(): Array<string>;
  setValueList(value: Array<string>): Header;
  clearValueList(): Header;
  addValue(value: string, index?: number): Header;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Header.AsObject;
  static toObject(includeInstance: boolean, msg: Header): Header.AsObject;
  static serializeBinaryToWriter(message: Header, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Header;
  static deserializeBinaryFromReader(message: Header, reader: jspb.BinaryReader): Header;
}

export namespace Header {
  export type AsObject = {
    name: string,
    valueList: Array<string>,
  }
}

export class RawHTTPRequest extends jspb.Message {
  getVerb(): string;
  setVerb(value: string): RawHTTPRequest;

  getUri(): string;
  setUri(value: string): RawHTTPRequest;

  getHeadersList(): Array<Header>;
  setHeadersList(value: Array<Header>): RawHTTPRequest;
  clearHeadersList(): RawHTTPRequest;
  addHeaders(value?: Header, index?: number): Header;

  getRawQueryParamsList(): Array<Header>;
  setRawQueryParamsList(value: Array<Header>): RawHTTPRequest;
  clearRawQueryParamsList(): RawHTTPRequest;
  addRawQueryParams(value?: Header, index?: number): Header;

  getEncodedQueryParamsList(): Array<RawHTTPRequest.EncodedQueryParam>;
  setEncodedQueryParamsList(value: Array<RawHTTPRequest.EncodedQueryParam>): RawHTTPRequest;
  clearEncodedQueryParamsList(): RawHTTPRequest;
  addEncodedQueryParams(value?: RawHTTPRequest.EncodedQueryParam, index?: number): RawHTTPRequest.EncodedQueryParam;

  getUnary(): MessageContents | undefined;
  setUnary(value?: MessageContents): RawHTTPRequest;
  hasUnary(): boolean;
  clearUnary(): RawHTTPRequest;

  getStream(): StreamContents | undefined;
  setStream(value?: StreamContents): RawHTTPRequest;
  hasStream(): boolean;
  clearStream(): RawHTTPRequest;

  getBodyCase(): RawHTTPRequest.BodyCase;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): RawHTTPRequest.AsObject;
  static toObject(includeInstance: boolean, msg: RawHTTPRequest): RawHTTPRequest.AsObject;
  static serializeBinaryToWriter(message: RawHTTPRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): RawHTTPRequest;
  static deserializeBinaryFromReader(message: RawHTTPRequest, reader: jspb.BinaryReader): RawHTTPRequest;
}

export namespace RawHTTPRequest {
  export type AsObject = {
    verb: string,
    uri: string,
    headersList: Array<Header.AsObject>,
    rawQueryParamsList: Array<Header.AsObject>,
    encodedQueryParamsList: Array<RawHTTPRequest.EncodedQueryParam.AsObject>,
    unary?: MessageContents.AsObject,
    stream?: StreamContents.AsObject,
  }

  export class EncodedQueryParam extends jspb.Message {
    getName(): string;
    setName(value: string): EncodedQueryParam;

    getValue(): MessageContents | undefined;
    setValue(value?: MessageContents): EncodedQueryParam;
    hasValue(): boolean;
    clearValue(): EncodedQueryParam;

    getBase64Encode(): boolean;
    setBase64Encode(value: boolean): EncodedQueryParam;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): EncodedQueryParam.AsObject;
    static toObject(includeInstance: boolean, msg: EncodedQueryParam): EncodedQueryParam.AsObject;
    static serializeBinaryToWriter(message: EncodedQueryParam, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): EncodedQueryParam;
    static deserializeBinaryFromReader(message: EncodedQueryParam, reader: jspb.BinaryReader): EncodedQueryParam;
  }

  export namespace EncodedQueryParam {
    export type AsObject = {
      name: string,
      value?: MessageContents.AsObject,
      base64Encode: boolean,
    }
  }


  export enum BodyCase { 
    BODY_NOT_SET = 0,
    UNARY = 6,
    STREAM = 7,
  }
}

export class MessageContents extends jspb.Message {
  getBinary(): Uint8Array | string;
  getBinary_asU8(): Uint8Array;
  getBinary_asB64(): string;
  setBinary(value: Uint8Array | string): MessageContents;

  getText(): string;
  setText(value: string): MessageContents;

  getBinaryMessage(): google_protobuf_any_pb.Any | undefined;
  setBinaryMessage(value?: google_protobuf_any_pb.Any): MessageContents;
  hasBinaryMessage(): boolean;
  clearBinaryMessage(): MessageContents;

  getCompression(): connectrpc_conformance_v1_config_pb.Compression;
  setCompression(value: connectrpc_conformance_v1_config_pb.Compression): MessageContents;

  getDataCase(): MessageContents.DataCase;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): MessageContents.AsObject;
  static toObject(includeInstance: boolean, msg: MessageContents): MessageContents.AsObject;
  static serializeBinaryToWriter(message: MessageContents, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): MessageContents;
  static deserializeBinaryFromReader(message: MessageContents, reader: jspb.BinaryReader): MessageContents;
}

export namespace MessageContents {
  export type AsObject = {
    binary: Uint8Array | string,
    text: string,
    binaryMessage?: google_protobuf_any_pb.Any.AsObject,
    compression: connectrpc_conformance_v1_config_pb.Compression,
  }

  export enum DataCase { 
    DATA_NOT_SET = 0,
    BINARY = 1,
    TEXT = 2,
    BINARY_MESSAGE = 3,
  }
}

export class StreamContents extends jspb.Message {
  getItemsList(): Array<StreamContents.StreamItem>;
  setItemsList(value: Array<StreamContents.StreamItem>): StreamContents;
  clearItemsList(): StreamContents;
  addItems(value?: StreamContents.StreamItem, index?: number): StreamContents.StreamItem;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): StreamContents.AsObject;
  static toObject(includeInstance: boolean, msg: StreamContents): StreamContents.AsObject;
  static serializeBinaryToWriter(message: StreamContents, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): StreamContents;
  static deserializeBinaryFromReader(message: StreamContents, reader: jspb.BinaryReader): StreamContents;
}

export namespace StreamContents {
  export type AsObject = {
    itemsList: Array<StreamContents.StreamItem.AsObject>,
  }

  export class StreamItem extends jspb.Message {
    getFlags(): number;
    setFlags(value: number): StreamItem;

    getLength(): number;
    setLength(value: number): StreamItem;
    hasLength(): boolean;
    clearLength(): StreamItem;

    getPayload(): MessageContents | undefined;
    setPayload(value?: MessageContents): StreamItem;
    hasPayload(): boolean;
    clearPayload(): StreamItem;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): StreamItem.AsObject;
    static toObject(includeInstance: boolean, msg: StreamItem): StreamItem.AsObject;
    static serializeBinaryToWriter(message: StreamItem, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): StreamItem;
    static deserializeBinaryFromReader(message: StreamItem, reader: jspb.BinaryReader): StreamItem;
  }

  export namespace StreamItem {
    export type AsObject = {
      flags: number,
      length?: number,
      payload?: MessageContents.AsObject,
    }

    export enum LengthCase { 
      _LENGTH_NOT_SET = 0,
      LENGTH = 2,
    }
  }

}

export class RawHTTPResponse extends jspb.Message {
  getStatusCode(): number;
  setStatusCode(value: number): RawHTTPResponse;

  getHeadersList(): Array<Header>;
  setHeadersList(value: Array<Header>): RawHTTPResponse;
  clearHeadersList(): RawHTTPResponse;
  addHeaders(value?: Header, index?: number): Header;

  getUnary(): MessageContents | undefined;
  setUnary(value?: MessageContents): RawHTTPResponse;
  hasUnary(): boolean;
  clearUnary(): RawHTTPResponse;

  getStream(): StreamContents | undefined;
  setStream(value?: StreamContents): RawHTTPResponse;
  hasStream(): boolean;
  clearStream(): RawHTTPResponse;

  getTrailersList(): Array<Header>;
  setTrailersList(value: Array<Header>): RawHTTPResponse;
  clearTrailersList(): RawHTTPResponse;
  addTrailers(value?: Header, index?: number): Header;

  getBodyCase(): RawHTTPResponse.BodyCase;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): RawHTTPResponse.AsObject;
  static toObject(includeInstance: boolean, msg: RawHTTPResponse): RawHTTPResponse.AsObject;
  static serializeBinaryToWriter(message: RawHTTPResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): RawHTTPResponse;
  static deserializeBinaryFromReader(message: RawHTTPResponse, reader: jspb.BinaryReader): RawHTTPResponse;
}

export namespace RawHTTPResponse {
  export type AsObject = {
    statusCode: number,
    headersList: Array<Header.AsObject>,
    unary?: MessageContents.AsObject,
    stream?: StreamContents.AsObject,
    trailersList: Array<Header.AsObject>,
  }

  export enum BodyCase { 
    BODY_NOT_SET = 0,
    UNARY = 3,
    STREAM = 4,
  }
}

