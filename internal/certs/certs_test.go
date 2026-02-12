package certs

import (
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGenerate(t *testing.T) {
	t.Run("generates valid certificate", func(t *testing.T) {
		opts := GenerateOptions{
			CommonName: "test.local",
			SANs:       []string{"test.local", "localhost"},
			ValidFor:   24 * time.Hour,
		}

		result, err := Generate(opts)
		if err != nil {
			t.Fatalf("Generate() error = %v", err)
		}

		if len(result.CertPEM) == 0 {
			t.Error("CertPEM is empty")
		}
		if len(result.KeyPEM) == 0 {
			t.Error("KeyPEM is empty")
		}

		// Decode and verify certificate.
		block, _ := pem.Decode(result.CertPEM)
		if block == nil {
			t.Fatal("failed to decode certificate PEM")
		}
		if block.Type != "CERTIFICATE" {
			t.Errorf("expected CERTIFICATE block, got %s", block.Type)
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			t.Fatalf("parse certificate: %v", err)
		}

		// Verify certificate fields.
		if cert.Subject.CommonName != opts.CommonName {
			t.Errorf("CommonName = %s, want %s", cert.Subject.CommonName, opts.CommonName)
		}
		if len(cert.DNSNames) != len(opts.SANs) {
			t.Errorf("DNSNames count = %d, want %d", len(cert.DNSNames), len(opts.SANs))
		}

		// Verify key usage.
		if cert.KeyUsage&x509.KeyUsageKeyEncipherment == 0 {
			t.Error("missing KeyUsageKeyEncipherment")
		}
		if cert.KeyUsage&x509.KeyUsageDigitalSignature == 0 {
			t.Error("missing KeyUsageDigitalSignature")
		}

		// Decode and verify private key.
		keyBlock, _ := pem.Decode(result.KeyPEM)
		if keyBlock == nil {
			t.Fatal("failed to decode key PEM")
		}
		if keyBlock.Type != "EC PRIVATE KEY" {
			t.Errorf("expected EC PRIVATE KEY block, got %s", keyBlock.Type)
		}

		key, err := x509.ParseECPrivateKey(keyBlock.Bytes)
		if err != nil {
			t.Fatalf("parse private key: %v", err)
		}
		if key == nil {
			t.Error("private key is nil")
		}
	})

	t.Run("uses defaults", func(t *testing.T) {
		result, err := Generate(GenerateOptions{})
		if err != nil {
			t.Fatalf("Generate() error = %v", err)
		}

		block, _ := pem.Decode(result.CertPEM)
		cert, _ := x509.ParseCertificate(block.Bytes)

		if cert.Subject.CommonName != "localhost" {
			t.Errorf("CommonName = %s, want localhost", cert.Subject.CommonName)
		}

		// Check validity period is approximately 1 year.
		validFor := cert.NotAfter.Sub(cert.NotBefore)
		expected := 365 * 24 * time.Hour
		if validFor < expected-time.Hour || validFor > expected+time.Hour {
			t.Errorf("ValidFor = %v, want ~%v", validFor, expected)
		}
	})
}

func TestEnsureCerts(t *testing.T) {
	t.Run("generates new certificates when missing", func(t *testing.T) {
		tmpDir := t.TempDir()
		certPath := filepath.Join(tmpDir, "test.pem")
		keyPath := filepath.Join(tmpDir, "test-key.pem")

		gotCert, gotKey, err := EnsureCerts(certPath, keyPath)
		if err != nil {
			t.Fatalf("EnsureCerts() error = %v", err)
		}

		if gotCert != certPath {
			t.Errorf("certPath = %s, want %s", gotCert, certPath)
		}
		if gotKey != keyPath {
			t.Errorf("keyPath = %s, want %s", gotKey, keyPath)
		}

		// Verify files exist and are valid.
		certData, err := os.ReadFile(certPath)
		if err != nil {
			t.Fatalf("read cert file: %v", err)
		}
		keyData, err := os.ReadFile(keyPath)
		if err != nil {
			t.Fatalf("read key file: %v", err)
		}

		if len(certData) == 0 {
			t.Error("cert file is empty")
		}
		if len(keyData) == 0 {
			t.Error("key file is empty")
		}

		// Verify certificate contains expected SANs.
		block, _ := pem.Decode(certData)
		cert, _ := x509.ParseCertificate(block.Bytes)
		expectedSANs := map[string]bool{"localhost": true, "127.0.0.1": true, "::1": true}
		for _, san := range cert.DNSNames {
			if !expectedSANs[san] {
				t.Errorf("unexpected SAN: %s", san)
			}
		}
	})

	t.Run("returns existing certificates", func(t *testing.T) {
		tmpDir := t.TempDir()
		certPath := filepath.Join(tmpDir, "existing.pem")
		keyPath := filepath.Join(tmpDir, "existing-key.pem")

		// Create initial certificates.
		if _, _, err := EnsureCerts(certPath, keyPath); err != nil {
			t.Fatalf("initial EnsureCerts() error = %v", err)
		}

		// Read original content.
		origCert, _ := os.ReadFile(certPath)
		origKey, _ := os.ReadFile(keyPath)

		// Call EnsureCerts again.
		gotCert, gotKey, err := EnsureCerts(certPath, keyPath)
		if err != nil {
			t.Fatalf("second EnsureCerts() error = %v", err)
		}

		if gotCert != certPath || gotKey != keyPath {
			t.Error("returned paths don't match")
		}

		// Verify files were not regenerated.
		newCert, _ := os.ReadFile(certPath)
		newKey, _ := os.ReadFile(keyPath)

		if string(origCert) != string(newCert) {
			t.Error("certificate was regenerated")
		}
		if string(origKey) != string(newKey) {
			t.Error("key was regenerated")
		}
	})

	t.Run("uses default paths when empty", func(t *testing.T) {
		// This test modifies the config dir, so we need to restore it.
		// Instead, just verify the function doesn't error.
		t.Skip("skipping test that would create files in user's config dir")
	})

	t.Run("creates directories if needed", func(t *testing.T) {
		tmpDir := t.TempDir()
		certPath := filepath.Join(tmpDir, "deep", "nested", "cert.pem")
		keyPath := filepath.Join(tmpDir, "deep", "nested", "key.pem")

		_, _, err := EnsureCerts(certPath, keyPath)
		if err != nil {
			t.Fatalf("EnsureCerts() error = %v", err)
		}

		if !fileExists(certPath) {
			t.Error("certificate file not created")
		}
		if !fileExists(keyPath) {
			t.Error("key file not created")
		}
	})
}
