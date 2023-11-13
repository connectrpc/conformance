// Copyright 2023 The Connect Authors
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
	"errors"
	"fmt"

	conformancev1alpha1 "connectrpc.com/conformance/internal/gen/proto/go/connectrpc/conformance/v1alpha1"
	"github.com/bufbuild/protoyaml-go"
)

// configCase is a resolved configuration case. This mirrors the
// ConfigCase protobuf message, but includes additional de-normalized
// fields.
type configCase struct {
	Version                conformancev1alpha1.HTTPVersion
	Protocol               conformancev1alpha1.Protocol
	Codec                  conformancev1alpha1.Codec
	Compression            conformancev1alpha1.Compression
	StreamType             conformancev1alpha1.StreamType
	UseTLS                 bool
	UseTLSClientCerts      bool
	UseConnectGET          bool
	UseMessageReceiveLimit bool
	ConnectVersionMode     conformancev1alpha1.TestSuite_ConnectVersionMode
}

// supportedFeatures is a resolved set of features. This mirrors
// the Features protobuf message, but without pointers/optional
// values.
type supportedFeatures struct {
	Versions                        []conformancev1alpha1.HTTPVersion
	Protocols                       []conformancev1alpha1.Protocol
	Codecs                          []conformancev1alpha1.Codec
	Compressions                    []conformancev1alpha1.Compression
	StreamTypes                     []conformancev1alpha1.StreamType
	SupportsH2C                     bool
	SupportsTLS                     bool
	SupportsTLSClientCerts          bool
	SupportsTrailers                bool
	SupportsHalfDuplexBidiOverHTTP1 bool
	SupportsConnectGet              bool
	SupportsMessageReceiveLimit     bool
	RequiresConnectVersionHeader    bool
}

// parseConfig loads all config cases from the given file name. If the given
// file name is blank, it returns all config cases based on default features.
func parseConfig(configFileName string, data []byte) ([]configCase, error) {
	var config conformancev1alpha1.Config
	if len(data) > 0 {
		opts := protoyaml.UnmarshalOptions{
			Path: configFileName,
		}
		if err := opts.Unmarshal(data, &config); err != nil {
			return nil, ensureFileName(err, configFileName)
		}
	}
	if config.Features == nil {
		config.Features = &conformancev1alpha1.Features{}
	}
	features, err := resolveFeatures(config.Features)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", configFileName, err)
	}
	cases := computeCasesFromFeatures(features, nil, nil, nil)
	for i, includeCase := range config.IncludeCases {
		resolvedIncludes, err := resolveCase(features, includeCase)
		if err != nil {
			return nil, fmt.Errorf("%s: include case #%d: %w", configFileName, i+1, err)
		}
		for include := range resolvedIncludes {
			cases[include] = struct{}{}
		}
	}
	for i, excludeCase := range config.ExcludeCases {
		resolvedExcludes, err := resolveCase(features, excludeCase)
		if err != nil {
			return nil, fmt.Errorf("%s: exclude case #%d: %w", configFileName, i+1, err)
		}
		for exclude := range resolvedExcludes {
			delete(cases, exclude)
		}
	}
	if len(cases) == 0 {
		return nil, fmt.Errorf("%s: configuration resulted in zero cases to test", configFileName)
	}
	casesSlice := make([]configCase, 0, len(cases))
	for c := range cases {
		casesSlice = append(casesSlice, c)
	}
	return casesSlice, nil
}

// resolveFeatures resolves all unspecified fields in the given features from the
// config file. It returns an error if the given features are invalid due to
// impossible or contradictory settings.
func resolveFeatures(features *conformancev1alpha1.Features) (supportedFeatures, error) { //nolint:gocyclo
	result := supportedFeatures{
		Versions:                        features.Versions,
		Protocols:                       features.Protocols,
		Codecs:                          features.Codecs,
		Compressions:                    features.Compressions,
		StreamTypes:                     features.StreamTypes,
		SupportsH2C:                     features.GetSupportsH2C(),
		SupportsTLS:                     features.GetSupportsTls(),
		SupportsTLSClientCerts:          features.GetSupportsTlsClientCerts(),
		SupportsTrailers:                features.GetSupportsTrailers(),
		SupportsHalfDuplexBidiOverHTTP1: features.GetSupportsHalfDuplexBidiOverHttp1(),
		SupportsConnectGet:              features.GetSupportsConnectGet(),
		SupportsMessageReceiveLimit:     features.GetSupportsMessageReceiveLimit(),
		RequiresConnectVersionHeader:    features.GetRequiresConnectVersionHeader(),
	}

	// These flags should default to true if not provided
	if features.SupportsH2C == nil {
		result.SupportsH2C = true
	}
	if features.SupportsTls == nil {
		result.SupportsTLS = true
	}
	if features.SupportsTrailers == nil {
		result.SupportsTrailers = true
	}
	if features.SupportsConnectGet == nil {
		result.SupportsConnectGet = true
	}
	if features.SupportsMessageReceiveLimit == nil {
		result.SupportsMessageReceiveLimit = true
	}

	if result.SupportsTLSClientCerts && !result.SupportsTLS {
		return result, errors.New("config features indicate TLS client certs are supported but not TLS")
	}

	if len(result.Versions) == 0 {
		if result.SupportsTLS || result.SupportsH2C {
			result.Versions = []conformancev1alpha1.HTTPVersion{
				conformancev1alpha1.HTTPVersion_HTTP_VERSION_1,
				conformancev1alpha1.HTTPVersion_HTTP_VERSION_2,
			}
		} else {
			result.Versions = []conformancev1alpha1.HTTPVersion{
				conformancev1alpha1.HTTPVersion_HTTP_VERSION_1,
			}
		}
	} else if features.SupportsH2C != nil && features.GetSupportsH2C() && !contains(result.Versions, conformancev1alpha1.HTTPVersion_HTTP_VERSION_2) {
		return result, errors.New("config features indicate H2C is supported but HTTP/2 is not a supported HTTP version")
	}

	includesHTTP3 := contains(result.Versions, conformancev1alpha1.HTTPVersion_HTTP_VERSION_3)
	if includesHTTP3 && !result.SupportsTLS {
		return result, errors.New("config features indicate HTTP/3 is supported but TLS is not")
	}
	includesHTTP2 := contains(result.Versions, conformancev1alpha1.HTTPVersion_HTTP_VERSION_2)
	canUseHTTP2 := result.SupportsH2C || result.SupportsTLS
	if includesHTTP2 && !canUseHTTP2 {
		return result, errors.New("config features indicate HTTP/2 is supported but neither H2C nor TLS are supported")
	}
	if len(result.Versions) == 0 {
		if canUseHTTP2 {
			includesHTTP2 = true
			result.Versions = []conformancev1alpha1.HTTPVersion{
				conformancev1alpha1.HTTPVersion_HTTP_VERSION_1,
				conformancev1alpha1.HTTPVersion_HTTP_VERSION_2,
			}
		} else {
			result.Versions = []conformancev1alpha1.HTTPVersion{conformancev1alpha1.HTTPVersion_HTTP_VERSION_1}
		}
	}

	includesGPRC := contains(result.Protocols, conformancev1alpha1.Protocol_PROTOCOL_GRPC)
	if includesGPRC && !result.SupportsTrailers {
		return result, errors.New("config features indicate gRPC protocol is supported but trailers are not")
	}
	if includesGPRC && !includesHTTP2 {
		return result, errors.New("config features indicate gRPC protocol is supported but HTTP/2 is not")
	}
	canUseGRPC := result.SupportsTrailers && includesHTTP2
	if len(result.Protocols) == 0 {
		if canUseGRPC {
			result.Protocols = []conformancev1alpha1.Protocol{
				conformancev1alpha1.Protocol_PROTOCOL_CONNECT,
				conformancev1alpha1.Protocol_PROTOCOL_GRPC,
				conformancev1alpha1.Protocol_PROTOCOL_GRPC_WEB,
			}
		} else {
			result.Protocols = []conformancev1alpha1.Protocol{
				conformancev1alpha1.Protocol_PROTOCOL_CONNECT,
				conformancev1alpha1.Protocol_PROTOCOL_GRPC_WEB,
			}
		}
	}

	if len(result.Codecs) == 0 {
		result.Codecs = []conformancev1alpha1.Codec{
			conformancev1alpha1.Codec_CODEC_PROTO,
			conformancev1alpha1.Codec_CODEC_JSON,
		}
	}
	if len(result.Compressions) == 0 {
		result.Compressions = []conformancev1alpha1.Compression{
			conformancev1alpha1.Compression_COMPRESSION_IDENTITY,
			conformancev1alpha1.Compression_COMPRESSION_GZIP,
		}
	}

	includesFullDuplex := contains(result.StreamTypes, conformancev1alpha1.StreamType_STREAM_TYPE_FULL_DUPLEX_BIDI_STREAM)
	onlyHTTP1 := !includesHTTP2 && !includesHTTP3
	if includesFullDuplex && onlyHTTP1 {
		return result, errors.New("config features indicate full-duplex bidi streams are supported but neither HTTP/2 nor HTTP/3 included")
	}
	includesHalfDuplex := contains(result.StreamTypes, conformancev1alpha1.StreamType_STREAM_TYPE_HALF_DUPLEX_BIDI_STREAM)
	if includesHalfDuplex && onlyHTTP1 && !result.SupportsHalfDuplexBidiOverHTTP1 {
		return result, errors.New("config features indicate half-duplex bidi streams are supported but not over HTTP/1.1, and neither HTTP/2 nor HTTP/3 included")
	}
	if len(result.StreamTypes) == 0 { //nolint:nestif
		if onlyHTTP1 {
			if result.SupportsHalfDuplexBidiOverHTTP1 {
				result.StreamTypes = []conformancev1alpha1.StreamType{
					conformancev1alpha1.StreamType_STREAM_TYPE_UNARY,
					conformancev1alpha1.StreamType_STREAM_TYPE_CLIENT_STREAM,
					conformancev1alpha1.StreamType_STREAM_TYPE_SERVER_STREAM,
					conformancev1alpha1.StreamType_STREAM_TYPE_HALF_DUPLEX_BIDI_STREAM,
				}
			} else {
				result.StreamTypes = []conformancev1alpha1.StreamType{
					conformancev1alpha1.StreamType_STREAM_TYPE_UNARY,
					conformancev1alpha1.StreamType_STREAM_TYPE_CLIENT_STREAM,
					conformancev1alpha1.StreamType_STREAM_TYPE_SERVER_STREAM,
				}
			}
		} else {
			result.StreamTypes = []conformancev1alpha1.StreamType{
				conformancev1alpha1.StreamType_STREAM_TYPE_UNARY,
				conformancev1alpha1.StreamType_STREAM_TYPE_CLIENT_STREAM,
				conformancev1alpha1.StreamType_STREAM_TYPE_SERVER_STREAM,
				conformancev1alpha1.StreamType_STREAM_TYPE_HALF_DUPLEX_BIDI_STREAM,
				conformancev1alpha1.StreamType_STREAM_TYPE_FULL_DUPLEX_BIDI_STREAM,
			}
		}
	}

	return result, nil
}

// computeCasesFromFeatures expands the given features into all matching config
// permutations.
func computeCasesFromFeatures(features supportedFeatures, tlsCases, tlsClientCertCases, msgRecvLimitCases []bool) map[configCase]struct{} {
	// if tlsCases, tlsClientCertCases, and msgRecvLimitCases not explicitly provided, derive them from features
	if len(tlsCases) == 0 {
		if features.SupportsTLS {
			tlsCases = []bool{false, true}
		} else {
			tlsCases = []bool{false}
		}
	}
	if len(tlsClientCertCases) == 0 {
		if features.SupportsTLSClientCerts {
			tlsClientCertCases = []bool{false, true}
		} else {
			tlsClientCertCases = []bool{false}
		}
	}
	if len(msgRecvLimitCases) == 0 {
		if features.SupportsMessageReceiveLimit {
			msgRecvLimitCases = []bool{false, true}
		} else {
			msgRecvLimitCases = []bool{false}
		}
	}
	cases := map[configCase]struct{}{}
	for _, version := range features.Versions {
		for _, tlsCase := range tlsCases {
			if !tlsCase &&
				(version == conformancev1alpha1.HTTPVersion_HTTP_VERSION_3 ||
					(version == conformancev1alpha1.HTTPVersion_HTTP_VERSION_2 && !features.SupportsH2C)) {
				continue // TLS required
			}
			for _, tlsClientCertCase := range tlsClientCertCases {
				if tlsClientCertCase && !tlsCase {
					// can't use client certs w/out TLS
					continue
				}
				for _, protocol := range features.Protocols {
					if protocol == conformancev1alpha1.Protocol_PROTOCOL_GRPC &&
						version != conformancev1alpha1.HTTPVersion_HTTP_VERSION_2 {
						continue // gRPC requires HTTP/2
					}

					connectGetCases := []bool{false}
					if protocol == conformancev1alpha1.Protocol_PROTOCOL_CONNECT && features.SupportsConnectGet {
						connectGetCases = []bool{false, true}
					}
					validateConnectVersionCases := []conformancev1alpha1.TestSuite_ConnectVersionMode{conformancev1alpha1.TestSuite_CONNECT_VERSION_MODE_UNSPECIFIED}
					if protocol == conformancev1alpha1.Protocol_PROTOCOL_CONNECT {
						if features.RequiresConnectVersionHeader {
							validateConnectVersionCases = append(validateConnectVersionCases, conformancev1alpha1.TestSuite_CONNECT_VERSION_MODE_REQUIRE)
						} else {
							validateConnectVersionCases = append(validateConnectVersionCases, conformancev1alpha1.TestSuite_CONNECT_VERSION_MODE_IGNORE)
						}
					}

					for _, streamType := range features.StreamTypes {
						switch streamType { //nolint:exhaustive
						case conformancev1alpha1.StreamType_STREAM_TYPE_HALF_DUPLEX_BIDI_STREAM:
							if !features.SupportsHalfDuplexBidiOverHTTP1 && version == conformancev1alpha1.HTTPVersion_HTTP_VERSION_1 {
								continue
							}
						case conformancev1alpha1.StreamType_STREAM_TYPE_FULL_DUPLEX_BIDI_STREAM:
							if version == conformancev1alpha1.HTTPVersion_HTTP_VERSION_1 {
								continue // HTTP/1.1 can't do full duplex
							}
						}

						for _, codec := range features.Codecs {
							for _, compression := range features.Compressions {
								for _, connectGetCase := range connectGetCases {
									for _, validateConnectVersionCase := range validateConnectVersionCases {
										for _, msgRecvLimitCase := range msgRecvLimitCases {
											cases[configCase{
												Version:                version,
												Protocol:               protocol,
												Codec:                  codec,
												Compression:            compression,
												StreamType:             streamType,
												UseTLS:                 tlsCase,
												UseTLSClientCerts:      tlsClientCertCase,
												UseConnectGET:          connectGetCase,
												UseMessageReceiveLimit: msgRecvLimitCase,
												ConnectVersionMode:     validateConnectVersionCase,
											}] = struct{}{}
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
	return cases
}

// resolveCase resolves all unspecified fields in the given config case from
// the config file. It returns an error if the given case is invalid due to
// impossible or contradictory settings. A single config case in the config
// file can map to numerous permutations if a field is omitted. For example,
// if the codecs field is unset/empty, the case is expanded to include all
// codecs indicated by the configured features.
func resolveCase(features supportedFeatures, unresolvedCase *conformancev1alpha1.ConfigCase) (map[configCase]struct{}, error) {
	// Build a set of supportedFeatures that matches the given ConfigCase.
	impliedFeatures := features // start as copy of supported features
	if unresolvedCase.Version != conformancev1alpha1.HTTPVersion_HTTP_VERSION_UNSPECIFIED {
		usingTLS := (unresolvedCase.UseTls != nil && unresolvedCase.GetUseTls()) ||
			(unresolvedCase.UseTls == nil && features.SupportsTLS)
		switch unresolvedCase.Version { //nolint:exhaustive
		case conformancev1alpha1.HTTPVersion_HTTP_VERSION_2:
			if !usingTLS && !features.SupportsH2C {
				return nil, errors.New("config case indicates HTTP/2 but not TLS, and features indicate that H2C not supported")
			}
		case conformancev1alpha1.HTTPVersion_HTTP_VERSION_3:
			if !usingTLS {
				return nil, errors.New("config case indicates HTTP/3 but not TLS")
			}
		}
		impliedFeatures.Versions = []conformancev1alpha1.HTTPVersion{unresolvedCase.Version}
	}
	if unresolvedCase.Protocol != conformancev1alpha1.Protocol_PROTOCOL_UNSPECIFIED {
		if unresolvedCase.Protocol == conformancev1alpha1.Protocol_PROTOCOL_GRPC &&
			!contains(impliedFeatures.Versions, conformancev1alpha1.HTTPVersion_HTTP_VERSION_2) {
			return nil, errors.New("config case indicates gRPC protocol but not HTTP/2")
		}
		impliedFeatures.Protocols = []conformancev1alpha1.Protocol{unresolvedCase.Protocol}
	}
	if unresolvedCase.Codec != conformancev1alpha1.Codec_CODEC_UNSPECIFIED {
		impliedFeatures.Codecs = []conformancev1alpha1.Codec{unresolvedCase.Codec}
	}
	if unresolvedCase.Compression != conformancev1alpha1.Compression_COMPRESSION_UNSPECIFIED {
		impliedFeatures.Compressions = []conformancev1alpha1.Compression{unresolvedCase.Compression}
	}
	if unresolvedCase.StreamType != conformancev1alpha1.StreamType_STREAM_TYPE_UNSPECIFIED {
		switch unresolvedCase.StreamType { //nolint:exhaustive
		case conformancev1alpha1.StreamType_STREAM_TYPE_HALF_DUPLEX_BIDI_STREAM:
			if !features.SupportsHalfDuplexBidiOverHTTP1 && only(impliedFeatures.Versions, conformancev1alpha1.HTTPVersion_HTTP_VERSION_1) {
				return nil, errors.New("config case indicates half-duplex bidi stream type, but features indicate only HTTP/1.1 and that half-duplex is not supported over HTTP1.1")
			}
		case conformancev1alpha1.StreamType_STREAM_TYPE_FULL_DUPLEX_BIDI_STREAM:
			if only(impliedFeatures.Versions, conformancev1alpha1.HTTPVersion_HTTP_VERSION_1) {
				return nil, errors.New("config case indicates full-duplex bidi stream type, but features indicate only HTTP/1.1 which cannot support full-duplex")
			}
		}
		impliedFeatures.StreamTypes = []conformancev1alpha1.StreamType{unresolvedCase.StreamType}
	}
	var tlsCases, tlsClientCertCases, msgReceiveLimitCases []bool
	if unresolvedCase.UseTls != nil {
		tlsCases = []bool{unresolvedCase.GetUseTls()}
	}
	if unresolvedCase.UseTlsClientCerts != nil {
		if unresolvedCase.UseTls != nil && !unresolvedCase.GetUseTls() {
			// use_tls explicitly set to false for this case?
			return nil, errors.New("config case indicates use of TLS client certs but also indicates NOT using TLS")
		}
		if !contains(tlsCases, true) && !features.SupportsTLS {
			// TLS not supported?
			return nil, errors.New("config case indicates use of TLS client certs but TLS is not supported")
		}
		tlsClientCertCases = []bool{unresolvedCase.GetUseTlsClientCerts()}
	}
	if unresolvedCase.UseMessageReceiveLimit != nil {
		msgReceiveLimitCases = []bool{unresolvedCase.GetUseMessageReceiveLimit()}
	}
	return computeCasesFromFeatures(impliedFeatures, tlsCases, tlsClientCertCases, msgReceiveLimitCases), nil
}

func contains[T comparable, S ~[]T](slice S, find T) bool {
	for _, elem := range slice {
		if elem == find {
			return true
		}
	}
	return false
}

func only[T comparable, S ~[]T](slice S, find T) bool {
	for _, elem := range slice {
		if elem != find {
			return false
		}
	}
	return true
}
