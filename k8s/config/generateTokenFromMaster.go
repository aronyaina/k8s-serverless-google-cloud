package config

import (
	"fmt"
	k8sService "k8s-serverless/k8s/service"

	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/compute"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/storage"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func GenerateTokenFromMasterAndUploadIt(ctx *pulumi.Context, workerNumber int, bucket *storage.Bucket, masterMachine *compute.Instance, privateKey string) (pulumi.Output, error) {
	masterExternalIp := masterMachine.NetworkInterfaces.Index(pulumi.Int(0)).AccessConfigs().Index(pulumi.Int(0)).NatIp()

	kubeConfig := masterExternalIp.ApplyT(func(ip *string) (interface{}, error) {
		if ip == nil {
			return nil, fmt.Errorf("masterExternalIp is nil")
		}

		ready, err := k8sService.WaitForLockFile(ctx, privateKey, *ip)
		if err != nil {
			return nil, err
		}

		copyCommand := pulumi.All(workerNumber, bucket.Name).ApplyT(func(args []interface{}) (string, error) {
			workerNum := args[0].(int)
			bucketName := args[1].(string)

			return fmt.Sprintf(`
			for i in $(seq 1 %d); do
				JOIN_COMMAND=$(sudo microk8s add-node | grep 'microk8s join' | sed -n '2p')
				if [ -n "$JOIN_COMMAND" ]; then
					echo $JOIN_COMMAND > /home/join-command-$i.txt
					gsutil cp /home/join-command-$i.txt gs://%s/join-command-$i.txt
					echo "Join command for node $i stored and uploaded." | tee -a $LOG_FILE
				fi
			done
			echo "Script execution completed"
			`, workerNum, bucketName), nil
		}).(pulumi.StringOutput)

		// Command arguments for fetching the key data
		keyCmdArgs := &remote.CommandArgs{
			Create: pulumi.String(copyCommand),
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
		keyResult := keyCmd.Stdout.ApplyT(func(result interface{}) (string, error) {
			if resultStr, ok := result.(string); ok {
				return resultStr, nil
			}
			return "", fmt.Errorf("unexpected type for key result")
		})

		// Combine results
		kubeConfig := pulumi.All(keyResult).ApplyT(func(results []interface{}) (string, error) {
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
