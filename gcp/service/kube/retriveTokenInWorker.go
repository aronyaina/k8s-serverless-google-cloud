package kube

import (
	"fmt"

	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/storage"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func RetrieveTokenFromBucket(ctx *pulumi.Context, workerExternalIp pulumi.StringPtrOutput, workerIndice int, bucket *storage.Bucket, privateKey string, triggers pulumi.Array) (pulumi.Output, error) {
	//masterExternalIp := masterMachine.NetworkInterfaces.Index(pulumi.Int(0)).AccessConfigs().Index(pulumi.Int(0)).NatIp()

	copyToken := workerExternalIp.ApplyT(func(ip *string) (interface{}, error) {
		if ip == nil {
			return nil, fmt.Errorf("masterExternalIp is nil")
		}

		ready, err := WaitForLockFile(ctx, privateKey, *ip)
		if err != nil {
			return nil, err
		}

		triggers = append(triggers, ready)

		copyCommand := pulumi.All(workerIndice, bucket.Name).ApplyT(func(args []interface{}) (string, error) {
			workerNum := args[0].(int)
			bucketName := args[1].(string)

			return fmt.Sprintf(`
			echo "Copying join command"
			gsutil cp gs://%s/join-command-%v.txt $JOIN_COMMAND_FILE
			JOIN_COMMAND=$(cat /home/join-command-%v.txt)
			echo "Token Initialized execution completed"
			sudo bash -c "$JOIN_COMMAND"
			echo "Join command successfully executed"
			`, bucketName, workerNum), nil
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

	return copyToken, nil
}
