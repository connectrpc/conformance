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
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestStrictJSONCodec(t *testing.T) {
	t.Parallel()
	codec := StrictJSONCodec{}
	err := codec.Unmarshal([]byte(`{
		"somefield": 123
	}`), &emptypb.Empty{})
	require.ErrorContains(t, err, `unknown field "somefield"`)
}

func TestStrictProtoCodec(t *testing.T) {
	t.Parallel()
	codec := StrictProtoCodec{}
	err := codec.Unmarshal([]byte{8, 0}, &emptypb.Empty{})
	require.ErrorContains(t, err, `message data includes unrecognized field 1 with varint wire type`)
}
