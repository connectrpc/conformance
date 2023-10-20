package client

import (
	"context"
	"flag"
	"fmt"
	"io"

	"connectrpc.com/conformance/internal/app"
	v1alpha1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1alpha1"
)

// Run runs the server according to server config read from the 'in' reader.
func Run(ctx context.Context, args []string, in io.ReadCloser, out io.WriteCloser) error {
	json := flag.Bool("json", false, "whether to use the JSON format for marshaling / unmarshaling messages")

	flag.Parse()

	// Read the server config from  the in reader
	data, err := io.ReadAll(in)
	if err != nil {
		return err
	}

	marshaler := app.NewMarshaler(*json)

	req := &v1alpha1.ClientCompatRequest{}
	if err := marshaler.Unmarshal(data, req); err != nil {
		return err
	}

	fmt.Printf("%+v", req)

	return nil
}
