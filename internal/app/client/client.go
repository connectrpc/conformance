package client

import (
	"context"
	"crypto/tls"
	"errors"
	"flag"
	"io"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"strconv"

	"connectrpc.com/conformance/internal/app"
	"connectrpc.com/conformance/internal/gen/proto/connect/connectrpc/conformance/v1alpha1/conformancev1alpha1connect"
	v1alpha1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1alpha1"
	"connectrpc.com/connect"
	"golang.org/x/net/http2"
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

	codec := app.NewCodec(*json)

	req := &v1alpha1.ClientCompatRequest{}
	if err := codec.Unmarshal(data, req); err != nil {
		return err
	}

	client, err := newClient(req)
	if err != nil {
		return err
	}

	// TODO - How do we know what type this is
	ur := &v1alpha1.UnaryRequest{}
	// TODO - Need to loop over these and send all of them
	req.RequestMessages[0].UnmarshalTo(ur)

	request := connect.NewRequest(ur)

	for _, header := range req.RequestHeaders {
		for _, val := range header.Value {
			request.Header().Add(header.Name, val)
		}
	}
	// TODO - Cleanup this reflection nonsense
	inv, err := Invoke(client, req.Method, context.Background(), request)
	if err != nil {
		return err
	}
	// TODO - How do we know what type this is
	reply := inv[0].Interface().(*connect.Response[v1alpha1.UnaryResponse])
	bytes, err := codec.Marshal(reply.Msg)
	if err != nil {
		return err
	}
	if _, err := out.Write(bytes); err != nil {
		return err
	}

	return nil
}

func Invoke(any interface{}, name string, args ...interface{}) ([]reflect.Value, error) {
	inputs := make([]reflect.Value, len(args))
	for i := range args {
		inputs[i] = reflect.ValueOf(args[i])
	}
	method := reflect.ValueOf(any).MethodByName(name)
	if !method.IsValid() {
		return nil, errors.New("method " + name + " does not exist")
	}

	return method.Call(inputs), nil
}

// TODO - This always assumes a ConformanceServiceClient
// Can we use the req.Service value somehow?
func newClient(req *v1alpha1.ClientCompatRequest) (conformancev1alpha1connect.ConformanceServiceClient, error) {
	var scheme string
	if req.ServerTlsCert != nil {
		scheme = "https://"
	} else {
		scheme = "http://"
	}
	urlString := scheme + net.JoinHostPort(req.Host, strconv.FormatUint(uint64(req.Port), 10))
	serverURL, err := url.ParseRequestURI(urlString)
	if err != nil {
		return nil, errors.New("invalid url: %s" + urlString)
	}

	var implementation string

	switch req.Protocol {
	case v1alpha1.Protocol_PROTOCOL_CONNECT:
		implementation = "connect"
	case v1alpha1.Protocol_PROTOCOL_GRPC:
		implementation = "connectgrpc"
	case v1alpha1.Protocol_PROTOCOL_GRPC_WEB:
		implementation = "connectgrpcweb"
	case v1alpha1.Protocol_PROTOCOL_UNSPECIFIED:
		return nil, errors.New("an protocol must be specified.")
	}

	switch req.HttpVersion {
	case v1alpha1.HTTPVersion_HTTP_VERSION_1:
		implementation += "h1"
	case v1alpha1.HTTPVersion_HTTP_VERSION_2:
		implementation += "h2"
	case v1alpha1.HTTPVersion_HTTP_VERSION_3:
		implementation += "h3"
	case v1alpha1.HTTPVersion_HTTP_VERSION_UNSPECIFIED:
		return nil, errors.New("an HTTP version must be specified.")
	}

	var transport http.RoundTripper
	// create transport base on HTTP protocol of the implementation
	switch implementation {
	case "connecth1", "connectgrpch1", "connectgrpcwebh1":
		transport = &http.Transport{}
	case "connecth2", "connectgrpch2", "connectgrpcwebh2":
		transport = &http2.Transport{
			AllowHTTP: true,
			DialTLSContext: func(ctx context.Context, network, addr string, cfg *tls.Config) (net.Conn, error) {
				return net.Dial(network, addr)
			},
		}
	default:
		return nil, errors.New("invalid implementation: " + implementation)
	}

	// Create client options based on protocol of the implementation
	clientOptions := []connect.ClientOption{connect.WithHTTPGet()}
	switch implementation {
	case "connectgrpch1", "connectgrpch2":
		clientOptions = append(clientOptions, connect.WithGRPC())
	case "connectgrpcwebh1", "connectgrpcwebh2":
		clientOptions = append(clientOptions, connect.WithGRPCWeb())
	}

	if req.Codec == v1alpha1.Codec_CODEC_JSON {
		clientOptions = append(clientOptions, connect.WithProtoJSON())
	}

	// TODO - How do we configure each compression algo? i.e.
	// how do we know the string to use?
	switch req.Compression {
	case v1alpha1.Compression_COMPRESSION_GZIP:
		clientOptions = append(clientOptions, connect.WithSendGzip())
	case v1alpha1.Compression_COMPRESSION_IDENTITY:
	case v1alpha1.Compression_COMPRESSION_BR:
	case v1alpha1.Compression_COMPRESSION_ZSTD:
	case v1alpha1.Compression_COMPRESSION_DEFLATE:
	case v1alpha1.Compression_COMPRESSION_SNAPPY:
	case v1alpha1.Compression_COMPRESSION_UNSPECIFIED:
		// Do nothing
	}

	return conformancev1alpha1connect.NewConformanceServiceClient(
		&http.Client{Transport: transport},
		serverURL.String(),
		clientOptions...,
	), nil
}
