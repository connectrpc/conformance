// Copyright 2023-2024 The Connect Authors
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

package connectconformance

import (
	"fmt"
	"sort"
	"testing"

	conformancev1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
)

func TestParseConfig_ComputesPermutations(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name          string
		config        string
		expectedCases []configCase
	}{
		{
			name:   "default",
			config: "",
			// Compute permutations for all of the default supported features.
			expectedCases: excludeDisallowed(
				computePermutations(
					[]conformancev1.HTTPVersion{
						conformancev1.HTTPVersion_HTTP_VERSION_1,
						conformancev1.HTTPVersion_HTTP_VERSION_2,
					},
					[]conformancev1.Protocol{
						conformancev1.Protocol_PROTOCOL_CONNECT,
						conformancev1.Protocol_PROTOCOL_GRPC,
						conformancev1.Protocol_PROTOCOL_GRPC_WEB,
					},
					[]conformancev1.Codec{
						conformancev1.Codec_CODEC_PROTO,
						conformancev1.Codec_CODEC_JSON,
					},
					[]conformancev1.Compression{
						conformancev1.Compression_COMPRESSION_IDENTITY,
						conformancev1.Compression_COMPRESSION_GZIP,
					},
					[]conformancev1.StreamType{
						conformancev1.StreamType_STREAM_TYPE_UNARY,
						conformancev1.StreamType_STREAM_TYPE_CLIENT_STREAM,
						conformancev1.StreamType_STREAM_TYPE_SERVER_STREAM,
						conformancev1.StreamType_STREAM_TYPE_HALF_DUPLEX_BIDI_STREAM,
						conformancev1.StreamType_STREAM_TYPE_FULL_DUPLEX_BIDI_STREAM,
					},
					[]bool{true, false}, // Use TLS modes: default to supporting TLS
					[]bool{false},       // Use TLS client cert modes: default to NOT supporting TLS client certs
					[]bool{true, false}, // Use Connect GET modes: default to supporting GET
					[]bool{true, false}, // Use message receive limit modes: default to supporting limits
					[]conformancev1.TestSuite_ConnectVersionMode{
						conformancev1.TestSuite_CONNECT_VERSION_MODE_UNSPECIFIED,
						// default to not requiring version
						conformancev1.TestSuite_CONNECT_VERSION_MODE_IGNORE,
					},
				),
				true,  // default to supporting H2C
				false, // default to NOT supporting half-duplex bidi over HTTP 1
			),
		},
		{
			name: "simple features",
			config: `features:
                        protocols: [PROTOCOL_CONNECT]			# only Connect protocol
                        compressions: [COMPRESSION_IDENTITY]	# no compression
                        streamTypes: [STREAM_TYPE_UNARY,STREAM_TYPE_SERVER_STREAM]
                        supportsH2c: false
                        supportsTlsClientCerts: true
                        supportsHalfDuplexBidiOverHttp1: true
                        supportsMessageReceiveLimit: false
                        supportsConnectGet: false
                        requiresConnectVersionHeader: true`,
			expectedCases: excludeDisallowed(
				computePermutations(
					[]conformancev1.HTTPVersion{
						conformancev1.HTTPVersion_HTTP_VERSION_1,
						conformancev1.HTTPVersion_HTTP_VERSION_2,
					},
					[]conformancev1.Protocol{
						conformancev1.Protocol_PROTOCOL_CONNECT,
					},
					[]conformancev1.Codec{
						conformancev1.Codec_CODEC_PROTO,
						conformancev1.Codec_CODEC_JSON,
					},
					[]conformancev1.Compression{
						conformancev1.Compression_COMPRESSION_IDENTITY,
					},
					[]conformancev1.StreamType{
						conformancev1.StreamType_STREAM_TYPE_UNARY,
						conformancev1.StreamType_STREAM_TYPE_SERVER_STREAM,
					},
					[]bool{true, false},
					[]bool{true, false},
					[]bool{false},
					[]bool{false},
					[]conformancev1.TestSuite_ConnectVersionMode{
						conformancev1.TestSuite_CONNECT_VERSION_MODE_UNSPECIFIED,
						conformancev1.TestSuite_CONNECT_VERSION_MODE_REQUIRE,
					},
				),
				false,
				true,
			),
		},
		{
			name: "simple include cases",
			config: `
                      features:
                        versions: [HTTP_VERSION_1]
                        protocols: [PROTOCOL_CONNECT]
                        compressions: [COMPRESSION_IDENTITY]
                        streamTypes: [STREAM_TYPE_UNARY]
                        supportsH2c: false
                        supportsTls: false
                      include_cases:
                      - version: HTTP_VERSION_2
                        protocol: PROTOCOL_GRPC
                        codec: CODEC_PROTO
                        streamType: STREAM_TYPE_UNARY
                        useTls: true
                        useTlsClientCerts: true
                        useMessageReceiveLimit: true
                      - version: HTTP_VERSION_2
                        protocol: PROTOCOL_GRPC
                        codec: CODEC_PROTO
                        streamType: STREAM_TYPE_SERVER_STREAM
                        useTls: true
                        useMessageReceiveLimit: false`,
			expectedCases: union(
				excludeDisallowed(
					computePermutations(
						[]conformancev1.HTTPVersion{
							conformancev1.HTTPVersion_HTTP_VERSION_1,
						},
						[]conformancev1.Protocol{
							conformancev1.Protocol_PROTOCOL_CONNECT,
						},
						[]conformancev1.Codec{
							conformancev1.Codec_CODEC_PROTO,
							conformancev1.Codec_CODEC_JSON,
						},
						[]conformancev1.Compression{
							conformancev1.Compression_COMPRESSION_IDENTITY,
						},
						[]conformancev1.StreamType{
							conformancev1.StreamType_STREAM_TYPE_UNARY,
						},
						[]bool{false},       // no TLS
						[]bool{false},       // ... so no TLS client certs either
						[]bool{true, false}, // but Connect GET supported
						[]bool{true, false}, // supports message receive limit
						[]conformancev1.TestSuite_ConnectVersionMode{
							conformancev1.TestSuite_CONNECT_VERSION_MODE_UNSPECIFIED,
							conformancev1.TestSuite_CONNECT_VERSION_MODE_IGNORE,
						},
					),
					false,
					false,
				),
				[]configCase{
					{
						Version:                conformancev1.HTTPVersion_HTTP_VERSION_2,
						Protocol:               conformancev1.Protocol_PROTOCOL_GRPC,
						Codec:                  conformancev1.Codec_CODEC_PROTO,
						Compression:            conformancev1.Compression_COMPRESSION_IDENTITY,
						StreamType:             conformancev1.StreamType_STREAM_TYPE_UNARY,
						UseTLS:                 true,
						UseTLSClientCerts:      true,
						UseMessageReceiveLimit: true,
					},
					{
						Version:     conformancev1.HTTPVersion_HTTP_VERSION_2,
						Protocol:    conformancev1.Protocol_PROTOCOL_GRPC,
						Codec:       conformancev1.Codec_CODEC_PROTO,
						Compression: conformancev1.Compression_COMPRESSION_IDENTITY,
						StreamType:  conformancev1.StreamType_STREAM_TYPE_SERVER_STREAM,
						UseTLS:      true,
					},
				},
			),
		},
		{
			name: "expanded include case",
			config: `
                      features:
                        compressions: [COMPRESSION_IDENTITY]
                        supportsConnectGet: false
                        supportsTlsClientCerts: true
                      include_cases:
                      # Since HTTP versions and codecs not specified, this will be expanded
                      # to include all supported versions and codecs.
                      - protocol: PROTOCOL_CONNECT
                        compression: COMPRESSION_GZIP # support GZIP only for Connect unary
                        streamType: STREAM_TYPE_UNARY`,
			expectedCases: union(
				excludeDisallowed(
					computePermutations(
						[]conformancev1.HTTPVersion{
							conformancev1.HTTPVersion_HTTP_VERSION_1,
							conformancev1.HTTPVersion_HTTP_VERSION_2,
						},
						[]conformancev1.Protocol{
							conformancev1.Protocol_PROTOCOL_CONNECT,
							conformancev1.Protocol_PROTOCOL_GRPC,
							conformancev1.Protocol_PROTOCOL_GRPC_WEB,
						},
						[]conformancev1.Codec{
							conformancev1.Codec_CODEC_PROTO,
							conformancev1.Codec_CODEC_JSON,
						},
						[]conformancev1.Compression{
							conformancev1.Compression_COMPRESSION_IDENTITY,
						},
						[]conformancev1.StreamType{
							conformancev1.StreamType_STREAM_TYPE_UNARY,
							conformancev1.StreamType_STREAM_TYPE_CLIENT_STREAM,
							conformancev1.StreamType_STREAM_TYPE_SERVER_STREAM,
							conformancev1.StreamType_STREAM_TYPE_HALF_DUPLEX_BIDI_STREAM,
							conformancev1.StreamType_STREAM_TYPE_FULL_DUPLEX_BIDI_STREAM,
						},
						[]bool{true, false}, // TLS supported
						[]bool{true, false}, // TLS client certs supported
						[]bool{false},       // but Connect GET is not
						[]bool{true, false}, // message receive limits supported
						[]conformancev1.TestSuite_ConnectVersionMode{
							conformancev1.TestSuite_CONNECT_VERSION_MODE_UNSPECIFIED,
							conformancev1.TestSuite_CONNECT_VERSION_MODE_IGNORE,
						},
					),
					true,
					false,
				),
				excludeDisallowed(
					computePermutations(
						[]conformancev1.HTTPVersion{
							conformancev1.HTTPVersion_HTTP_VERSION_1,
							conformancev1.HTTPVersion_HTTP_VERSION_2,
						},
						[]conformancev1.Protocol{
							conformancev1.Protocol_PROTOCOL_CONNECT,
						},
						[]conformancev1.Codec{
							conformancev1.Codec_CODEC_PROTO,
							conformancev1.Codec_CODEC_JSON,
						},
						[]conformancev1.Compression{
							conformancev1.Compression_COMPRESSION_GZIP,
						},
						[]conformancev1.StreamType{
							conformancev1.StreamType_STREAM_TYPE_UNARY,
						},
						[]bool{true, false}, // TLS supported
						[]bool{true, false}, // TLS client certs supported
						[]bool{false},       // but Connect GET is not
						[]bool{true, false}, // message receive limits supported
						[]conformancev1.TestSuite_ConnectVersionMode{
							conformancev1.TestSuite_CONNECT_VERSION_MODE_UNSPECIFIED,
							conformancev1.TestSuite_CONNECT_VERSION_MODE_IGNORE,
						},
					),
					true,
					false,
				),
			),
		},
		{
			name: "exclude cases",
			config: `
                      features:
                        compressions: [COMPRESSION_IDENTITY]
                      exclude_cases:
                      - protocol: PROTOCOL_CONNECT  # Connect unary not yet implemented
                        streamType: STREAM_TYPE_UNARY`,
			expectedCases: minus(
				excludeDisallowed(
					computePermutations(
						[]conformancev1.HTTPVersion{
							conformancev1.HTTPVersion_HTTP_VERSION_1,
							conformancev1.HTTPVersion_HTTP_VERSION_2,
						},
						[]conformancev1.Protocol{
							conformancev1.Protocol_PROTOCOL_CONNECT,
							conformancev1.Protocol_PROTOCOL_GRPC,
							conformancev1.Protocol_PROTOCOL_GRPC_WEB,
						},
						[]conformancev1.Codec{
							conformancev1.Codec_CODEC_PROTO,
							conformancev1.Codec_CODEC_JSON,
						},
						[]conformancev1.Compression{
							conformancev1.Compression_COMPRESSION_IDENTITY,
						},
						[]conformancev1.StreamType{
							conformancev1.StreamType_STREAM_TYPE_UNARY,
							conformancev1.StreamType_STREAM_TYPE_CLIENT_STREAM,
							conformancev1.StreamType_STREAM_TYPE_SERVER_STREAM,
							conformancev1.StreamType_STREAM_TYPE_HALF_DUPLEX_BIDI_STREAM,
							conformancev1.StreamType_STREAM_TYPE_FULL_DUPLEX_BIDI_STREAM,
						},
						[]bool{true, false},
						[]bool{false},
						[]bool{true, false},
						[]bool{true, false},
						[]conformancev1.TestSuite_ConnectVersionMode{
							conformancev1.TestSuite_CONNECT_VERSION_MODE_UNSPECIFIED,
							conformancev1.TestSuite_CONNECT_VERSION_MODE_IGNORE,
						},
					),
					true,
					false,
				),
				excludeDisallowed(
					computePermutations(
						[]conformancev1.HTTPVersion{
							conformancev1.HTTPVersion_HTTP_VERSION_1,
							conformancev1.HTTPVersion_HTTP_VERSION_2,
						},
						[]conformancev1.Protocol{
							conformancev1.Protocol_PROTOCOL_CONNECT,
						},
						[]conformancev1.Codec{
							conformancev1.Codec_CODEC_PROTO,
							conformancev1.Codec_CODEC_JSON,
						},
						[]conformancev1.Compression{
							conformancev1.Compression_COMPRESSION_IDENTITY,
						},
						[]conformancev1.StreamType{
							conformancev1.StreamType_STREAM_TYPE_UNARY,
						},
						[]bool{true, false},
						[]bool{false},
						[]bool{true, false},
						[]bool{true, false},
						[]conformancev1.TestSuite_ConnectVersionMode{
							conformancev1.TestSuite_CONNECT_VERSION_MODE_UNSPECIFIED,
							conformancev1.TestSuite_CONNECT_VERSION_MODE_IGNORE,
						},
					),
					true,
					false,
				),
			),
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			cases, err := parseConfig("config.yaml", []byte(testCase.config))
			require.NoError(t, err)
			sortCases(cases)
			sortCases(testCase.expectedCases)
			require.Empty(t, cmp.Diff(testCase.expectedCases, cases), "- wanted; + got")
		})
	}
}

func TestParseConfig_RejectsInvalidConfigurations(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name        string
		config      string
		expectedErr string
	}{
		{
			name: "features: HTTP/3 without TLS",
			config: `features:
                        versions: [HTTP_VERSION_1, HTTP_VERSION_2, HTTP_VERSION_3]
                        supportsTls: false`,
			expectedErr: "config features indicate HTTP/3 is supported but TLS is not",
		},
		{
			name: "features: HTTP/2 without TLS or H2C",
			config: `features:
                        versions: [HTTP_VERSION_1, HTTP_VERSION_2]
                        supportsH2c: false
                        supportsTls: false`,
			expectedErr: "config features indicate HTTP/2 is supported but neither H2C nor TLS are supported",
		},
		{
			name: "features: gRPC without HTTP/2",
			config: `features:
                        versions: [HTTP_VERSION_1]
                        protocols: [PROTOCOL_CONNECT,PROTOCOL_GRPC,PROTOCOL_GRPC_WEB]`,
			expectedErr: "config features indicate gRPC protocol is supported but HTTP/2 is not",
		},
		{
			name: "features: gRPC without trailers",
			config: `features:
                        protocols: [PROTOCOL_CONNECT,PROTOCOL_GRPC,PROTOCOL_GRPC_WEB]
                        supportsTrailers: false`,
			expectedErr: "config features indicate gRPC protocol is supported but trailers are not",
		},
		{
			name: "features: H2C without HTTP/2",
			config: `features:
                        versions: [HTTP_VERSION_1]
                        supportsH2c: true`,
			expectedErr: "config features indicate H2C is supported but HTTP/2 is not a supported HTTP version",
		},
		{
			name: "features: half-duplex with only HTTP/1",
			config: `features:
                        versions: [HTTP_VERSION_1]
                        streamTypes:
                        - STREAM_TYPE_UNARY
                        - STREAM_TYPE_CLIENT_STREAM
                        - STREAM_TYPE_SERVER_STREAM
                        - STREAM_TYPE_HALF_DUPLEX_BIDI_STREAM
                        supportsHalfDuplexBidiOverHttp1: false`,
			expectedErr: "config features indicate half-duplex bidi streams are supported but not over HTTP/1.1, and neither HTTP/2 nor HTTP/3 included",
		},
		{
			name: "features: full-duplex with only HTTP/1",
			config: `features:
                        versions: [HTTP_VERSION_1]
                        streamTypes:
                        - STREAM_TYPE_UNARY
                        - STREAM_TYPE_CLIENT_STREAM
                        - STREAM_TYPE_SERVER_STREAM
                        - STREAM_TYPE_HALF_DUPLEX_BIDI_STREAM
                        - STREAM_TYPE_FULL_DUPLEX_BIDI_STREAM
                        supportsHalfDuplexBidiOverHttp1: true`,
			expectedErr: "config features indicate full-duplex bidi streams are supported but neither HTTP/2 nor HTTP/3 included",
		},
		{
			name: "features: TLS client certs without TLS",
			config: `features:
                        supportsTlsClientCerts: true
                        supportsTls: false`,
			expectedErr: "config features indicate TLS client certs are supported but not TLS",
		},
		{
			name: "included case: HTTP/3 without TLS",
			config: `
                     features:
                        supportsTls: false
                     include_cases:
                      - version: HTTP_VERSION_3
                        protocol: PROTOCOL_CONNECT`,
			expectedErr: "config case indicates HTTP/3 but not TLS",
		},
		{
			name: "included case: HTTP/3 without TLS (b)",
			config: `
                     features:
                        supportsTls: true
                     include_cases:
                      - version: HTTP_VERSION_3
                        protocol: PROTOCOL_CONNECT
                        useTls: false`,
			expectedErr: "config case indicates HTTP/3 but not TLS",
		},
		{
			name: "included case: HTTP/2 without TLS or H2C",
			config: `
                     features:
                        supportsTls: false
                        supportsH2c: false
                     include_cases:
                      - version: HTTP_VERSION_2
                        protocol: PROTOCOL_GRPC`,
			expectedErr: "config case indicates HTTP/2 but not TLS, and features indicate that H2C not supported",
		},
		{
			name: "included case: HTTP/2 without TLS or H2C (b)",
			config: `
                     features:
                        supportsTls: true
                        supportsH2c: false
                     include_cases:
                      - version: HTTP_VERSION_2
                        protocol: PROTOCOL_GRPC
                        useTls: false`,
			expectedErr: "config case indicates HTTP/2 but not TLS, and features indicate that H2C not supported",
		},
		{
			name: "included case: gRPC without HTTP/2",
			config: `
                     features:
                        versions: [HTTP_VERSION_1]
                        protocols: [PROTOCOL_CONNECT, PROTOCOL_GRPC_WEB]
                     include_cases:
                      - protocol: PROTOCOL_GRPC`,
			expectedErr: "config case indicates gRPC protocol but not HTTP/2",
		},
		{
			name: "included case: gRPC without HTTP/2 (b)",
			config: `
                     features:
                        versions: [HTTP_VERSION_1, HTTP_VERSION_2]
                        protocols: [PROTOCOL_CONNECT, PROTOCOL_GRPC_WEB]
                     include_cases:
                      - version: HTTP_VERSION_1
                        protocol: PROTOCOL_GRPC`,
			expectedErr: "config case indicates gRPC protocol but not HTTP/2",
		},
		{
			name: "included case: half-duplex with only HTTP/1",
			config: `
                     features:
                        versions: [HTTP_VERSION_1]
                        streamTypes: [STREAM_TYPE_UNARY, STREAM_TYPE_CLIENT_STREAM, STREAM_TYPE_SERVER_STREAM]
                     include_cases:
                      - streamType: STREAM_TYPE_HALF_DUPLEX_BIDI_STREAM`,
			expectedErr: "config case indicates half-duplex bidi stream type, but features indicate only HTTP/1.1 and that half-duplex is not supported over HTTP1.1",
		},
		{
			name: "included case: half-duplex with only HTTP/1 (b)",
			config: `
                     features:
                        streamTypes: [STREAM_TYPE_UNARY, STREAM_TYPE_CLIENT_STREAM, STREAM_TYPE_SERVER_STREAM]
                     include_cases:
                      - version: HTTP_VERSION_1
                        streamType: STREAM_TYPE_HALF_DUPLEX_BIDI_STREAM`,
			expectedErr: "config case indicates half-duplex bidi stream type, but features indicate only HTTP/1.1 and that half-duplex is not supported over HTTP1.1",
		},
		{
			name: "included case: full-duplex with only HTTP/1",
			config: `
                     features:
                        versions: [HTTP_VERSION_1]
                        streamTypes: [STREAM_TYPE_UNARY, STREAM_TYPE_CLIENT_STREAM, STREAM_TYPE_SERVER_STREAM, STREAM_TYPE_HALF_DUPLEX_BIDI_STREAM]
                        supportsHalfDuplexBidiOverHttp1: true
                     include_cases:
                      - streamType: STREAM_TYPE_FULL_DUPLEX_BIDI_STREAM`,
			expectedErr: "config case indicates full-duplex bidi stream type, but features indicate only HTTP/1.1 which cannot support full-duplex",
		},
		{
			name: "included case: full-duplex with only HTTP/1 (b)",
			config: `
                     features:
                        streamTypes: [STREAM_TYPE_UNARY, STREAM_TYPE_CLIENT_STREAM, STREAM_TYPE_SERVER_STREAM, STREAM_TYPE_HALF_DUPLEX_BIDI_STREAM]
                        supportsHalfDuplexBidiOverHttp1: true
                     include_cases:
                      - version: HTTP_VERSION_1
                        streamType: STREAM_TYPE_FULL_DUPLEX_BIDI_STREAM`,
			expectedErr: "config case indicates full-duplex bidi stream type, but features indicate only HTTP/1.1 which cannot support full-duplex",
		},
		{
			name: "included case: TLS client certs without TLS (a)",
			config: `
                     features:
                     include_cases:
                      - useTlsClientCerts: true
                        useTls: false`,
			expectedErr: "config case indicates use of TLS client certs but also indicates NOT using TLS",
		},
		{
			name: "included case: TLS client certs without TLS (a)",
			config: `
                     features:
                        supportsTls: false
                     include_cases:
                      - useTlsClientCerts: true`,
			expectedErr: "config case indicates use of TLS client certs but TLS is not supported",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			_, err := parseConfig("config.yaml", []byte(testCase.config))
			require.ErrorContains(t, err, testCase.expectedErr)
			t.Log(err)
		})
	}
}

func computePermutations(
	versions []conformancev1.HTTPVersion,
	protocols []conformancev1.Protocol,
	codecs []conformancev1.Codec,
	compressions []conformancev1.Compression,
	streamTypes []conformancev1.StreamType,
	useTLSOptions []bool,
	useTLSClientCertOptions []bool,
	useConnectGETOptions []bool,
	useMaxRecvLimitOptions []bool,
	connectVersionModes []conformancev1.TestSuite_ConnectVersionMode,
) []configCase {
	size := len(versions) * len(protocols) * len(codecs) * len(compressions) * len(streamTypes) * len(useTLSOptions) * len(useConnectGETOptions) * len(connectVersionModes)
	results := make([]configCase, 0, size)
	for _, version := range versions {
		for _, protocol := range protocols {
			for _, codec := range codecs {
				for _, compression := range compressions {
					for _, streamType := range streamTypes {
						for _, useTLS := range useTLSOptions {
							for _, useTLSClientCerts := range useTLSClientCertOptions {
								for _, useConnectGET := range useConnectGETOptions {
									for _, useMaxRecvLimit := range useMaxRecvLimitOptions {
										for _, connectVersionMode := range connectVersionModes {
											results = append(results, configCase{
												Version:                version,
												Protocol:               protocol,
												Codec:                  codec,
												Compression:            compression,
												StreamType:             streamType,
												UseTLS:                 useTLS,
												UseTLSClientCerts:      useTLSClientCerts,
												UseConnectGET:          useConnectGET,
												UseMessageReceiveLimit: useMaxRecvLimit,
												ConnectVersionMode:     connectVersionMode,
											})
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
	return results
}

func excludeDisallowed(cases []configCase, supportsH2C, supportsHalfDuplexBidiHTTP1 bool) []configCase {
	disallowed := map[int]struct{}{}
	for i, cfgCase := range cases {
		switch {
		case !cfgCase.UseTLS && cfgCase.Version == conformancev1.HTTPVersion_HTTP_VERSION_3:
			// can't use HTTP/3 w/out TLS
			disallowed[i] = struct{}{}
		case !cfgCase.UseTLS && cfgCase.Version == conformancev1.HTTPVersion_HTTP_VERSION_2 && !supportsH2C:
			// can't use HTTP/2 w/out TLS unless H2C is supported
			disallowed[i] = struct{}{}
		case !cfgCase.UseTLS && cfgCase.UseTLSClientCerts:
			// can't use client certs w/out TLS
			disallowed[i] = struct{}{}
		case cfgCase.Protocol == conformancev1.Protocol_PROTOCOL_GRPC && cfgCase.Version != conformancev1.HTTPVersion_HTTP_VERSION_2:
			// can't use gRPC w/out HTTP/2
			disallowed[i] = struct{}{}
		case cfgCase.UseConnectGET && cfgCase.Protocol != conformancev1.Protocol_PROTOCOL_CONNECT:
			// GET is only for the Connect protocol
			disallowed[i] = struct{}{}
		case cfgCase.StreamType == conformancev1.StreamType_STREAM_TYPE_FULL_DUPLEX_BIDI_STREAM && cfgCase.Version == conformancev1.HTTPVersion_HTTP_VERSION_1:
			// Can't do full-duplex streams w/ HTTP 1.1
			disallowed[i] = struct{}{}
		case cfgCase.StreamType == conformancev1.StreamType_STREAM_TYPE_HALF_DUPLEX_BIDI_STREAM && cfgCase.Version == conformancev1.HTTPVersion_HTTP_VERSION_1 && !supportsHalfDuplexBidiHTTP1:
			// Can't do half-duplex streams w/ HTTP 1.1 either if impl doesn't support that
			disallowed[i] = struct{}{}
		case cfgCase.ConnectVersionMode != conformancev1.TestSuite_CONNECT_VERSION_MODE_UNSPECIFIED && cfgCase.Protocol != conformancev1.Protocol_PROTOCOL_CONNECT:
			// Connect version requirement only applies to Connect protocol
			disallowed[i] = struct{}{}
		default:
		}
	}
	if len(disallowed) == 0 {
		return cases
	}
	filtered := make([]configCase, len(cases)-len(disallowed))
	j := 0 //nolint: varnamelen
	for i := range cases {
		if _, ok := disallowed[i]; ok {
			continue
		}
		filtered[j] = cases[i]
		j++
	}
	return filtered
}

func sortCases(cases []configCase) {
	sort.Slice(cases, func(i, j int) bool {
		return fmt.Sprintf("%#v", cases[i]) < fmt.Sprintf("%#v", cases[j])
	})
}

func union[T comparable](a, b []T) []T {
	set := make(map[T]struct{}, len(a)+len(b))
	for _, elem := range a {
		set[elem] = struct{}{}
	}
	for _, elem := range b {
		set[elem] = struct{}{}
	}
	results := make([]T, 0, len(set))
	for elem := range set {
		results = append(results, elem)
	}
	return results
}

func minus[T comparable](a, b []T) []T {
	set := make(map[T]struct{}, len(a)+len(b))
	for _, elem := range a {
		set[elem] = struct{}{}
	}
	for _, elem := range b {
		delete(set, elem)
	}
	results := make([]T, 0, len(set))
	for elem := range set {
		results = append(results, elem)
	}
	return results
}
