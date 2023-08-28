package interopconnect

import (
	"context"
	"errors"
	"io"
	"net/http"
	"sync/atomic"
)

// Trailers is a container for trailers captured during the course of an HTTP round trip.
type Trailers struct {
	val atomic.Pointer[http.Header]
}

// Get returns the trailers captured. Trailers are not captured until the response body is
// exhausted.
func (t *Trailers) Get() http.Header {
	headerPtr := t.val.Load()
	if headerPtr == nil {
		return nil
	}
	return *headerPtr
}

type trailersKey struct{}

// CaptureTrailers returns a context to be used with HTTP operations to capture trailers.
// Each HTTP operation used with the returned context will store its HTTP trailers into
// the returned *Trailers value.
func CaptureTrailers(ctx context.Context) (context.Context, *Trailers) {
	trailers := &Trailers{}
	ctx = context.WithValue(ctx, trailersKey{}, trailers)
	return ctx, trailers
}

// TrailerInterceptor is an HTTP transport that supports capturing trailers. Callers
// must decorate a transport with this type, and then they can use CaptureTrailers to
// sniff the HTTP trailers from the request.
type TrailerInterceptor struct {
	Transport http.RoundTripper
}

func (t *TrailerInterceptor) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.Transport.RoundTrip(req)
	trailers, ok := req.Context().Value(trailersKey{}).(*Trailers)
	if err != nil || !ok {
		return resp, err
	}
	resp.Body = &captureTrailersAtEOF{r: resp.Body, resp: resp, trailers: trailers}
	return resp, nil
}

type captureTrailersAtEOF struct {
	r        io.ReadCloser
	resp     *http.Response
	trailers *Trailers
}

func (c *captureTrailersAtEOF) Read(p []byte) (n int, err error) {
	n, err = c.r.Read(p)
	if errors.Is(err, io.EOF) {
		meta := c.resp.Trailer
		c.trailers.val.Store(&meta)
	}
	return n, err
}

func (c *captureTrailersAtEOF) Close() error {
	return c.r.Close()
}
