package app

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
