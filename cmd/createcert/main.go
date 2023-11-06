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

// createcert is a simple standalone Go program to generate certificates for
// use by the conformance tests. It generates a CA, and then generates a
// client cert, two server certs (for connect and gRPC), and a client/server
// cert for envoy. The certificates expire after 10 years, and once updated
// any downstream projects will need to incorporate the new certs to run the
// conformance tests.
//
// It borrows heavily from https://mkcert.dev/ but tweaks the certificates
// to the types needed for the conformance suite.

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"time"
)

type CertType int

const (
	Client CertType = 1 << iota
	Server
)

const certDir = "cert"

func main() {
	// Generate certificates that last 10 years
	notAfter := time.Now().AddDate(10, 0, 0)
	var certgen generator
	if err := certgen.createCA("ConformanceCA", notAfter); err != nil {
		log.Fatalf("failed to create CA certificate: %v", err)
	}
	if err := certgen.createCert("client", Client, notAfter); err != nil {
		log.Fatalf("failed to create client certificate: %v", err)
	}
	if err := certgen.createCert("envoy", Client|Server, notAfter); err != nil {
		log.Fatalf("failed to create envoy certificate: %v", err)
	}
	for _, name := range []string{"server-connect", "server-grpc"} {
		if err := certgen.createCert(name, Server, notAfter); err != nil {
			log.Fatalf("failed to create server certificate %q: %v", name, err)
		}
	}
}

type generator struct {
	caCert *x509.Certificate
	caKey  *rsa.PrivateKey
}

func (g *generator) createCert(name string, certType CertType, notAfter time.Time) error {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate certificate key: %w", err)
	}
	pub := key.Public()
	serialNumber, err := randomSerialNumber()
	if err != nil {
		return err
	}
	tpl := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: name,
		},
		NotBefore: time.Now(),
		NotAfter:  notAfter,
		KeyUsage:  x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		DNSNames:  []string{name},
	}
	if certType&Client != 0 {
		tpl.ExtKeyUsage = append(tpl.ExtKeyUsage, x509.ExtKeyUsageClientAuth)
	}
	if certType&Server != 0 {
		tpl.ExtKeyUsage = append(tpl.ExtKeyUsage, x509.ExtKeyUsageServerAuth)
	}
	cert, err := x509.CreateCertificate(rand.Reader, tpl, g.caCert, pub, g.caKey)
	if err != nil {
		return fmt.Errorf("failed to generate certificate: %w", err)
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert})
	privDER, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return fmt.Errorf("failed to encode certificate key: %w", err)
	}
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: privDER})
	if err := os.WriteFile(filepath.Join(certDir, name+".crt"), certPEM, 0644); err != nil {
		return fmt.Errorf("failed to save certificate: %w", err)
	}
	if err := os.WriteFile(filepath.Join(certDir, name+".key"), privPEM, 0644); err != nil {
		return fmt.Errorf("failed to save certificate key: %w", err)
	}
	return nil
}

func (g *generator) createCA(name string, notAfter time.Time) error {
	caKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return fmt.Errorf("failed to generate CA key: %w", err)
	}
	pub := caKey.Public()
	spkiASN1, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return fmt.Errorf("failed to encode public key: %w", err)
	}

	var spki struct {
		Algorithm        pkix.AlgorithmIdentifier
		SubjectPublicKey asn1.BitString
	}
	if _, err = asn1.Unmarshal(spkiASN1, &spki); err != nil {
		return fmt.Errorf("failed to decode public key: %w", err)
	}

	skid := sha1.Sum(spki.SubjectPublicKey.Bytes)
	serialNumber, err := randomSerialNumber()
	if err != nil {
		return err
	}
	tpl := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: name,
		},
		SubjectKeyId:          skid[:],
		NotAfter:              notAfter,
		NotBefore:             time.Now(),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLenZero:        true,
	}
	cert, err := x509.CreateCertificate(rand.Reader, tpl, tpl, pub, caKey)
	if err != nil {
		return fmt.Errorf("failed to create certificate: %w", err)
	}
	caCert, err := x509.ParseCertificate(cert)
	if err != nil {
		return fmt.Errorf("failed to parse certificate: %w", err)
	}
	privDER, err := x509.MarshalPKCS8PrivateKey(caKey)
	if err != nil {
		return fmt.Errorf("failed encode save CA key: %w", err)
	}

	if err := os.WriteFile(filepath.Join(certDir, name+".key"), pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: privDER}), 0644); err != nil {
		return fmt.Errorf("failed to save CA certificate")
	}
	if err := os.WriteFile(filepath.Join(certDir, name+".crt"), pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert}), 0644); err != nil {
		return fmt.Errorf("failed to save CA certificate")
	}
	g.caCert = caCert
	g.caKey = caKey
	return nil
}

func randomSerialNumber() (*big.Int, error) {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number: %w", err)
	}
	return serialNumber, nil
}
