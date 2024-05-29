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

package internal

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
)

// StreamDecoder is used to decode messages from a stream. This is used
// when the input contains a sequence of messages, not just one.
type StreamDecoder interface {
	DecodeNext(msg proto.Message) error
}

// StreamEncoder is used to encode messages to a stream. This is used
// when the output will contain a sequence of messages, not just one.
type StreamEncoder interface {
	Encode(msg proto.Message) error
}

// Codec describes anything that can marshal and unmarshal proto messages.
type Codec interface {
	NewDecoder(io.Reader) StreamDecoder
	NewEncoder(io.Writer) StreamEncoder
}

// NewCodec returns a new Codec.
func NewCodec(json bool) Codec {
	if json {
		return &jsonCodec{MarshalOptions: protojson.MarshalOptions{Multiline: true}}
	}
	return &protoCodec{}
}

// jsonCodec marshals and unmarshals the JSON format.
type jsonCodec struct {
	protojson.MarshalOptions
	protojson.UnmarshalOptions
}

func (c *jsonCodec) NewDecoder(in io.Reader) StreamDecoder {
	dec := json.NewDecoder(in)
	return &jsonDecoder{
		opts:    c.UnmarshalOptions,
		decoder: dec,
	}
}

func (c *jsonCodec) NewEncoder(out io.Writer) StreamEncoder {
	return &jsonEncoder{
		opts: c.MarshalOptions,
		out:  out,
	}
}

type jsonDecoder struct {
	opts    protojson.UnmarshalOptions
	decoder *json.Decoder
}

func (j *jsonDecoder) DecodeNext(msg proto.Message) error {
	var msgData json.RawMessage
	if err := j.decoder.Decode(&msgData); err != nil {
		if errors.Is(err, io.EOF) {
			return err
		}
		return fmt.Errorf("failed to decode JSON message from input: %w", err)
	}
	if err := j.opts.Unmarshal(msgData, msg); err != nil {
		return fmt.Errorf("failed to unmarshal JSON message: %w", err)
	}
	return nil
}

type jsonEncoder struct {
	opts protojson.MarshalOptions
	out  io.Writer
}

func (j *jsonEncoder) Encode(msg proto.Message) error {
	data, err := j.opts.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message to JSON: %w", err)
	}
	if _, err := j.out.Write(data); err != nil {
		return fmt.Errorf("failed to write message to output: %w", err)
	}
	if len(data) > 0 || data[len(data)-1] != '\n' {
		_, _ = j.out.Write([]byte{'\n'}) // best effort newline between JSON outputs
	}
	return nil
}

// protoCodec marshals and unmarshals the Protobuf binary format.
type protoCodec struct {
	proto.MarshalOptions
	proto.UnmarshalOptions
}

func (c *protoCodec) NewDecoder(in io.Reader) StreamDecoder {
	return &protoDecoder{
		opts: c.UnmarshalOptions,
		in:   in,
	}
}

func (c *protoCodec) NewEncoder(out io.Writer) StreamEncoder {
	return &protoEncoder{
		opts: c.MarshalOptions,
		out:  out,
	}
}

type protoDecoder struct {
	opts proto.UnmarshalOptions
	in   io.Reader
}

func (p *protoDecoder) DecodeNext(msg proto.Message) error {
	var lenBuffer [4]byte
	if _, err := io.ReadFull(p.in, lenBuffer[:]); err != nil {
		return err
	}
	data := make([]byte, binary.BigEndian.Uint32(lenBuffer[:]))
	if _, err := io.ReadFull(p.in, data); err != nil {
		if errors.Is(err, io.EOF) {
			err = io.ErrUnexpectedEOF
		}
		return err
	}
	if err := p.opts.Unmarshal(data, msg); err != nil {
		return fmt.Errorf("failed to unmarshal binary message: %w", err)
	}
	return nil
}

type protoEncoder struct {
	opts proto.MarshalOptions
	out  io.Writer
}

func (p *protoEncoder) Encode(msg proto.Message) error {
	data, err := p.opts.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal response to binary: %w", err)
	}
	return writeDelimitedMessageRaw(p.out, data)
}

// StrictJSONCodec is a codec for connect (and implements the optional methods
// for stable output and for appending to existing bytes for better performance).
// It is like Connect's builtin JSON codec, except that it does NOT allow
// unrecognized field names. (Connect's default JSON codec is lenient and discards
// unrecognized fields.)
type StrictJSONCodec struct{}

var _ connect.Codec = StrictJSONCodec{}

func (s StrictJSONCodec) Name() string {
	return "json"
}

func (s StrictJSONCodec) Marshal(msg any) ([]byte, error) {
	return s.MarshalAppend(nil, msg)
}

func (s StrictJSONCodec) Unmarshal(data []byte, msg any) error {
	protoMsg, ok := msg.(proto.Message)
	if !ok {
		return fmt.Errorf("message type %T is not a proto.Message", msg)
	}
	// The default JSON codec for connect has DiscardUnknown set.
	//    https://github.com/connectrpc/connect-go/blob/main/codec.go#L178-L180
	// We don't  set it so that we get stricter behavior: an error will occur
	// if the client sends any unrecognized fields.
	return protojson.Unmarshal(data, protoMsg)
}

func (s StrictJSONCodec) MarshalAppend(b []byte, msg any) ([]byte, error) {
	protoMsg, ok := msg.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("message type %T is not a proto.Message", msg)
	}
	return protojson.MarshalOptions{}.MarshalAppend(b, protoMsg)
}

func (s StrictJSONCodec) MarshalStable(msg any) ([]byte, error) {
	data, err := s.Marshal(msg)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := json.Compact(&buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (s StrictJSONCodec) IsBinary() bool {
	return false
}

// StrictProtoCodec is a codec for connect (and implements the optional methods
// for stable output and for appending to existing bytes for better performance).
// It is like Connect's builtin Proto codec, except that it does NOT allow
// unrecognized fields. If the peer sends a message where not all elements are
// recognized, it will return an error.
type StrictProtoCodec struct{}

var _ connect.Codec = StrictProtoCodec{}

func (s StrictProtoCodec) Name() string {
	return "proto"
}

func (s StrictProtoCodec) Marshal(msg any) ([]byte, error) {
	return s.MarshalAppend(nil, msg)
}

func (s StrictProtoCodec) Unmarshal(data []byte, msg any) error {
	protoMsg, ok := msg.(proto.Message)
	if !ok {
		return fmt.Errorf("message type %T is not a proto.Message", msg)
	}
	if err := proto.Unmarshal(data, protoMsg); err != nil {
		return err
	}
	// We are being strict and thus disallow any unrecognized fields.
	unrecognized := protoMsg.ProtoReflect().GetUnknown()
	if len(unrecognized) == 0 {
		return nil
	}
	num, typ, length := protowire.ConsumeTag(unrecognized)
	if length <= 0 {
		// Should not be possible since above call to proto.Unmarshal succeeded.
		l := len(unrecognized)
		var suffix string
		if l > 50 {
			unrecognized = unrecognized[:50]
			suffix = "..."
		}
		return fmt.Errorf("message data included %d unprocessable bytes: %x%s", l, unrecognized, suffix)
	}
	var wireType string
	switch typ {
	case protowire.VarintType:
		wireType = "varint"
	case protowire.Fixed32Type:
		wireType = "fixed32"
	case protowire.Fixed64Type:
		wireType = "fixed64"
	case protowire.BytesType:
		wireType = "bytes"
	case protowire.StartGroupType:
		wireType = "start-group"
	case protowire.EndGroupType:
		// This and the default case below should not really be possible
		// since above call to proto.Unmarshal succeeded.
		return fmt.Errorf("message data included field %d that incorrectly starts with end-group wire type", num)
	default:
		return fmt.Errorf("message data included field %d that uses unknown wire type %d", num, typ)
	}
	return fmt.Errorf("message data includes unrecognized field %d with %s wire type", num, wireType)
}

func (s StrictProtoCodec) MarshalAppend(b []byte, msg any) ([]byte, error) {
	protoMsg, ok := msg.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("message type %T is not a proto.Message", msg)
	}
	return protojson.MarshalOptions{}.MarshalAppend(b, protoMsg)
}

func (s StrictProtoCodec) MarshalStable(msg any) ([]byte, error) {
	protoMsg, ok := msg.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("message type %T is not a proto.Message", msg)
	}
	return proto.MarshalOptions{Deterministic: true}.Marshal(protoMsg)
}

func (s StrictProtoCodec) IsBinary() bool {
	return true
}
