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
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/encoding/prototext"
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
	data, err := readDelimitedMessageRaw(p.in)
	if err != nil {
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

// TextConnectCodec implements the connect.Codec interface, providing the
// protobuf text format.
type TextConnectCodec struct {
	prototext.MarshalOptions
	prototext.UnmarshalOptions
}

func (t *TextConnectCodec) Name() string {
	return "text"
}

func (t *TextConnectCodec) Marshal(a any) ([]byte, error) {
	msg, ok := a.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("message type %T does not implement proto.Message", a)
	}
	return t.MarshalOptions.Marshal(msg)
}

func (t *TextConnectCodec) Unmarshal(bytes []byte, a any) error {
	msg, ok := a.(proto.Message)
	if !ok {
		return fmt.Errorf("message type %T does not implement proto.Message", a)
	}
	return t.UnmarshalOptions.Unmarshal(bytes, msg)
}
