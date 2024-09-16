package kube

import (
	"fmt"

	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/storage"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func RetrieveTokenFromBucket(ctx *pulumi.Context, workerExternalIp pulumi.StringPtrOutput, workerIndice int, bucket *storage.Bucket, privateKey string, triggers pulumi.Array) (pulumi.Output, error) {
	//masterExternalIp := masterMachine.NetworkInterfaces.Index(pulumi.Int(0)).AccessConfigs().Index(pulumi.Int(0)).NatIp()

	copyCommand := pulumi.All(workerIndice, bucket.Name).ApplyT(func(args []interface{}) (string, error) {
		workerNum := args[0].(int)
		bucketName := args[1].(string)

		return fmt.Sprintf(`
			MAX_COPY_RETRIES=10
			COPY_RETRY_DELAY=15
			JOIN_COMMAND_FILE="/home/$(whoami)/join-command-%v.txt"
			for ((i=1; i<=MAX_COPY_RETRIES; i++)); do
				echo "Attempt $i to copy join command from GCS"
				sudo gsutil cp gs://%s/join-command-%v.txt $JOIN_COMMAND_FILE
				if [ -f "$JOIN_COMMAND_FILE" ]; then
					echo "Join command file copied successfully on attempt $i"
					break
				else
					echo "Join command file not found, retrying in $COPY_RETRY_DELAY seconds..."
					sleep $COPY_RETRY_DELAY
				fi
			done
			JOIN_COMMAND=$(cat /home/$(whoami)/join-command-%v.txt)

			# Retry mechanism with backoff delay
			MAX_RETRIES=5
			RETRY_DELAY=10
			for ((i=1; i<=MAX_RETRIES; i++)); do
				echo "Attempt $i to join the cluster"
				sudo bash -c "$JOIN_COMMAND"
				STATUS=$?
				if [ $STATUS -eq 0 ]; then
					echo "Successfully joined the cluster on attempt $i"
					break
				else
					echo "Failed to join the cluster, retrying in $RETRY_DELAY seconds..."
					sleep $RETRY_DELAY
				fi
			done
			`, workerNum, bucketName, workerNum, workerNum), nil

	}).(pulumi.StringOutput)
	copyToken := workerExternalIp.ApplyT(func(ip *string) (interface{}, error) {
		if ip == nil {
			return nil, fmt.Errorf("workerExternalIp is nil")
		}

		ready, err := WaitForLockFile(ctx, privateKey, fmt.Sprintf("wait-lock-worker-%v", workerIndice), *ip)
		if err != nil {
			return nil, err
		}

		triggers = append(triggers, ready)

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
		copyCmd, err := remote.NewCommand(ctx, fmt.Sprintf("copy-token-to-%v", workerIndice), copyTokenCmdArgs)
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
