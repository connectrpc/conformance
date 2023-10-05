package main

import (
	"context"

	"connectrpc.com/conformance/internal/gen/proto/connect/connectrpc/conformance/v1alpha1/conformancev1alpha1connect"
	"connectrpc.com/connect"
)

// NewConformanceServiceHandler returns a new ConformanceServiceHandler.
func NewTestServiceHandler() conformancev1alpha1connect.ConformanceServiceHandler {
	return &testServer{}
}

type testServer struct {
	conformanceconnect.UnimplementedTestServiceHandler
}

func (s *testServer) CacheableUnaryCall(ctx context.Context, request *connect.Request[conformance.SimpleRequest]) (*connect.Response[conformance.SimpleResponse], error) {
	response, err := s.UnaryCall(ctx, request)
	if response != nil {
		if request.Peer().Query.Has("message") {
			response.Header().Set("Get-Request", "true")
		}
	}
	return response, err
}
