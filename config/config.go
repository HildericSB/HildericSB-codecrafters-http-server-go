package config

import (
	"crypto/tls"
)

const CRLF = "\r\n"
const BufferSize = 4096

type Config struct {
	*TLSConfig
	Port       string
	FileDir    string
	LogLevel   string
	BufferSize int
}

type TLSConfig struct {
	TLSPort     string
	CertPem     []byte
	KeyPem      []byte
	certificate *tls.Certificate
	TlSConfig   *tls.Config
}

func defaultTlsConfig() *TLSConfig {
	tlsConfig := TLSConfig{
		CertPem: []byte(`
-----BEGIN CERTIFICATE-----
MIIBhTCCASugAwIBAgIQIRi6zePL6mKjOipn+dNuaTAKBggqhkjOPQQDAjASMRAw
DgYDVQQKEwdBY21lIENvMB4XDTE3MTAyMDE5NDMwNloXDTE4MTAyMDE5NDMwNlow
EjEQMA4GA1UEChMHQWNtZSBDbzBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABD0d
7VNhbWvZLWPuj/RtHFjvtJBEwOkhbN/BnnE8rnZR8+sbwnc/KhCk3FhnpHZnQz7B
5aETbbIgmuvewdjvSBSjYzBhMA4GA1UdDwEB/wQEAwICpDATBgNVHSUEDDAKBggr
BgEFBQcDATAPBgNVHRMBAf8EBTADAQH/MCkGA1UdEQQiMCCCDmxvY2FsaG9zdDo1
NDUzgg4xMjcuMC4wLjE6NTQ1MzAKBggqhkjOPQQDAgNIADBFAiEA2zpJEPQyz6/l
Wf86aX6PepsntZv2GYlA5UpabfT2EZICICpJ5h/iI+i341gBmLiAFQOyTDT+/wQc
6MF9+Yw1Yy0t
-----END CERTIFICATE-----`),
		KeyPem: []byte(`
-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIIrYSSNQFaA2Hwf1duRSxKtLYX5CB04fSeQ6tF1aY/PuoAoGCCqGSM49
AwEHoUQDQgAEPR3tU2Fta9ktY+6P9G0cWO+0kETA6SFs38GecTyudlHz6xvCdz8q
EKTcWGekdmdDPsHloRNtsiCa697B2O9IFA==
-----END EC PRIVATE KEY-----`),
	}

	cert, err := tls.X509KeyPair(tlsConfig.CertPem, tlsConfig.KeyPem)
	if err != nil {
		panic("Error with TLS Key pair")
	}

	tlsConfig.certificate = &cert

	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	tlsConfig.TlSConfig = config
	tlsConfig.TLSPort = "4222"

	return &tlsConfig
}

func DefaultConfig() *Config {
	tlsConfig := defaultTlsConfig()

	return &Config{
		Port:      "4221",
		FileDir:   "/tmp/",
		LogLevel:  "info",
		TLSConfig: tlsConfig,
	}
}
