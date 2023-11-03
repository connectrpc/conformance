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

package connectconformance

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"google.golang.org/protobuf/proto"
)

func readDelimitedMessage[T proto.Message](in io.Reader, msg T) error {
	var lenBuffer [4]byte
	if _, err := io.ReadFull(in, lenBuffer[:]); err != nil {
		return err
	}
	data := make([]byte, binary.BigEndian.Uint32(lenBuffer[:]))
	if _, err := io.ReadFull(in, data); err != nil {
		if errors.Is(err, io.EOF) {
			err = io.ErrUnexpectedEOF
		}
		return err
	}
	if err := proto.Unmarshal(data, msg); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return nil
}

func writeDelimitedMessage[T proto.Message](out io.Writer, msg T) error {
	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
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
