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

package grpcutil

import (
	"errors"

	"google.golang.org/grpc/encoding"
	"google.golang.org/protobuf/encoding/protojson"
	proto "google.golang.org/protobuf/proto"
)

func init() { //nolint:gochecknoinits // gRPC requires this to register a codec.
	encoding.RegisterCodec(&jsonCodec{})
}

// jsonCodec is a codec for use with a gRPC server.
type jsonCodec struct{}

// Name returns the name of the json codec for use with gRPC.
func (j *jsonCodec) Name() string {
	return "json"
}

// Marshal marshals a given message. If the given parameter is not a proto.Message,
// function returns an error.
func (j *jsonCodec) Marshal(v any) (out []byte, err error) {
	pm, ok := v.(proto.Message)
	if !ok {
		return nil, errors.New("message is not a proto message and cannot be marshalled ")
	}
	return protojson.Marshal(pm)
}

// Marshal unmarshals a given message. If the given parameter is not a proto.Message,
// function returns an error.
func (j *jsonCodec) Unmarshal(data []byte, v interface{}) (err error) {
	pm, ok := v.(proto.Message)
	if !ok {
		return errors.New("message is not a proto message and cannot be unmarshalled")
	}
	return protojson.Unmarshal(data, pm)
}
