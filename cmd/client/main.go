// Copyright 2022-2023 The Connect Authors
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

package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"connectrpc.com/conformance/internal/gen/proto/connect/connectrpc/conformance/v1alpha1/conformancev1alpha1connect"
	v1alpha1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1alpha1"
	"connectrpc.com/connect"
	"github.com/spf13/cobra"
	"golang.org/x/net/http2"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
)

const (
	hostFlagName           = "host"
	portFlagName           = "port"
	implementationFlagName = "implementation"
)

type flags struct {
	host           string
	port           string
	implementation string
}

const (
	connectH1        = "connect-h1"
	connectH2        = "connect-h2"
	connectGRPCH1    = "connect-grpc-h1"
	connectGRPCH2    = "connect-grpc-h2"
	connectGRPCWebH1 = "connect-grpc-web-h1"
	connectGRPCWebH2 = "connect-grpc-web-h2"
)

func main() {
	flagset := &flags{}
	rootCmd := &cobra.Command{
		Use:   "client",
		Short: "Starts a Connect Go client based on implementation",
		Run: func(cmd *cobra.Command, args []string) {
			run(flagset)
		},
	}
	if err := bind(rootCmd, flagset); err != nil {
		os.Exit(1)
	}
	_ = rootCmd.Execute()
}

func bind(cmd *cobra.Command, flags *flags) error {
	cmd.Flags().StringVar(&flags.host, hostFlagName, "127.0.0.1", "the host name of the test server")
	cmd.Flags().StringVar(&flags.port, portFlagName, "", "the port of the test server")
	cmd.Flags().StringVarP(
		&flags.implementation,
		implementationFlagName,
		"i",
		"",
		fmt.Sprintf(
			"the client implementation tested, accepted values are %q, %q, %q, %q, %q, %q",
			connectH1,
			connectH2,
			connectGRPCH1,
			connectGRPCH2,
			connectGRPCWebH1,
			connectGRPCWebH2,
		),
	)
	return nil
}

func run(flags *flags) {
	// tests for connect clients
	var scheme string
	serverURL, err := url.ParseRequestURI("http://" + net.JoinHostPort(flags.host, flags.port))
	if err != nil {
		log.Fatalf("invalid url: %s", scheme+net.JoinHostPort(flags.host, flags.port))
	}
	// create transport base on HTTP protocol of the implementation
	var transport http.RoundTripper

	// create transport base on HTTP protocol of the implementation
	switch flags.implementation {
	case connectH1, connectGRPCH1, connectGRPCWebH1:
		transport = &http.Transport{}
	case connectGRPCH2, connectH2, connectGRPCWebH2:
		transport = &http2.Transport{
			AllowHTTP: true,
			DialTLSContext: func(ctx context.Context, network, addr string, cfg *tls.Config) (net.Conn, error) {
				return net.Dial(network, addr)
			},
		}
	default:
		log.Fatalf(`the --implementation or -i flag is invalid"`)
	}

	// create client options based on protocol of the implementation
	clientOptions := []connect.ClientOption{connect.WithHTTPGet()}
	switch flags.implementation {
	case connectGRPCH1, connectGRPCH2:
		clientOptions = append(clientOptions, connect.WithGRPC())
	case connectGRPCWebH1, connectGRPCWebH2:
		clientOptions = append(clientOptions, connect.WithGRPCWeb())
	}
	client := conformancev1alpha1connect.NewConformanceServiceClient(
		&http.Client{Transport: transport},
		serverURL.String(),
		clientOptions...,
	)

	runUnary(client)
	runServerStream(client)
	runBidiStream(client)
}

func runUnary(client conformancev1alpha1connect.ConformanceServiceClient) {
	runUnary1(client)
	runUnary2(client)
	runUnary3(client)
}

func runUnary1(client conformancev1alpha1connect.ConformanceServiceClient) {
	fmt.Println("runUnary1 (happy path w/ response data) --------")
	req := connect.NewRequest(&v1alpha1.UnaryRequest{
		ResponseDefinition: &v1alpha1.UnaryResponseDefinition{
			Response: &v1alpha1.UnaryResponseDefinition_ResponseData{
				ResponseData: []byte("test response"),
			},
			ResponseHeaders: []*v1alpha1.Header{
				{
					Name:  "x-custom-header",
					Value: []string{"foo", "bar", "baz"},
				},
			},
			ResponseTrailers: []*v1alpha1.Header{
				{
					Name:  "x-custom-trailer",
					Value: []string{"bing", "quux"},
				},
			},
		},
	})
	req.Header().Set(
		"Greet-Emoji-Bin",
		connect.EncodeBinaryHeader([]byte("👋")),
	)
	reply, err := client.Unary(context.Background(), req)
	fmt.Printf("Response: %+v\n", reply)
	if err != nil {
		printError(err)
	}
}

func runUnary2(client conformancev1alpha1connect.ConformanceServiceClient) {
	fmt.Println("runUnary2 (happy path w/ error response) --------")
	retryInfo := &errdetails.RetryInfo{
		RetryDelay: durationpb.New(10 * time.Second),
	}
	retryAny, err := anypb.New(retryInfo)
	if err != nil {
		fmt.Println(err)
		return
	}
	reply, err := client.Unary(
		context.Background(),
		connect.NewRequest(&v1alpha1.UnaryRequest{
			ResponseDefinition: &v1alpha1.UnaryResponseDefinition{
				Response: &v1alpha1.UnaryResponseDefinition_Error{
					Error: &v1alpha1.Error{
						Code:    int32(connect.CodeAborted),
						Message: "The request has failed",
						Details: []*anypb.Any{retryAny},
					},
				},
				ResponseHeaders: []*v1alpha1.Header{
					{
						Name:  "x-custom-header",
						Value: []string{"foo", "bar", "baz", "bing"},
					},
				},
			},
		}),
	)
	fmt.Printf("Response: %+v\n", reply)
	if err != nil {
		printError(err)
	}
}

func runUnary3(client conformancev1alpha1connect.ConformanceServiceClient) {
	fmt.Println("runUnary3 (empty UnaryRequest) --------")
	reply, err := client.Unary(
		context.Background(),
		connect.NewRequest(&v1alpha1.UnaryRequest{}),
	)
	fmt.Printf("Response: %+v\n", reply)
	if err != nil {
		printError(err)
	}
}

func printError(err error) {
	if connectErr := new(connect.Error); errors.As(err, &connectErr) {
		fmt.Printf("Error Code: %d\n", connectErr.Code())
		fmt.Printf("Error Message: %s\n", connectErr.Message())
		fmt.Printf("Error Details: (%d)\n", len(connectErr.Details()))
		for _, detail := range connectErr.Details() {
			fmt.Printf("  %+v\n", detail.Type())
			msg, valueErr := detail.Value()
			if valueErr != nil {
				fmt.Println(valueErr)
				continue
			}
			if retryInfo, ok := msg.(*errdetails.RetryInfo); ok {
				fmt.Printf("  %+v\n", retryInfo)
			}
		}
	}
}

func runServerStream(client conformancev1alpha1connect.ConformanceServiceClient) {
	runServerStream1(client)
	runServerStream2(client)
	runServerStream3(client)
}

func runServerStream1(client conformancev1alpha1connect.ConformanceServiceClient) {
	fmt.Println("runServerStream1 (happy path w/ no error) --------")
	req := connect.NewRequest(&v1alpha1.ServerStreamRequest{
		ResponseDefinition: &v1alpha1.StreamResponseDefinition{
			ResponseData:    [][]byte{[]byte("response 1"), []byte("response 2"), []byte("response 3")},
			ResponseDelayMs: 2000,
			ResponseHeaders: []*v1alpha1.Header{
				{
					Name:  "x-custom-header",
					Value: []string{"foo", "bar", "baz"},
				},
			},
			ResponseTrailers: []*v1alpha1.Header{
				{
					Name:  "x-custom-trailer",
					Value: []string{"bing", "quux"},
				},
			},
		},
	})
	req.Header().Set(
		"Greet-Emoji-Bin",
		connect.EncodeBinaryHeader([]byte("👋")),
	)
	stream, err := client.ServerStream(context.Background(), req)
	for stream.Receive() {
		if stream.Err() != nil {
			printError(err)
			return
		}
		fmt.Printf("Response: %+v\n", stream.Msg())
		fmt.Println(stream.ResponseHeader())
		fmt.Println(stream.ResponseTrailer())
	}
}

func runServerStream2(client conformancev1alpha1connect.ConformanceServiceClient) {
	fmt.Println("runServerStream2 (happy path w/ error response) --------")
	retryInfo := &errdetails.RetryInfo{
		RetryDelay: durationpb.New(10 * time.Second),
	}
	retryAny, err := anypb.New(retryInfo)
	req := connect.NewRequest(&v1alpha1.ServerStreamRequest{
		ResponseDefinition: &v1alpha1.StreamResponseDefinition{
			ResponseData:    [][]byte{[]byte("response 1"), []byte("response 2"), []byte("response 3")},
			ResponseDelayMs: 2000,
			ResponseHeaders: []*v1alpha1.Header{
				{
					Name:  "x-custom-header",
					Value: []string{"foo", "bar", "baz"},
				},
			},
			ResponseTrailers: []*v1alpha1.Header{
				{
					Name:  "x-custom-trailer",
					Value: []string{"bing", "quux"},
				},
			},
			Error: &v1alpha1.Error{
				Code:    int32(connect.CodeAborted),
				Message: "The request has failed",
				Details: []*anypb.Any{retryAny},
			},
		},
	})
	req.Header().Set(
		"Greet-Emoji-Bin",
		connect.EncodeBinaryHeader([]byte("👋")),
	)
	stream, err := client.ServerStream(context.Background(), req)
	for stream.Receive() {
		if stream.Err() != nil {
			printError(err)
			return
		}
		fmt.Printf("Response: %+v\n", stream.Msg())
	}
	if err := stream.Err(); err != nil {
		printError(err)
	}
}

func runServerStream3(client conformancev1alpha1connect.ConformanceServiceClient) {
	fmt.Println("runServerStream3 (no response, only an error) --------")
	retryInfo := &errdetails.RetryInfo{
		RetryDelay: durationpb.New(10 * time.Second),
	}
	retryAny, err := anypb.New(retryInfo)
	req := connect.NewRequest(&v1alpha1.ServerStreamRequest{
		ResponseDefinition: &v1alpha1.StreamResponseDefinition{
			ResponseDelayMs: 2000,
			ResponseHeaders: []*v1alpha1.Header{
				{
					Name:  "x-custom-header",
					Value: []string{"foo", "bar", "baz"},
				},
			},
			ResponseTrailers: []*v1alpha1.Header{
				{
					Name:  "x-custom-trailer",
					Value: []string{"bing", "quux"},
				},
			},
			Error: &v1alpha1.Error{
				Code:    int32(connect.CodeAborted),
				Message: "The request has failed",
				Details: []*anypb.Any{retryAny},
			},
		},
	})
	req.Header().Set(
		"Greet-Emoji-Bin",
		connect.EncodeBinaryHeader([]byte("👋")),
	)
	stream, err := client.ServerStream(context.Background(), req)
	for stream.Receive() {
		if stream.Err() != nil {
			printError(err)
			return
		}
		fmt.Printf("Response: %+v\n", stream.Msg())
	}
	if err := stream.Err(); err != nil {
		printError(err)
	}
}

func runBidiStream(client conformancev1alpha1connect.ConformanceServiceClient) {
	runBidiStream1(client)
	runBidiStream2(client)
}

func runBidiStream1(client conformancev1alpha1connect.ConformanceServiceClient) {
	fmt.Println("runBidiStream1 (full duplex) --------")
	retryInfo := &errdetails.RetryInfo{
		RetryDelay: durationpb.New(10 * time.Second),
	}
	retryAny, _ := anypb.New(retryInfo)
	req := &v1alpha1.BidiStreamRequest{
		FullDuplex: true,
		ResponseDefinition: &v1alpha1.StreamResponseDefinition{
			ResponseData:    [][]byte{[]byte("response 1"), []byte("response 2"), []byte("response 3")},
			ResponseDelayMs: 2000,
			ResponseHeaders: []*v1alpha1.Header{
				{
					Name:  "x-custom-header",
					Value: []string{"foo", "bar", "baz"},
				},
			},
			ResponseTrailers: []*v1alpha1.Header{
				{
					Name:  "x-custom-trailer",
					Value: []string{"bing", "quux"},
				},
			},
			Error: &v1alpha1.Error{
				Code:    int32(connect.CodeAborted),
				Message: "The request has failed",
				Details: []*anypb.Any{retryAny},
			},
		}}

	// req.Header().Set(
	// 	"Greet-Emoji-Bin",
	// 	connect.EncodeBinaryHeader([]byte("👋")),
	// )

	stream := client.BidiStream(context.Background())
	fmt.Println("Sending first message")
	stream.Send(req)
	res, err := stream.Receive()
	fmt.Printf("Response to first message: %+v\n", res)
	if err != nil {
		printError(err)
	}

	fmt.Println("Sending second message")
	stream.Send(&v1alpha1.BidiStreamRequest{
		RequestData: []byte("second request"),
	})
	res, err = stream.Receive()
	fmt.Printf("Response to second message: %+v\n", res)
	if err != nil {
		printError(err)
	}

	fmt.Println("Sending third message")
	stream.Send(&v1alpha1.BidiStreamRequest{
		RequestData: []byte("third request"),
	})
	res, err = stream.Receive()
	fmt.Printf("Response to third message: %+v\n", res)
	if err != nil {
		printError(err)
	}
	stream.CloseRequest()
	res, err = stream.Receive()
	if res != nil {
		fmt.Printf("error: received an unexpected message %+v: ", res)
	}
	if !errors.Is(err, io.EOF) {
		printError(err)
	}
	stream.CloseResponse()
}

func runBidiStream2(client conformancev1alpha1connect.ConformanceServiceClient) {
	fmt.Println("runBidiStream2 (half duplex) --------")
	retryInfo := &errdetails.RetryInfo{
		RetryDelay: durationpb.New(10 * time.Second),
	}
	retryAny, _ := anypb.New(retryInfo)
	req := &v1alpha1.BidiStreamRequest{
		FullDuplex: true,
		ResponseDefinition: &v1alpha1.StreamResponseDefinition{
			ResponseData:    [][]byte{[]byte("response 1"), []byte("response 2"), []byte("response 3")},
			ResponseDelayMs: 2000,
			ResponseHeaders: []*v1alpha1.Header{
				{
					Name:  "x-custom-header",
					Value: []string{"foo", "bar", "baz"},
				},
			},
			ResponseTrailers: []*v1alpha1.Header{
				{
					Name:  "x-custom-trailer",
					Value: []string{"bing", "quux"},
				},
			},
			Error: &v1alpha1.Error{
				Code:    int32(connect.CodeAborted),
				Message: "The request has failed",
				Details: []*anypb.Any{retryAny},
			},
		}}

	stream := client.BidiStream(context.Background())
	fmt.Println("Sending first message")
	stream.Send(req)
	fmt.Println("Sending second message")
	stream.Send(&v1alpha1.BidiStreamRequest{
		RequestData: []byte("second request"),
	})
	fmt.Println("Sending third message")
	stream.Send(&v1alpha1.BidiStreamRequest{
		RequestData: []byte("third request"),
	})
	stream.CloseRequest()

	for i := 0; i < len(req.ResponseDefinition.ResponseData); i++ {
		res, err := stream.Receive()
		fmt.Printf("Response: %+v\n", res)
		if err != nil {
			printError(err)
		}
	}
	res, err := stream.Receive()
	if res != nil {
		fmt.Printf("error: received an unexpected message %+v: ", res)
	}
	if !errors.Is(err, io.EOF) {
		printError(err)
	}
	stream.CloseResponse()
}
