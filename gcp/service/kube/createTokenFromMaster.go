package kube

import (
	"fmt"

	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/compute"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/storage"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func GenerateTokenFromMasterAndUploadIt(ctx *pulumi.Context, masterExternalIp pulumi.StringPtrOutput, workerNumber int, bucket *storage.Bucket, masterMachine *compute.Instance, privateKey string, triggers pulumi.Array) (pulumi.Output, error) {
	//masterExternalIp := masterMachine.NetworkInterfaces.Index(pulumi.Int(0)).AccessConfigs().Index(pulumi.Int(0)).NatIp()
	CopyToken := masterExternalIp.ApplyT(func(ip *string) (interface{}, error) {
		if ip == nil {
			return nil, fmt.Errorf("masterExternalIp is nil")
		}

		k8sReady, err := WaitForLockFile(ctx, privateKey, *ip)
		if err != nil {
			return nil, err
		}
		triggers = append(triggers, k8sReady)

		copyCommand := pulumi.All(workerNumber, bucket.Name).ApplyT(func(args []interface{}) (string, error) {
			workerNum := args[0].(int)
			bucketName := args[1].(string)

			return fmt.Sprintf(`
			for i in $(seq 1 %d); do
				echo "Generating join command for node $i"
				JOIN_COMMAND=$(sudo microk8s add-node | grep 'microk8s join' | sed -n '2p')
				if [ -n "$JOIN_COMMAND" ]; then
					echo $JOIN_COMMAND > /home/join-command-$i.txt
					gsutil cp /home/join-command-$i.txt gs://%s/join-command-$i.txt
					echo "Join command for node $i stored and uploaded."
				fi
			done
			echo "Token generation completed"
			`, workerNum, bucketName), nil
		}).(pulumi.StringOutput)

		// Command arguments for fetching the key data
		copyTokenCmdArgs := &remote.CommandArgs{
			Create: copyCommand,
			Connection: &remote.ConnectionArgs{
				Host:       pulumi.String(*ip),
				User:       pulumi.String("pulumi"),
				PrivateKey: pulumi.String(privateKey),
			},
			Triggers: triggers,
		}

		// Create the remote command for the key
		copyCmd, err := remote.NewCommand(ctx, "keyCmd", copyTokenCmdArgs)
		if err != nil {
			return nil, err
		}

		// Wait for both commands to complete and get their results
		keyResult := copyCmd.Stdout.ApplyT(func(result interface{}) (string, error) {
			if resultStr, ok := result.(string); ok {
				return resultStr, nil
			}
			return "", fmt.Errorf("unexpected type for key result")
		})

		// Combine results
		return keyResult, nil
	})

	return CopyToken, nil
}
