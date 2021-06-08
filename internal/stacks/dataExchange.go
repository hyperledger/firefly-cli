package stacks

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"time"
)

type DataExchangeListenerConfig struct {
	Hostname string `json:"hostname,omitempty"`
	Port     int    `json:"port,omitempty"`
}

type PeerConfig struct {
	ID       string `json:"id,omitempty"`
	Endpoint string `json:"endpoint,omitempty"`
}

type DataExchangeConfig struct {
	API   *DataExchangeListenerConfig `json:"api,omitempty"`
	P2P   *DataExchangeListenerConfig `json:"p2p,omitempty"`
	Peers []*PeerConfig               `json:"peers"`
}

func (s *Stack) GenerateDataExchangeConfig(memberId string) *DataExchangeConfig {

	peers := make([]*PeerConfig, len(s.Members)-1)
	i := 0

	for _, member := range s.Members {
		if member.ID != memberId {
			peers[i] = &PeerConfig{
				ID:       member.ID,
				Endpoint: fmt.Sprintf("https://dataexchange_%s:3001", member.ID),
			}

		}
	}

	return &DataExchangeConfig{
		API: &DataExchangeListenerConfig{
			Hostname: "0.0.0.0",
			Port:     3000,
		},
		P2P: &DataExchangeListenerConfig{
			Hostname: "0.0.0.0",
			Port:     3001,
		},
		Peers: []*PeerConfig{},
	}
}

func CreateCert(memberId string) (*bytes.Buffer, *bytes.Buffer, error) {
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(2019),
		Subject: pkix.Name{
			CommonName:   "localhost",
			Organization: []string{memberId},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	caPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, err
	}

	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return nil, nil, err
	}

	caPEM := new(bytes.Buffer)
	pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})

	caPrivKeyPEM := new(bytes.Buffer)
	pem.Encode(caPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(caPrivKey),
	})

	cert := &x509.Certificate{
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			CommonName:   "localhost",
			Organization: []string{memberId},
		},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	certPrivKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, ca, &certPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return nil, nil, err
	}

	certPEM := new(bytes.Buffer)
	pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	privKeyBytes := x509.MarshalPKCS1PrivateKey(certPrivKey)
	if err != nil {
		return nil, nil, err
	}

	certPrivKeyPEM := new(bytes.Buffer)
	pem.Encode(certPrivKeyPEM, &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privKeyBytes,
	})

	return certPEM, certPrivKeyPEM, nil
}
