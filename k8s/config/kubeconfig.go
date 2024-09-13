package config

import (
	"fmt"
	k8sService "k8s-serverless/k8s/service"

	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func GenerateMasterKubeConfig(ctx *pulumi.Context, masterMachine *compute.Instance, privateKey string) (pulumi.Output, error) {
	masterExternalIp := masterMachine.NetworkInterfaces.Index(pulumi.Int(0)).AccessConfigs().Index(pulumi.Int(0)).NatIp()

	kubeConfig := masterExternalIp.ApplyT(func(ip *string) (interface{}, error) {
		if ip == nil {
			return nil, fmt.Errorf("masterExternalIp is nil")
		}

		ready, err := k8sService.WaitForLockFile(ctx, privateKey, *ip)
		if err != nil {
			return nil, err
		}

		certificateData := fmt.Sprintf(`sudo microk8s kubectl config view --raw -o jsonpath='{.users[?(@.name=="%s")].user.client-certificate-data}'`, "admin")
		certificateKey := fmt.Sprintf(`sudo microk8s kubectl config view --raw -o jsonpath='{.users[?(@.name=="%s")].user.client-key-data}'`, "admin")

		// Command arguments for fetching the certificate data
		certificateCmdArgs := &remote.CommandArgs{
			Create: pulumi.String(certificateData),
			Connection: &remote.ConnectionArgs{
				Host:       pulumi.String(*ip),
				User:       pulumi.String("pulumi"),
				PrivateKey: pulumi.String(privateKey),
			},
			Triggers: pulumi.Array{ready},
		}

		// Create the remote command for the certificate
		certificateCmd, err := remote.NewCommand(ctx, "certificateCmd", certificateCmdArgs)
		if err != nil {
			return nil, err
		}

		// Command arguments for fetching the key data
		keyCmdArgs := &remote.CommandArgs{
			Create: pulumi.String(certificateKey),
			Connection: &remote.ConnectionArgs{
				Host:       pulumi.String(*ip),
				User:       pulumi.String("pulumi"),
				PrivateKey: pulumi.String(privateKey),
			},
			Triggers: pulumi.Array{ready},
		}

		// Create the remote command for the key
		keyCmd, err := remote.NewCommand(ctx, "keyCmd", keyCmdArgs)
		if err != nil {
			return nil, err
		}

		// Wait for both commands to complete and get their results
		certificateResult := certificateCmd.Stdout.ApplyT(func(result interface{}) (string, error) {
			if resultStr, ok := result.(string); ok {
				return resultStr, nil
			}
			return "", fmt.Errorf("unexpected type for certificate result")
		})

		keyResult := keyCmd.Stdout.ApplyT(func(result interface{}) (string, error) {
			if resultStr, ok := result.(string); ok {
				return resultStr, nil
			}
			return "", fmt.Errorf("unexpected type for key result")
		})

		// Combine results
		kubeConfig := pulumi.All(certificateResult, keyResult).ApplyT(func(results []interface{}) (string, error) {
			cert, key := results[0], results[1]
			kConfig := fmt.Sprintf(`
apiVersion: v1
clusters:
- cluster:
    insecure-skip-tls-verify: true
    server: https://%s:16443
  name: microk8s-cluster
contexts:
- context:
    cluster: microk8s-cluster
    user: admin
  name: microk8s
current-context: microk8s
kind: Config
preferences: {}
users:
- name: admin
  user:
    client-certificate-data: %s
    client-key-data: %s
`, *ip, cert, key)

			return kConfig, nil
		})

		return kubeConfig, nil
	})

	return kubeConfig, nil
}
