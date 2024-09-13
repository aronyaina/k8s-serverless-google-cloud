package service

import (
	"k8s-serverless/pkg"
	"log"
)

func GenerateSshIfItDoesntExist() (privateKey string, publicKey string, err error) {
	privateKeyPath := "./private-key.pem"
	publicKeyPath := "./public-key.pem"

	if pkg.FileExist(privateKeyPath) && pkg.FileExist(publicKeyPath) {
		privateKey, publicKey, err = pkg.LoadPrivateKeyAndPublicKey(privateKeyPath, publicKeyPath)
		if err != nil {
			log.Println("Failed to load SSH key pair:", err)
			return "", "", err
		}
	} else {
		privateKey, publicKey, err = pkg.GenerateSSHKeyPair(privateKeyPath, publicKeyPath)
		if err != nil {
			log.Println("Failed to generate SSH key pair:", err)
			return "", "", err
		}
	}
	return privateKey, publicKey, nil
}
