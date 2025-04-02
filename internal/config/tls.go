/*
Security in distributed services can be broken down into three steps:
1. Encrypt data in-flight to protect against the man-in-the-middle attacks,
2. Authenticate to identify clients, and
3. Authorize to determine the permissions of the identified clients.
*/

package config

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
)

type TLSConfig struct {
	CertFile      string
	KeyFile       string
	CAFile        string
	ServerAddress string
	Server        bool
}

func SetupTLSConfig(cfg TLSConfig) (*tls.Config, error) {
	var err error
	tlsConfig := &tls.Config{}

	// Load server certificate and key
	if cfg.CertFile != "" && cfg.KeyFile != "" {
		tlsConfig.Certificates = make([]tls.Certificate, 1)
		tlsConfig.Certificates[0], err = tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load X.509 key pair: %w", err)
		}
	}

	// Load CA file
	if cfg.CAFile != "" {
		b, err := os.ReadFile(cfg.CAFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA file: %w", err)
		}
		ca := x509.NewCertPool()
		if !ca.AppendCertsFromPEM(b) {
			return nil, fmt.Errorf("failed to parse root certificate file: %q", cfg.CAFile)
		}

		if cfg.Server {
			tlsConfig.ClientCAs = ca
			tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		} else {
			tlsConfig.RootCAs = ca
		}
	}

	// Ensure ServerName is set for clients
	if !cfg.Server && cfg.ServerAddress != "" {
		tlsConfig.ServerName = cfg.ServerAddress
	}

	return tlsConfig, nil
}
