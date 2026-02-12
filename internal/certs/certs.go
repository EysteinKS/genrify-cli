package certs

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"time"
)

// GenerateOptions contains options for certificate generation.
type GenerateOptions struct {
	// SANs are Subject Alternative Names (DNS names and IP addresses).
	SANs []string
	// ValidFor is the duration the certificate is valid for.
	ValidFor time.Duration
	// CommonName is the certificate's CN field.
	CommonName string
}

// Result contains the generated certificate and key in PEM format.
type Result struct {
	CertPEM []byte
	KeyPEM  []byte
}

// Generate creates a self-signed ECDSA P-256 certificate and private key.
func Generate(opts GenerateOptions) (*Result, error) {
	if opts.ValidFor == 0 {
		opts.ValidFor = 365 * 24 * time.Hour // 1 year
	}
	if opts.CommonName == "" {
		opts.CommonName = "localhost"
	}

	// Generate ECDSA P-256 private key.
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generate private key: %w", err)
	}

	// Create certificate template.
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, fmt.Errorf("generate serial number: %w", err)
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(opts.ValidFor)

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: opts.CommonName,
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// Add SANs if provided.
	if len(opts.SANs) > 0 {
		for _, san := range opts.SANs {
			template.DNSNames = append(template.DNSNames, san)
		}
	}

	// Create self-signed certificate.
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, fmt.Errorf("create certificate: %w", err)
	}

	// Encode certificate to PEM.
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	// Encode private key to PEM.
	keyBytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("marshal private key: %w", err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: keyBytes,
	})

	return &Result{
		CertPEM: certPEM,
		KeyPEM:  keyPEM,
	}, nil
}

// EnsureCerts checks if cert and key files exist at the given paths.
// If they exist and are valid, it returns the paths unchanged.
// If they don't exist, it generates new self-signed certificates and writes them.
// If certPath or keyPath are empty, defaults to ~/.config/genrify/.certs/localhost.pem and localhost-key.pem.
func EnsureCerts(certPath, keyPath string) (string, string, error) {
	// Use default paths if not provided.
	if certPath == "" || keyPath == "" {
		configDir, err := os.UserConfigDir()
		if err != nil {
			return "", "", fmt.Errorf("get user config dir: %w", err)
		}
		certDir := filepath.Join(configDir, "genrify", ".certs")
		if certPath == "" {
			certPath = filepath.Join(certDir, "localhost.pem")
		}
		if keyPath == "" {
			keyPath = filepath.Join(certDir, "localhost-key.pem")
		}
	}

	// Check if both files exist.
	if fileExists(certPath) && fileExists(keyPath) {
		return certPath, keyPath, nil
	}

	// Generate new certificates.
	result, err := Generate(GenerateOptions{
		CommonName: "localhost",
		SANs:       []string{"localhost", "127.0.0.1", "::1"},
		ValidFor:   365 * 24 * time.Hour,
	})
	if err != nil {
		return "", "", fmt.Errorf("generate certificates: %w", err)
	}

	// Ensure directory exists.
	certDir := filepath.Dir(certPath)
	if err := os.MkdirAll(certDir, 0o700); err != nil {
		return "", "", fmt.Errorf("create cert directory: %w", err)
	}

	// Write certificate file.
	if err := os.WriteFile(certPath, result.CertPEM, 0o600); err != nil {
		return "", "", fmt.Errorf("write certificate: %w", err)
	}

	// Write key file.
	if err := os.WriteFile(keyPath, result.KeyPEM, 0o600); err != nil {
		return "", "", fmt.Errorf("write private key: %w", err)
	}

	return certPath, keyPath, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
