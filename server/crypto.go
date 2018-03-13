package server

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"os"
	"time"

	"github.com/dgraph-io/badger"
)

const (
	keyKey        = "/config/key"
	privateKeyKey = "/config/privateKey"
	certKey       = "/config/cert"

	validFor = 10 * 365 * 24 * time.Hour
)

func publicKey(priv interface{}) interface{} {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	default:
		return nil
	}
}

func pemBlockForKey(priv interface{}) *pem.Block {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(k)}
	case *ecdsa.PrivateKey:
		b, err := x509.MarshalECPrivateKey(k)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to marshal ECDSA private key: %v", err)
			os.Exit(2)
		}
		return &pem.Block{Type: "EC PRIVATE KEY", Bytes: b}
	default:
		return nil
	}
}

func (s *Server) loadCert() error {
	if err := s.db.View(func(txn *badger.Txn) error {
		keyItem, err := txn.Get([]byte(keyKey))
		if err != nil {
			return err
		}
		keyValue, err := keyItem.Value()
		if err != nil {
			return err
		}
		certItem, err := txn.Get([]byte(certKey))
		if err != nil {
			return err
		}
		certValue, err := certItem.Value()
		if err != nil {
			return err
		}

		s.certPublic = string(certValue)

		cert, err := tls.X509KeyPair(certValue, keyValue)
		if err != nil {
			return err
		}
		s.cert = &cert

		privateKeyItem, err := txn.Get([]byte(privateKeyKey))
		if err != nil {
			return err
		}
		privKey, err := privateKeyItem.Value()
		if err != nil {
			return err
		}
		s.key, err = x509.ParseECPrivateKey(privKey)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	return ErrUnimplemented
}

func (s *Server) generateCert() error {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatalf("failed to generate private key: %s", err)
	}
	s.key = priv
	privKey, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return err
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(validFor)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		log.Fatalf("failed to generate serial number: %s", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"UBC 416"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	template.IPAddresses = append(template.IPAddresses, getOutboundIP())

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, publicKey(priv), priv)
	if err != nil {
		return err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	keyPEM := pem.EncodeToMemory(pemBlockForKey(priv))
	s.certPublic = string(certPEM)

	if err := s.db.Update(func(txn *badger.Txn) error {
		if err := txn.Set([]byte(privateKeyKey), privKey); err != nil {
			return err
		}
		if err := txn.Set([]byte(certKey), certPEM); err != nil {
			return err
		}
		if err := txn.Set([]byte(keyKey), keyPEM); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return err
	}
	s.cert = &cert

	return nil
}

func (s *Server) loadOrGenerateCert() error {
	if err := s.loadCert(); err == badger.ErrKeyNotFound {
		if err := s.generateCert(); err != nil {
			return err
		}
	}
	return nil
}
