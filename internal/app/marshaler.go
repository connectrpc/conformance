package app

import (
	"google.golang.org/protobuf/encoding/protojson"
	proto "google.golang.org/protobuf/proto"
)

// Marshaler describes anything that can marshal and unmarshal proto messages.
type Marshaler interface {
	Marshal(msg proto.Message) ([]byte, error)
	Unmarshal(b []byte, msg proto.Message) error
}

// NewMarshaler returns a new marshaler.
func NewMarshaler(json bool) Marshaler {
	if json {
		return &jsonMarshaler{}
	}
	return &protoMarshaler{}
}

// jsonMarshaler marshals and unmarshals the JSON format.
type jsonMarshaler struct{}

func (j *jsonMarshaler) Marshal(msg proto.Message) ([]byte, error) {
	return protojson.Marshal(msg)
}
func (j *jsonMarshaler) Unmarshal(b []byte, msg proto.Message) error {
	return protojson.Unmarshal(b, msg)
}

// protoMarshaler marshals and unmarshals the Protobuf binary format.
type protoMarshaler struct{}

func (j *protoMarshaler) Marshal(msg proto.Message) ([]byte, error) {
	return proto.Marshal(msg)
}
func (j *protoMarshaler) Unmarshal(b []byte, msg proto.Message) error {
	return proto.Unmarshal(b, msg)
}
