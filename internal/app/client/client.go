package client

import (
	"context"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"

	"connectrpc.com/conformance/internal/app"
	v1alpha1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1alpha1"
	"connectrpc.com/connect"
	"golang.org/x/net/http2"
)

const (
	ProtocolConnect = "connect"
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

	resp, err := callClient(ctx, req)
	if err != nil {
		return err
	}
	bytes, err := codec.Marshal(resp)
	if err != nil {
		return err
	}
	if _, err := out.Write(bytes); err != nil {
		return err
	}

	return nil
}

func callClient(ctx context.Context, req *v1alpha1.ClientCompatRequest) (*v1alpha1.ClientCompatResponse, error) {
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

	impl := fmt.Sprintf("%s%s", req.Protocol, req.HttpVersion)

	fmt.Println(impl)
	switch impl {
	case v1alpha1.Protocol_PROTOCOL_CONNECT.String() + "-" + v1alpha1.HTTPVersion_HTTP_VERSION_1.String():
		fmt.Println("got eem!!")
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

	var wrapper ClientWrapper
	switch req.Service {
	case "connectrpc.conformance.v1alpha1.ConformanceService":
		fallthrough
	default:
		wrapper = NewConformanceClientWrapper(transport, serverURL, clientOptions)
	}
	return wrapper.Invoke(ctx, req)
}
