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
	"errors"
	"fmt"
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
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
)

const (
	hostFlagName = "host"
	portFlagName = "port"
)

type flags struct {
	host string
	port string
}

func main() {
	flagset := &flags{}
	rootCmd := &cobra.Command{
		Use:   "client",
		Short: "Starts a Connect Go client",
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
	transport := &http.Transport{}
	clientOptions := []connect.ClientOption{connect.WithHTTPGet()}
	client := conformancev1alpha1connect.NewConformanceServiceClient(
		&http.Client{Transport: transport},
		serverURL.String(),
		clientOptions...,
	)

	runUnary1(client)
	runUnary2(client)
}

func runUnary1(client conformancev1alpha1connect.ConformanceServiceClient) {
	fmt.Println("runUnary1 --------")
	reply, err := client.Unary(
		context.Background(),
		connect.NewRequest(&v1alpha1.UnaryRequest{
			Response: &v1alpha1.UnaryRequest_ResponseData{
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
		}),
	)
	fmt.Printf("Response: %+v\n", reply)
	if err != nil {
		printError(err)
	}
}

func runUnary2(client conformancev1alpha1connect.ConformanceServiceClient) {
	fmt.Println("runUnary2 --------")
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
			Response: &v1alpha1.UnaryRequest_Error{
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
		}),
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
