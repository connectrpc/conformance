import * as jspb from 'google-protobuf'

import * as google_protobuf_any_pb from 'google-protobuf/google/protobuf/any_pb';
import * as google_protobuf_wrappers_pb from 'google-protobuf/google/protobuf/wrappers_pb';


export class Payload extends jspb.Message {
  getType(): PayloadType;
  setType(value: PayloadType): Payload;

  getBody(): Uint8Array | string;
  getBody_asU8(): Uint8Array;
  getBody_asB64(): string;
  setBody(value: Uint8Array | string): Payload;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Payload.AsObject;
  static toObject(includeInstance: boolean, msg: Payload): Payload.AsObject;
  static serializeBinaryToWriter(message: Payload, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Payload;
  static deserializeBinaryFromReader(message: Payload, reader: jspb.BinaryReader): Payload;
}

export namespace Payload {
  export type AsObject = {
    type: PayloadType,
    body: Uint8Array | string,
  }
}

export class EchoStatus extends jspb.Message {
  getCode(): number;
  setCode(value: number): EchoStatus;

  getMessage(): string;
  setMessage(value: string): EchoStatus;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): EchoStatus.AsObject;
  static toObject(includeInstance: boolean, msg: EchoStatus): EchoStatus.AsObject;
  static serializeBinaryToWriter(message: EchoStatus, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): EchoStatus;
  static deserializeBinaryFromReader(message: EchoStatus, reader: jspb.BinaryReader): EchoStatus;
}

export namespace EchoStatus {
  export type AsObject = {
    code: number,
    message: string,
  }
}

export class SimpleRequest extends jspb.Message {
  getResponseType(): PayloadType;
  setResponseType(value: PayloadType): SimpleRequest;

  getResponseSize(): number;
  setResponseSize(value: number): SimpleRequest;

  getPayload(): Payload | undefined;
  setPayload(value?: Payload): SimpleRequest;
  hasPayload(): boolean;
  clearPayload(): SimpleRequest;

  getFillUsername(): boolean;
  setFillUsername(value: boolean): SimpleRequest;

  getFillOauthScope(): boolean;
  setFillOauthScope(value: boolean): SimpleRequest;

  getResponseCompressed(): google_protobuf_wrappers_pb.BoolValue | undefined;
  setResponseCompressed(value?: google_protobuf_wrappers_pb.BoolValue): SimpleRequest;
  hasResponseCompressed(): boolean;
  clearResponseCompressed(): SimpleRequest;

  getResponseStatus(): EchoStatus | undefined;
  setResponseStatus(value?: EchoStatus): SimpleRequest;
  hasResponseStatus(): boolean;
  clearResponseStatus(): SimpleRequest;

  getExpectCompressed(): google_protobuf_wrappers_pb.BoolValue | undefined;
  setExpectCompressed(value?: google_protobuf_wrappers_pb.BoolValue): SimpleRequest;
  hasExpectCompressed(): boolean;
  clearExpectCompressed(): SimpleRequest;

  getFillServerId(): boolean;
  setFillServerId(value: boolean): SimpleRequest;

  getFillGrpclbRouteType(): boolean;
  setFillGrpclbRouteType(value: boolean): SimpleRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): SimpleRequest.AsObject;
  static toObject(includeInstance: boolean, msg: SimpleRequest): SimpleRequest.AsObject;
  static serializeBinaryToWriter(message: SimpleRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): SimpleRequest;
  static deserializeBinaryFromReader(message: SimpleRequest, reader: jspb.BinaryReader): SimpleRequest;
}

export namespace SimpleRequest {
  export type AsObject = {
    responseType: PayloadType,
    responseSize: number,
    payload?: Payload.AsObject,
    fillUsername: boolean,
    fillOauthScope: boolean,
    responseCompressed?: google_protobuf_wrappers_pb.BoolValue.AsObject,
    responseStatus?: EchoStatus.AsObject,
    expectCompressed?: google_protobuf_wrappers_pb.BoolValue.AsObject,
    fillServerId: boolean,
    fillGrpclbRouteType: boolean,
  }
}

export class SimpleResponse extends jspb.Message {
  getPayload(): Payload | undefined;
  setPayload(value?: Payload): SimpleResponse;
  hasPayload(): boolean;
  clearPayload(): SimpleResponse;

  getUsername(): string;
  setUsername(value: string): SimpleResponse;

  getOauthScope(): string;
  setOauthScope(value: string): SimpleResponse;

  getServerId(): string;
  setServerId(value: string): SimpleResponse;

  getGrpclbRouteType(): GrpclbRouteType;
  setGrpclbRouteType(value: GrpclbRouteType): SimpleResponse;

  getHostname(): string;
  setHostname(value: string): SimpleResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): SimpleResponse.AsObject;
  static toObject(includeInstance: boolean, msg: SimpleResponse): SimpleResponse.AsObject;
  static serializeBinaryToWriter(message: SimpleResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): SimpleResponse;
  static deserializeBinaryFromReader(message: SimpleResponse, reader: jspb.BinaryReader): SimpleResponse;
}

export namespace SimpleResponse {
  export type AsObject = {
    payload?: Payload.AsObject,
    username: string,
    oauthScope: string,
    serverId: string,
    grpclbRouteType: GrpclbRouteType,
    hostname: string,
  }
}

export class StreamingInputCallRequest extends jspb.Message {
  getPayload(): Payload | undefined;
  setPayload(value?: Payload): StreamingInputCallRequest;
  hasPayload(): boolean;
  clearPayload(): StreamingInputCallRequest;

  getExpectCompressed(): google_protobuf_wrappers_pb.BoolValue | undefined;
  setExpectCompressed(value?: google_protobuf_wrappers_pb.BoolValue): StreamingInputCallRequest;
  hasExpectCompressed(): boolean;
  clearExpectCompressed(): StreamingInputCallRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): StreamingInputCallRequest.AsObject;
  static toObject(includeInstance: boolean, msg: StreamingInputCallRequest): StreamingInputCallRequest.AsObject;
  static serializeBinaryToWriter(message: StreamingInputCallRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): StreamingInputCallRequest;
  static deserializeBinaryFromReader(message: StreamingInputCallRequest, reader: jspb.BinaryReader): StreamingInputCallRequest;
}

export namespace StreamingInputCallRequest {
  export type AsObject = {
    payload?: Payload.AsObject,
    expectCompressed?: google_protobuf_wrappers_pb.BoolValue.AsObject,
  }
}

export class StreamingInputCallResponse extends jspb.Message {
  getAggregatedPayloadSize(): number;
  setAggregatedPayloadSize(value: number): StreamingInputCallResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): StreamingInputCallResponse.AsObject;
  static toObject(includeInstance: boolean, msg: StreamingInputCallResponse): StreamingInputCallResponse.AsObject;
  static serializeBinaryToWriter(message: StreamingInputCallResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): StreamingInputCallResponse;
  static deserializeBinaryFromReader(message: StreamingInputCallResponse, reader: jspb.BinaryReader): StreamingInputCallResponse;
}

export namespace StreamingInputCallResponse {
  export type AsObject = {
    aggregatedPayloadSize: number,
  }
}

export class ResponseParameters extends jspb.Message {
  getSize(): number;
  setSize(value: number): ResponseParameters;

  getIntervalUs(): number;
  setIntervalUs(value: number): ResponseParameters;

  getCompressed(): google_protobuf_wrappers_pb.BoolValue | undefined;
  setCompressed(value?: google_protobuf_wrappers_pb.BoolValue): ResponseParameters;
  hasCompressed(): boolean;
  clearCompressed(): ResponseParameters;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ResponseParameters.AsObject;
  static toObject(includeInstance: boolean, msg: ResponseParameters): ResponseParameters.AsObject;
  static serializeBinaryToWriter(message: ResponseParameters, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ResponseParameters;
  static deserializeBinaryFromReader(message: ResponseParameters, reader: jspb.BinaryReader): ResponseParameters;
}

export namespace ResponseParameters {
  export type AsObject = {
    size: number,
    intervalUs: number,
    compressed?: google_protobuf_wrappers_pb.BoolValue.AsObject,
  }
}

export class StreamingOutputCallRequest extends jspb.Message {
  getResponseType(): PayloadType;
  setResponseType(value: PayloadType): StreamingOutputCallRequest;

  getResponseParametersList(): Array<ResponseParameters>;
  setResponseParametersList(value: Array<ResponseParameters>): StreamingOutputCallRequest;
  clearResponseParametersList(): StreamingOutputCallRequest;
  addResponseParameters(value?: ResponseParameters, index?: number): ResponseParameters;

  getPayload(): Payload | undefined;
  setPayload(value?: Payload): StreamingOutputCallRequest;
  hasPayload(): boolean;
  clearPayload(): StreamingOutputCallRequest;

  getResponseStatus(): EchoStatus | undefined;
  setResponseStatus(value?: EchoStatus): StreamingOutputCallRequest;
  hasResponseStatus(): boolean;
  clearResponseStatus(): StreamingOutputCallRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): StreamingOutputCallRequest.AsObject;
  static toObject(includeInstance: boolean, msg: StreamingOutputCallRequest): StreamingOutputCallRequest.AsObject;
  static serializeBinaryToWriter(message: StreamingOutputCallRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): StreamingOutputCallRequest;
  static deserializeBinaryFromReader(message: StreamingOutputCallRequest, reader: jspb.BinaryReader): StreamingOutputCallRequest;
}

export namespace StreamingOutputCallRequest {
  export type AsObject = {
    responseType: PayloadType,
    responseParametersList: Array<ResponseParameters.AsObject>,
    payload?: Payload.AsObject,
    responseStatus?: EchoStatus.AsObject,
  }
}

export class StreamingOutputCallResponse extends jspb.Message {
  getPayload(): Payload | undefined;
  setPayload(value?: Payload): StreamingOutputCallResponse;
  hasPayload(): boolean;
  clearPayload(): StreamingOutputCallResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): StreamingOutputCallResponse.AsObject;
  static toObject(includeInstance: boolean, msg: StreamingOutputCallResponse): StreamingOutputCallResponse.AsObject;
  static serializeBinaryToWriter(message: StreamingOutputCallResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): StreamingOutputCallResponse;
  static deserializeBinaryFromReader(message: StreamingOutputCallResponse, reader: jspb.BinaryReader): StreamingOutputCallResponse;
}

export namespace StreamingOutputCallResponse {
  export type AsObject = {
    payload?: Payload.AsObject,
  }
}

export class ReconnectParams extends jspb.Message {
  getMaxReconnectBackoffMs(): number;
  setMaxReconnectBackoffMs(value: number): ReconnectParams;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ReconnectParams.AsObject;
  static toObject(includeInstance: boolean, msg: ReconnectParams): ReconnectParams.AsObject;
  static serializeBinaryToWriter(message: ReconnectParams, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ReconnectParams;
  static deserializeBinaryFromReader(message: ReconnectParams, reader: jspb.BinaryReader): ReconnectParams;
}

export namespace ReconnectParams {
  export type AsObject = {
    maxReconnectBackoffMs: number,
  }
}

export class ReconnectInfo extends jspb.Message {
  getPassed(): boolean;
  setPassed(value: boolean): ReconnectInfo;

  getBackoffMsList(): Array<number>;
  setBackoffMsList(value: Array<number>): ReconnectInfo;
  clearBackoffMsList(): ReconnectInfo;
  addBackoffMs(value: number, index?: number): ReconnectInfo;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ReconnectInfo.AsObject;
  static toObject(includeInstance: boolean, msg: ReconnectInfo): ReconnectInfo.AsObject;
  static serializeBinaryToWriter(message: ReconnectInfo, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ReconnectInfo;
  static deserializeBinaryFromReader(message: ReconnectInfo, reader: jspb.BinaryReader): ReconnectInfo;
}

export namespace ReconnectInfo {
  export type AsObject = {
    passed: boolean,
    backoffMsList: Array<number>,
  }
}

export class LoadBalancerStatsRequest extends jspb.Message {
  getNumRpcs(): number;
  setNumRpcs(value: number): LoadBalancerStatsRequest;

  getTimeoutSec(): number;
  setTimeoutSec(value: number): LoadBalancerStatsRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): LoadBalancerStatsRequest.AsObject;
  static toObject(includeInstance: boolean, msg: LoadBalancerStatsRequest): LoadBalancerStatsRequest.AsObject;
  static serializeBinaryToWriter(message: LoadBalancerStatsRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): LoadBalancerStatsRequest;
  static deserializeBinaryFromReader(message: LoadBalancerStatsRequest, reader: jspb.BinaryReader): LoadBalancerStatsRequest;
}

export namespace LoadBalancerStatsRequest {
  export type AsObject = {
    numRpcs: number,
    timeoutSec: number,
  }
}

export class LoadBalancerStatsResponse extends jspb.Message {
  getRpcsByPeerMap(): jspb.Map<string, number>;
  clearRpcsByPeerMap(): LoadBalancerStatsResponse;

  getNumFailures(): number;
  setNumFailures(value: number): LoadBalancerStatsResponse;

  getRpcsByMethodMap(): jspb.Map<string, LoadBalancerStatsResponse.RpcsByPeer>;
  clearRpcsByMethodMap(): LoadBalancerStatsResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): LoadBalancerStatsResponse.AsObject;
  static toObject(includeInstance: boolean, msg: LoadBalancerStatsResponse): LoadBalancerStatsResponse.AsObject;
  static serializeBinaryToWriter(message: LoadBalancerStatsResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): LoadBalancerStatsResponse;
  static deserializeBinaryFromReader(message: LoadBalancerStatsResponse, reader: jspb.BinaryReader): LoadBalancerStatsResponse;
}

export namespace LoadBalancerStatsResponse {
  export type AsObject = {
    rpcsByPeerMap: Array<[string, number]>,
    numFailures: number,
    rpcsByMethodMap: Array<[string, LoadBalancerStatsResponse.RpcsByPeer.AsObject]>,
  }

  export class RpcsByPeer extends jspb.Message {
    getRpcsByPeerMap(): jspb.Map<string, number>;
    clearRpcsByPeerMap(): RpcsByPeer;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): RpcsByPeer.AsObject;
    static toObject(includeInstance: boolean, msg: RpcsByPeer): RpcsByPeer.AsObject;
    static serializeBinaryToWriter(message: RpcsByPeer, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): RpcsByPeer;
    static deserializeBinaryFromReader(message: RpcsByPeer, reader: jspb.BinaryReader): RpcsByPeer;
  }

  export namespace RpcsByPeer {
    export type AsObject = {
      rpcsByPeerMap: Array<[string, number]>,
    }
  }

}

export class LoadBalancerAccumulatedStatsRequest extends jspb.Message {
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): LoadBalancerAccumulatedStatsRequest.AsObject;
  static toObject(includeInstance: boolean, msg: LoadBalancerAccumulatedStatsRequest): LoadBalancerAccumulatedStatsRequest.AsObject;
  static serializeBinaryToWriter(message: LoadBalancerAccumulatedStatsRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): LoadBalancerAccumulatedStatsRequest;
  static deserializeBinaryFromReader(message: LoadBalancerAccumulatedStatsRequest, reader: jspb.BinaryReader): LoadBalancerAccumulatedStatsRequest;
}

export namespace LoadBalancerAccumulatedStatsRequest {
  export type AsObject = {
  }
}

export class LoadBalancerAccumulatedStatsResponse extends jspb.Message {
  getNumRpcsStartedByMethodMap(): jspb.Map<string, number>;
  clearNumRpcsStartedByMethodMap(): LoadBalancerAccumulatedStatsResponse;

  getNumRpcsSucceededByMethodMap(): jspb.Map<string, number>;
  clearNumRpcsSucceededByMethodMap(): LoadBalancerAccumulatedStatsResponse;

  getNumRpcsFailedByMethodMap(): jspb.Map<string, number>;
  clearNumRpcsFailedByMethodMap(): LoadBalancerAccumulatedStatsResponse;

  getStatsPerMethodMap(): jspb.Map<string, LoadBalancerAccumulatedStatsResponse.MethodStats>;
  clearStatsPerMethodMap(): LoadBalancerAccumulatedStatsResponse;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): LoadBalancerAccumulatedStatsResponse.AsObject;
  static toObject(includeInstance: boolean, msg: LoadBalancerAccumulatedStatsResponse): LoadBalancerAccumulatedStatsResponse.AsObject;
  static serializeBinaryToWriter(message: LoadBalancerAccumulatedStatsResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): LoadBalancerAccumulatedStatsResponse;
  static deserializeBinaryFromReader(message: LoadBalancerAccumulatedStatsResponse, reader: jspb.BinaryReader): LoadBalancerAccumulatedStatsResponse;
}

export namespace LoadBalancerAccumulatedStatsResponse {
  export type AsObject = {
    numRpcsStartedByMethodMap: Array<[string, number]>,
    numRpcsSucceededByMethodMap: Array<[string, number]>,
    numRpcsFailedByMethodMap: Array<[string, number]>,
    statsPerMethodMap: Array<[string, LoadBalancerAccumulatedStatsResponse.MethodStats.AsObject]>,
  }

  export class MethodStats extends jspb.Message {
    getRpcsStarted(): number;
    setRpcsStarted(value: number): MethodStats;

    getResultMap(): jspb.Map<number, number>;
    clearResultMap(): MethodStats;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): MethodStats.AsObject;
    static toObject(includeInstance: boolean, msg: MethodStats): MethodStats.AsObject;
    static serializeBinaryToWriter(message: MethodStats, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): MethodStats;
    static deserializeBinaryFromReader(message: MethodStats, reader: jspb.BinaryReader): MethodStats;
  }

  export namespace MethodStats {
    export type AsObject = {
      rpcsStarted: number,
      resultMap: Array<[number, number]>,
    }
  }

}

export class ClientConfigureRequest extends jspb.Message {
  getTypesList(): Array<ClientConfigureRequest.RpcType>;
  setTypesList(value: Array<ClientConfigureRequest.RpcType>): ClientConfigureRequest;
  clearTypesList(): ClientConfigureRequest;
  addTypes(value: ClientConfigureRequest.RpcType, index?: number): ClientConfigureRequest;

  getMetadataList(): Array<ClientConfigureRequest.Metadata>;
  setMetadataList(value: Array<ClientConfigureRequest.Metadata>): ClientConfigureRequest;
  clearMetadataList(): ClientConfigureRequest;
  addMetadata(value?: ClientConfigureRequest.Metadata, index?: number): ClientConfigureRequest.Metadata;

  getTimeoutSec(): number;
  setTimeoutSec(value: number): ClientConfigureRequest;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ClientConfigureRequest.AsObject;
  static toObject(includeInstance: boolean, msg: ClientConfigureRequest): ClientConfigureRequest.AsObject;
  static serializeBinaryToWriter(message: ClientConfigureRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ClientConfigureRequest;
  static deserializeBinaryFromReader(message: ClientConfigureRequest, reader: jspb.BinaryReader): ClientConfigureRequest;
}

export namespace ClientConfigureRequest {
  export type AsObject = {
    typesList: Array<ClientConfigureRequest.RpcType>,
    metadataList: Array<ClientConfigureRequest.Metadata.AsObject>,
    timeoutSec: number,
  }

  export class Metadata extends jspb.Message {
    getType(): ClientConfigureRequest.RpcType;
    setType(value: ClientConfigureRequest.RpcType): Metadata;

    getKey(): string;
    setKey(value: string): Metadata;

    getValue(): string;
    setValue(value: string): Metadata;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Metadata.AsObject;
    static toObject(includeInstance: boolean, msg: Metadata): Metadata.AsObject;
    static serializeBinaryToWriter(message: Metadata, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Metadata;
    static deserializeBinaryFromReader(message: Metadata, reader: jspb.BinaryReader): Metadata;
  }

  export namespace Metadata {
    export type AsObject = {
      type: ClientConfigureRequest.RpcType,
      key: string,
      value: string,
    }
  }


  export enum RpcType { 
    EMPTY_CALL = 0,
    UNARY_CALL = 1,
  }
}

export class ClientConfigureResponse extends jspb.Message {
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ClientConfigureResponse.AsObject;
  static toObject(includeInstance: boolean, msg: ClientConfigureResponse): ClientConfigureResponse.AsObject;
  static serializeBinaryToWriter(message: ClientConfigureResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ClientConfigureResponse;
  static deserializeBinaryFromReader(message: ClientConfigureResponse, reader: jspb.BinaryReader): ClientConfigureResponse;
}

export namespace ClientConfigureResponse {
  export type AsObject = {
  }
}

export class ErrorDetail extends jspb.Message {
  getReason(): string;
  setReason(value: string): ErrorDetail;

  getDomain(): string;
  setDomain(value: string): ErrorDetail;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ErrorDetail.AsObject;
  static toObject(includeInstance: boolean, msg: ErrorDetail): ErrorDetail.AsObject;
  static serializeBinaryToWriter(message: ErrorDetail, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ErrorDetail;
  static deserializeBinaryFromReader(message: ErrorDetail, reader: jspb.BinaryReader): ErrorDetail;
}

export namespace ErrorDetail {
  export type AsObject = {
    reason: string,
    domain: string,
  }
}

export class ErrorStatus extends jspb.Message {
  getCode(): number;
  setCode(value: number): ErrorStatus;

  getMessage(): string;
  setMessage(value: string): ErrorStatus;

  getDetailsList(): Array<google_protobuf_any_pb.Any>;
  setDetailsList(value: Array<google_protobuf_any_pb.Any>): ErrorStatus;
  clearDetailsList(): ErrorStatus;
  addDetails(value?: google_protobuf_any_pb.Any, index?: number): google_protobuf_any_pb.Any;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ErrorStatus.AsObject;
  static toObject(includeInstance: boolean, msg: ErrorStatus): ErrorStatus.AsObject;
  static serializeBinaryToWriter(message: ErrorStatus, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ErrorStatus;
  static deserializeBinaryFromReader(message: ErrorStatus, reader: jspb.BinaryReader): ErrorStatus;
}

export namespace ErrorStatus {
  export type AsObject = {
    code: number,
    message: string,
    detailsList: Array<google_protobuf_any_pb.Any.AsObject>,
  }
}

export enum PayloadType { 
  COMPRESSABLE = 0,
}
export enum GrpclbRouteType { 
  GRPCLB_ROUTE_TYPE_UNKNOWN = 0,
  GRPCLB_ROUTE_TYPE_FALLBACK = 1,
  GRPCLB_ROUTE_TYPE_BACKEND = 2,
}
