package client

import (
	"context"
	"net/http"
	"net/url"

	"connectrpc.com/conformance/internal/gen/proto/connect/connectrpc/conformance/v1alpha1/conformancev1alpha1connect"
	v1alpha1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1alpha1"
	"connectrpc.com/connect"
)

type ClientWrapper interface {
	Invoke(context.Context, *v1alpha1.ClientCompatRequest) (*v1alpha1.ClientCompatResponse, error)
}

type conformanceClientWrapper struct {
	client conformancev1alpha1connect.ConformanceServiceClient
}

func (w *conformanceClientWrapper) Invoke(ctx context.Context, req *v1alpha1.ClientCompatRequest) (*v1alpha1.ClientCompatResponse, error) {
	switch req.Method {
	case "Unary":
		for _, msg := range req.RequestMessages {
			ur := &v1alpha1.UnaryRequest{}
			msg.UnmarshalTo(ur)

			request := connect.NewRequest(ur)

			for _, header := range req.RequestHeaders {
				for _, val := range header.Value {
					request.Header().Add(header.Name, val)
				}
			}

			_, err := w.unary(ctx, request)
			if err != nil {
				return nil, err
			}
		}
		// TODO convert the above to clientcompatresonse
		resp := &v1alpha1.ClientCompatResponse{
			TestName: req.TestName,
		}
		return resp, nil
	}

	return nil, nil
}

func (w *conformanceClientWrapper) unary(ctx context.Context, req *connect.Request[v1alpha1.UnaryRequest]) (*v1alpha1.UnaryResponse, error) {
	resp, err := w.client.Unary(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Msg, nil
}

func NewConformanceClientWrapper(transport http.RoundTripper, url *url.URL, opts []connect.ClientOption) ClientWrapper {
	client := conformancev1alpha1connect.NewConformanceServiceClient(
		&http.Client{Transport: transport},
		url.String(),
		opts...,
	)
	return &conformanceClientWrapper{
		client: client,
	}
}
