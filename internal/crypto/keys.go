package crypto

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
)

// KeyManager defines the interface for generating and managing keys.
type KeyManager interface {
	Generate() error
	PrivateKeyPath() string
	PublicKeyPath() string
}

type keyManager struct {
	projectDir string
}

// NewKeyManager returns a KeyManager that stores keys in projectDir/.revora/keys.
func NewKeyManager(projectDir string) KeyManager {
	return &keyManager{projectDir: projectDir}
}

func (km *keyManager) PrivateKeyPath() string {
	return filepath.Join(km.projectDir, ".revora", "keys", "private.pem")
}

func (km *keyManager) PublicKeyPath() string {
	return filepath.Join(km.projectDir, ".revora", "keys", "public.pem")
}

// Generate creates an RSA-4096 key pair and saves them to disk.
func (km *keyManager) Generate() error {
	keysDir := filepath.Join(km.projectDir, ".revora", "keys")
	if err := os.MkdirAll(keysDir, 0700); err != nil {
		return fmt.Errorf("create keys dir: %w", err)
	}

	priv, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return fmt.Errorf("generate key: %w", err)
	}

	// Save private key
	privFile, err := os.OpenFile(km.PrivateKeyPath(), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer privFile.Close()
	if err := pem.Encode(privFile, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(priv),
	}); err != nil {
		return err
	}

	// Save public key
	pubFile, err := os.OpenFile(km.PublicKeyPath(), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer pubFile.Close()
	pubBytes, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		return err
	}
	if err := pem.Encode(pubFile, &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	}); err != nil {
		return err
	}
	return nil
}

// LoadPrivateKey reads an RSA private key from a PEM file.
func LoadPrivateKey(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read private key: %w", err)
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("no PEM block found in %s", path)
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}
	return priv, nil
}

// Sign signs data with the given RSA private key using SHA-256 and PKCS1v15.
func Sign(priv *rsa.PrivateKey, data []byte) ([]byte, error) {
	hashed := sha256.Sum256(data)
	signature, err := rsa.SignPKCS1v15(rand.Reader, priv, crypto.SHA256, hashed[:])
	if err != nil {
		return nil, fmt.Errorf("sign: %w", err)
	}
	return signature, nil
}
