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
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"google.golang.org/protobuf/proto"
)

// ReadDelimitedMessage reads the next message from in. This first reads a
// fixed four byte preface, which is a network-encoded (i.e. big-endian)
// 32-bit integer that represents the message size. This then reads a
// number of bytes equal to that size and unmarshals it into msg.
func ReadDelimitedMessage[T proto.Message](in io.Reader, msg T) error {
	data, err := readDelimitedMessageRaw(in)
	if err != nil {
		return err
	}
	if err := proto.Unmarshal(data, msg); err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}
	return nil
}

func readDelimitedMessageRaw(in io.Reader) ([]byte, error) {
	var lenBuffer [4]byte
	if _, err := io.ReadFull(in, lenBuffer[:]); err != nil {
		return nil, err
	}
	data := make([]byte, binary.BigEndian.Uint32(lenBuffer[:]))
	if _, err := io.ReadFull(in, data); err != nil {
		if errors.Is(err, io.EOF) {
			err = io.ErrUnexpectedEOF
		}
		return nil, err
	}
	return data, nil
}

// WriteDelimitedMessage writes msg to out in a way that can be read by ReadDelimitedMessage.
func WriteDelimitedMessage[T proto.Message](out io.Writer, msg T) error {
	data, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}
	return writeDelimitedMessageRaw(out, data)
}

func writeDelimitedMessageRaw(out io.Writer, data []byte) error {
	var lenBuffer [4]byte
	binary.BigEndian.PutUint32(lenBuffer[:], uint32(len(data)))
	if _, err := out.Write(lenBuffer[:]); err != nil {
		return err
	}
	if _, err := out.Write(data); err != nil {
		return err
	}
	return nil
}
