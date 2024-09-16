package initialization

import (
	"fmt"
	"k8s-serverless/utils"
	"log"
)

func GenerateSshIfItDoesntExist(instance_name string) (privateKey string, publicKey string, err error) {
	privateKeyPath := fmt.Sprintf("./private-key-%s.pem", instance_name)
	publicKeyPath := fmt.Sprintf("./public-key-%s.pem", instance_name)

	if utils.FileExist(privateKeyPath) && utils.FileExist(publicKeyPath) {
		privateKey, publicKey, err = utils.LoadPrivateKeyAndPublicKey(privateKeyPath, publicKeyPath)
		if err != nil {
			log.Println("Failed to load SSH key pair:", err)
			return "", "", err
		}
	} else {
		privateKey, publicKey, err = utils.GenerateSSHKeyPair(privateKeyPath, publicKeyPath)
		if err != nil {
			log.Println("Failed to generate SSH key pair:", err)
			return "", "", err
		}
	}
	return privateKey, publicKey, nil
}
