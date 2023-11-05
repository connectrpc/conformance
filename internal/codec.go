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

package internal

import (
	"google.golang.org/protobuf/encoding/protojson"
	proto "google.golang.org/protobuf/proto"
)

// codec describes anything that can marshal and unmarshal proto messages.
type codec interface {
	Marshal(msg proto.Message) ([]byte, error)
	Unmarshal(b []byte, msg proto.Message) error
}

// NewCodec returns a new Codec.
func NewCodec(json bool) codec {
	if json {
		return &jsonCodec{}
	}
	return &protoCodec{}
}

// jsonCodec marshals and unmarshals the JSON format.
type jsonCodec struct {
	protojson.MarshalOptions
	protojson.UnmarshalOptions
}

// protoCodec marshals and unmarshals the Protobuf binary format.
type protoCodec struct {
	proto.MarshalOptions
	proto.UnmarshalOptions
}
