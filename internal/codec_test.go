package internal

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestStrictJSONCodec(t *testing.T) {
	codec := StrictJSONCodec{}
	err := codec.Unmarshal([]byte(`{
		"somefield": 123
	}`), &emptypb.Empty{})
	require.ErrorContains(t, err, `unknown field "somefield"`)
}

func TestStrictProtoCodec(t *testing.T) {
	codec := StrictProtoCodec{}
	err := codec.Unmarshal([]byte{8, 0}, &emptypb.Empty{})
	require.ErrorContains(t, err, `message data includes unrecognized field 1 with varint wire type`)
}
