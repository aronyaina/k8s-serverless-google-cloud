package main

import (
	"fmt"
	"k8s-serverless/connexion"
	"k8s-serverless/gcp/access"
	"k8s-serverless/gcp/infra"
	"k8s-serverless/gcp/network"
	"log"
	"strconv"

	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/compute"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const VM_NUMBER = 1

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		privateKeyPath := "./private-key.pem"
		privateKey, publicKey, err := connexion.GenerateSSHKeyPair(privateKeyPath)
		if err != nil {
			log.Println("Failed to generate SSH key pair:", err)
			return err
		}

		network1, err := network.GeneratePrivateNetwork(ctx)
		if err != nil {
			log.Println("error While creating network")
			return err
		}

		err = network.GenerateFirewall(ctx, network1)
		if err != nil {
			log.Println("error While creating firewall")
			return err
		}

		subnet, err := network.GenerateSubNetwork(ctx, network1)
		if err != nil {
			log.Println("error while creating subnet")
			return err
		}

		serviceAccount, err := access.GenerateServiceAccount(ctx)
		if err != nil {
			log.Println("error while creating service account")
			return err
		}

		bucket, err := infra.GenerateBucket(ctx, "my-bucket-329102")
		if err != nil {
			log.Println("error while creating bucket")
			return err
		}

		_, err = access.GenerateIamMember(ctx, serviceAccount)
		if err != nil {
			log.Println("error while creating iam member")
			return err
		}

		_, err = access.GenerateIamBindingOfBucket(ctx, bucket, serviceAccount)
		if err != nil {
			log.Println("error while creating iam binding")
			return err
		}

		masterMachine, err := infra.GenerateMasterMachine(ctx, VM_NUMBER, "master-node", network1, subnet, bucket, publicKey, serviceAccount)
		if err != nil {
			log.Println("error while creating master machine")
			return err
		}

		var allMachines []*compute.Instance
		var last *compute.Instance = masterMachine

		if VM_NUMBER >= 1 {
			for i := 1; i <= VM_NUMBER; i++ {
				machine, err := infra.GenerateWorkerMachine(ctx, i, last, "worker-node-"+strconv.Itoa(i), network1, subnet, bucket, serviceAccount)

				if err != nil {
					return err
				}

				// Ensure the next worker depends on the current one
				last = machine
				allMachines = append(allMachines, machine)
			}
		}

		masterExternalIp := masterMachine.NetworkInterfaces.Index(pulumi.Int(0)).AccessConfigs().Index(pulumi.Int(0)).NatIp()

		kubeConfig := masterExternalIp.ApplyT(func(ip *string) (interface{}, error) {
			if ip == nil {
				return nil, fmt.Errorf("masterExternalIp is nil")
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

		k8sConfig := pulumi.Output(kubeConfig).ApplyT(func(result interface{}) (string, error) {
			if resultStr, ok := result.(string); ok {
				return resultStr, nil
			}
			return "", fmt.Errorf("unexpected type for kubeconfig result")
		}).(pulumi.StringOutput)

		provider, err := kubernetes.NewProvider(ctx, "first-provider", &kubernetes.ProviderArgs{
			Kubeconfig: k8sConfig,
		})

		if err != nil {
			return err
		}

		ctx.Export("Provider", provider)
		ctx.Export("Kubeconfig", k8sConfig)

		ctx.Export("serviceAccountName", serviceAccount.Name)
		ctx.Export("BucketName", bucket.Name)

		for i, machine := range allMachines {
			ctx.Export("instanceName"+strconv.Itoa(i+1), machine.Name) // Add an index to make export names unique
			ctx.Export("instanceExternalIP"+strconv.Itoa(i+1), machine.NetworkInterfaces.Index(pulumi.Int(0)).AccessConfigs().Index(pulumi.Int(0)).NatIp())
			ctx.Export("instanceInternalIP"+strconv.Itoa(i+1), machine.NetworkInterfaces.Index(pulumi.Int(0)).NetworkIp())
		}
		ctx.Export("masterName", masterMachine.Name) // Add an index to make export names unique
		ctx.Export("masterExternalIP", masterMachine.NetworkInterfaces.Index(pulumi.Int(0)).AccessConfigs().Index(pulumi.Int(0)).NatIp())
		ctx.Export("masterInternalIP", masterMachine.NetworkInterfaces.Index(pulumi.Int(0)).NetworkIp())

		//ctx.Export("masterMachinePrivateKey", pulumi.String(privateKey))
		//ctx.Export("masterMachinePublicKey", pulumi.String(publicKey))

		ctx.Export("networkName", network1.Name)
		ctx.Export("subnetName", subnet.Name)
		return nil
	})
}
