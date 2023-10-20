package server

import (
	"google.golang.org/protobuf/encoding/protojson"
	proto "google.golang.org/protobuf/proto"
)

type Marshaler interface {
	Marshal(msg proto.Message) (b []byte, err error)
	Unmarshal(b []byte, msg proto.Message) error
}

type jsonMarshaler struct{}

func (j *jsonMarshaler) Marshal(msg proto.Message) (b []byte, err error) {
	return protojson.Marshal(msg)
}
func (j *jsonMarshaler) Unmarshal(b []byte, msg proto.Message) error {
	return protojson.Unmarshal(b, msg)
}

type protoMarshaler struct{}

func (j *protoMarshaler) Marshal(msg proto.Message) (b []byte, err error) {
	return proto.Marshal(msg)
}
func (j *protoMarshaler) Unmarshal(b []byte, msg proto.Message) error {
	return proto.Unmarshal(b, msg)
}

func NewMarshaler(json bool) Marshaler {
	if json {
		return &jsonMarshaler{}
	}
	return &protoMarshaler{}
}
