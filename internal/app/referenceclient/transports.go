package referenceclient

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	conformancev1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1"
	"connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1/conformancev1connect"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"golang.org/x/net/http2"
)

type transportSpec struct {
	httpVersion   conformancev1.HTTPVersion
	serverTLSCert string
	clientTLSCert string
	clientTLSKey  string
}

type transports struct {
	cache sync.Map // map[transportSpec]http.RoundTripper
}

func (t *transports) get(req *conformancev1.ClientCompatRequest) (http.RoundTripper, error) {
	spec := transportSpec{
		httpVersion:   req.GetHttpVersion(),
		serverTLSCert: string(req.GetServerTlsCert()),
		clientTLSCert: string(req.GetClientTlsCreds().GetCert()),
		clientTLSKey:  string(req.GetClientTlsCreds().GetKey()),
	}

	// Optimistically skip logic if it's already cached. We will still do an
	// atomic store to share the transport in all cases even if this misses.
	if tr, ok := t.cache.Load(spec); ok {
		return tr.(http.RoundTripper), nil //nolint:errcheck,forcetypeassert
	}

	tlsConf, err := createTLSConfig(req)
	if err != nil {
		return nil, err
	}
	var transport http.RoundTripper
	switch req.HttpVersion {
	case conformancev1.HTTPVersion_HTTP_VERSION_1:
		if tlsConf != nil {
			tlsConf.NextProtos = []string{"http/1.1"}
		}
		tx := &http.Transport{
			DisableCompression: true,
			TLSClientConfig:    tlsConf,
		}
		transport = roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			resp, err := tx.RoundTrip(req)
			if resp != nil &&
				strings.HasSuffix(req.URL.Path, conformancev1connect.ConformanceServiceBidiStreamProcedure) {
				// To force support for bidirectional RPC over HTTP 1.1 (for half-duplex testing),
				// we "trick" the client into thinking this is HTTP/2. We have to do this because
				// otherwise, connect-go refuses to support bidi streams over HTTP 1.1.
				resp.ProtoMajor, resp.ProtoMinor = 2, 0
			}
			return resp, err
		})
	case conformancev1.HTTPVersion_HTTP_VERSION_2:
		if tlsConf != nil {
			tlsConf.NextProtos = []string{"h2"}
			transport = &http.Transport{
				DisableCompression: true,
				TLSClientConfig:    tlsConf,
				ForceAttemptHTTP2:  true,
			}
		} else {
			transport = &http2.Transport{
				DisableCompression: true,
				AllowHTTP:          true,
				DialTLSContext: func(ctx context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
					return (&net.Dialer{}).DialContext(ctx, network, addr)
				},
			}
		}
	case conformancev1.HTTPVersion_HTTP_VERSION_3:
		if tlsConf == nil {
			return nil, errors.New("HTTP/3 indicated in request but no TLS info provided")
		}
		transport = &contextFixTransport{http3.Transport{
			DisableCompression: true,
			TLSClientConfig:    tlsConf,
			QUICConfig:         &quic.Config{MaxIdleTimeout: 20 * time.Second, KeepAlivePeriod: 5 * time.Second},
		}}
	case conformancev1.HTTPVersion_HTTP_VERSION_UNSPECIFIED:
		return nil, errors.New("an HTTP version must be specified")
	}

	// Even if two requests for the same spec make it here, they will use the same connection.
	actual, _ := t.cache.LoadOrStore(spec, transport)
	return actual.(http.RoundTripper), nil //nolint:errcheck,forcetypeassert
}

// contextFixTransport wraps an HTTP/3 transport so that context errors can be correctly
// classified by the connect-go framework. This is a work-around until a fix
// can be implemented in connect-go and/or quic-go.
// See: https://github.com/quic-go/quic-go/issues/4196
type contextFixTransport struct {
	http3.Transport
}

func (t *contextFixTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()
	resp, err := t.Transport.RoundTrip(req)
	if err != nil {
		return nil, maybeWrapContextError(ctx, err)
	}
	resp.Body = &contextFixReader{ctx: ctx, r: resp.Body}
	return resp, nil
}

type contextFixReader struct {
	ctx context.Context //nolint:containedctx
	r   io.ReadCloser
}

func (r *contextFixReader) Read(data []byte) (int, error) {
	n, err := r.r.Read(data)
	return n, maybeWrapContextError(r.ctx, err)
}

func (r *contextFixReader) Close() error {
	return maybeWrapContextError(r.ctx, r.r.Close())
}

func maybeWrapContextError(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}
	ctxErr := ctx.Err()
	if ctxErr == nil {
		return err
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return &contextFixError{timeout: true, error: err}
	}
	var httpErr *http3.Error
	if errors.As(err, &httpErr) && httpErr.ErrorCode == http3.ErrCodeRequestCanceled {
		return &contextFixError{timeout: errors.Is(ctxErr, context.DeadlineExceeded), error: err}
	}
	return err
}

type contextFixError struct {
	timeout bool
	error
}

//nolint:goerr113
func (e *contextFixError) Is(err error) bool {
	return (e.timeout && err == context.DeadlineExceeded) ||
		(!e.timeout && err == context.Canceled)
}
