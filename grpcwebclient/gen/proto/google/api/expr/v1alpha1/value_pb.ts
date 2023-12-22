// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// @generated by protoc-gen-es v1.6.0 with parameter "target=ts"
// @generated from file google/api/expr/v1alpha1/value.proto (package google.api.expr.v1alpha1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import type { BinaryReadOptions, FieldList, JsonReadOptions, JsonValue, PartialMessage, PlainMessage } from "@bufbuild/protobuf";
import { Any, Message, NullValue, proto3 } from "@bufbuild/protobuf";

/**
 * Represents a CEL value.
 *
 * This is similar to `google.protobuf.Value`, but can represent CEL's full
 * range of values.
 *
 * @generated from message google.api.expr.v1alpha1.Value
 */
export class Value extends Message<Value> {
  /**
   * Required. The valid kinds of values.
   *
   * @generated from oneof google.api.expr.v1alpha1.Value.kind
   */
  kind: {
    /**
     * Null value.
     *
     * @generated from field: google.protobuf.NullValue null_value = 1;
     */
    value: NullValue;
    case: "nullValue";
  } | {
    /**
     * Boolean value.
     *
     * @generated from field: bool bool_value = 2;
     */
    value: boolean;
    case: "boolValue";
  } | {
    /**
     * Signed integer value.
     *
     * @generated from field: int64 int64_value = 3;
     */
    value: bigint;
    case: "int64Value";
  } | {
    /**
     * Unsigned integer value.
     *
     * @generated from field: uint64 uint64_value = 4;
     */
    value: bigint;
    case: "uint64Value";
  } | {
    /**
     * Floating point value.
     *
     * @generated from field: double double_value = 5;
     */
    value: number;
    case: "doubleValue";
  } | {
    /**
     * UTF-8 string value.
     *
     * @generated from field: string string_value = 6;
     */
    value: string;
    case: "stringValue";
  } | {
    /**
     * Byte string value.
     *
     * @generated from field: bytes bytes_value = 7;
     */
    value: Uint8Array;
    case: "bytesValue";
  } | {
    /**
     * An enum value.
     *
     * @generated from field: google.api.expr.v1alpha1.EnumValue enum_value = 9;
     */
    value: EnumValue;
    case: "enumValue";
  } | {
    /**
     * The proto message backing an object value.
     *
     * @generated from field: google.protobuf.Any object_value = 10;
     */
    value: Any;
    case: "objectValue";
  } | {
    /**
     * Map value.
     *
     * @generated from field: google.api.expr.v1alpha1.MapValue map_value = 11;
     */
    value: MapValue;
    case: "mapValue";
  } | {
    /**
     * List value.
     *
     * @generated from field: google.api.expr.v1alpha1.ListValue list_value = 12;
     */
    value: ListValue;
    case: "listValue";
  } | {
    /**
     * Type value.
     *
     * @generated from field: string type_value = 15;
     */
    value: string;
    case: "typeValue";
  } | { case: undefined; value?: undefined } = { case: undefined };

  constructor(data?: PartialMessage<Value>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "google.api.expr.v1alpha1.Value";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "null_value", kind: "enum", T: proto3.getEnumType(NullValue), oneof: "kind" },
    { no: 2, name: "bool_value", kind: "scalar", T: 8 /* ScalarType.BOOL */, oneof: "kind" },
    { no: 3, name: "int64_value", kind: "scalar", T: 3 /* ScalarType.INT64 */, oneof: "kind" },
    { no: 4, name: "uint64_value", kind: "scalar", T: 4 /* ScalarType.UINT64 */, oneof: "kind" },
    { no: 5, name: "double_value", kind: "scalar", T: 1 /* ScalarType.DOUBLE */, oneof: "kind" },
    { no: 6, name: "string_value", kind: "scalar", T: 9 /* ScalarType.STRING */, oneof: "kind" },
    { no: 7, name: "bytes_value", kind: "scalar", T: 12 /* ScalarType.BYTES */, oneof: "kind" },
    { no: 9, name: "enum_value", kind: "message", T: EnumValue, oneof: "kind" },
    { no: 10, name: "object_value", kind: "message", T: Any, oneof: "kind" },
    { no: 11, name: "map_value", kind: "message", T: MapValue, oneof: "kind" },
    { no: 12, name: "list_value", kind: "message", T: ListValue, oneof: "kind" },
    { no: 15, name: "type_value", kind: "scalar", T: 9 /* ScalarType.STRING */, oneof: "kind" },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): Value {
    return new Value().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): Value {
    return new Value().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): Value {
    return new Value().fromJsonString(jsonString, options);
  }

  static equals(a: Value | PlainMessage<Value> | undefined, b: Value | PlainMessage<Value> | undefined): boolean {
    return proto3.util.equals(Value, a, b);
  }
}

/**
 * An enum value.
 *
 * @generated from message google.api.expr.v1alpha1.EnumValue
 */
export class EnumValue extends Message<EnumValue> {
  /**
   * The fully qualified name of the enum type.
   *
   * @generated from field: string type = 1;
   */
  type = "";

  /**
   * The value of the enum.
   *
   * @generated from field: int32 value = 2;
   */
  value = 0;

  constructor(data?: PartialMessage<EnumValue>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "google.api.expr.v1alpha1.EnumValue";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "type", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "value", kind: "scalar", T: 5 /* ScalarType.INT32 */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): EnumValue {
    return new EnumValue().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): EnumValue {
    return new EnumValue().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): EnumValue {
    return new EnumValue().fromJsonString(jsonString, options);
  }

  static equals(a: EnumValue | PlainMessage<EnumValue> | undefined, b: EnumValue | PlainMessage<EnumValue> | undefined): boolean {
    return proto3.util.equals(EnumValue, a, b);
  }
}

/**
 * A list.
 *
 * Wrapped in a message so 'not set' and empty can be differentiated, which is
 * required for use in a 'oneof'.
 *
 * @generated from message google.api.expr.v1alpha1.ListValue
 */
export class ListValue extends Message<ListValue> {
  /**
   * The ordered values in the list.
   *
   * @generated from field: repeated google.api.expr.v1alpha1.Value values = 1;
   */
  values: Value[] = [];

  constructor(data?: PartialMessage<ListValue>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "google.api.expr.v1alpha1.ListValue";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "values", kind: "message", T: Value, repeated: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ListValue {
    return new ListValue().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ListValue {
    return new ListValue().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ListValue {
    return new ListValue().fromJsonString(jsonString, options);
  }

  static equals(a: ListValue | PlainMessage<ListValue> | undefined, b: ListValue | PlainMessage<ListValue> | undefined): boolean {
    return proto3.util.equals(ListValue, a, b);
  }
}

/**
 * A map.
 *
 * Wrapped in a message so 'not set' and empty can be differentiated, which is
 * required for use in a 'oneof'.
 *
 * @generated from message google.api.expr.v1alpha1.MapValue
 */
export class MapValue extends Message<MapValue> {
  /**
   * The set of map entries.
   *
   * CEL has fewer restrictions on keys, so a protobuf map represenation
   * cannot be used.
   *
   * @generated from field: repeated google.api.expr.v1alpha1.MapValue.Entry entries = 1;
   */
  entries: MapValue_Entry[] = [];

  constructor(data?: PartialMessage<MapValue>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "google.api.expr.v1alpha1.MapValue";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "entries", kind: "message", T: MapValue_Entry, repeated: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): MapValue {
    return new MapValue().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): MapValue {
    return new MapValue().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): MapValue {
    return new MapValue().fromJsonString(jsonString, options);
  }

  static equals(a: MapValue | PlainMessage<MapValue> | undefined, b: MapValue | PlainMessage<MapValue> | undefined): boolean {
    return proto3.util.equals(MapValue, a, b);
  }
}

/**
 * An entry in the map.
 *
 * @generated from message google.api.expr.v1alpha1.MapValue.Entry
 */
export class MapValue_Entry extends Message<MapValue_Entry> {
  /**
   * The key.
   *
   * Must be unique with in the map.
   * Currently only boolean, int, uint, and string values can be keys.
   *
   * @generated from field: google.api.expr.v1alpha1.Value key = 1;
   */
  key?: Value;

  /**
   * The value.
   *
   * @generated from field: google.api.expr.v1alpha1.Value value = 2;
   */
  value?: Value;

  constructor(data?: PartialMessage<MapValue_Entry>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "google.api.expr.v1alpha1.MapValue.Entry";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "key", kind: "message", T: Value },
    { no: 2, name: "value", kind: "message", T: Value },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): MapValue_Entry {
    return new MapValue_Entry().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): MapValue_Entry {
    return new MapValue_Entry().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): MapValue_Entry {
    return new MapValue_Entry().fromJsonString(jsonString, options);
  }

  static equals(a: MapValue_Entry | PlainMessage<MapValue_Entry> | undefined, b: MapValue_Entry | PlainMessage<MapValue_Entry> | undefined): boolean {
    return proto3.util.equals(MapValue_Entry, a, b);
  }
}
