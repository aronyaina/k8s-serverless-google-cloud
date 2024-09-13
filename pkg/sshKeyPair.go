package pkg

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"

	"golang.org/x/crypto/ssh"
)

func GenerateSSHKeyPair(privateKeyPath string, publicKeyPath string) (string, string, error) {
	// Generate RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate private key: %v", err)
	}

	// Marshal the private key to PEM format
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	if err := SaveToFile(privateKeyPath, privateKeyPEM); err != nil {
		return "", "", fmt.Errorf("failed to save private key to file: %v", err)
	}
	// Create the corresponding public key
	publicKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate public key: %v", err)
	}

	publicKeyBytes := ssh.MarshalAuthorizedKey(publicKey)

	if err := SaveToFile(publicKeyPath, publicKeyBytes); err != nil {
		return "", "", fmt.Errorf("failed to save public key to file: %v", err)
	}

	return string(privateKeyPEM), string(publicKeyBytes), nil
}

func LoadPrivateKeyAndPublicKey(privateKeyPath, publicKeyPath string) (string, string, error) {
	// Load private key
	privateKeyData, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to read private key file: %v", err)
	}

	// Decode PEM block
	block, _ := pem.Decode(privateKeyData)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return "", "", fmt.Errorf("failed to decode private key PEM block")
	}

	// Parse the private key (this step ensures it's a valid private key but returns a parsed structure)
	_, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse private key: %v", err)
	}

	// Load public key
	publicKeyData, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to read public key file: %v", err)
	}

	// Parse the public key
	_, _, _, _, err = ssh.ParseAuthorizedKey(publicKeyData)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse public key: %v", err)
	}

	// Return keys as strings (no need to parse again, just return the PEM and public key strings)
	return string(privateKeyData), string(publicKeyData), nil
}
