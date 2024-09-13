package infra

import (
	"fmt"

	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/compute"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/serviceaccount"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/storage"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func GenerateWorkerMachine(ctx *pulumi.Context, workerIndex int, lastInstance *compute.Instance, instanceName string, network *compute.Network, subnetwork *compute.Subnetwork, bucket *storage.Bucket, service_account *serviceaccount.Account) (*compute.Instance, error) {
	var machineType pulumi.String = "e2-small"
	var region pulumi.String = "us-central1-a"
	var image pulumi.String = "ubuntu-os-cloud/ubuntu-2004-lts"
	var networkTier pulumi.String = "STANDARD"

	workerCommand := pulumi.All(bucket.Name, workerIndex).ApplyT(func(args []interface{}) (string, error) {
		bucketName := args[0].(string)
		workerIndice := args[1].(int)
		return fmt.Sprintf(`#!/bin/bash
			LOG_FILE="/var/log/startup_script.log"
			exec > >(tee -a $LOG_FILE) 2>&1

			sudo apt-get update | tee -a $LOG_FILE
			sudo snap install microk8s --classic --channel=1.31 | tee -a $LOG_FILE
			sudo usermod -aG microk8s $(whoami) | tee -a $LOG_FILE
			newgrp microk8s 
			while ! sudo microk8s status --wait-ready; do
				echo "waiting for microk8s to be ready" | tee -a $LOG_FILE
				sleep 5
			done

			MAX_COPY_RETRIES=10
			COPY_RETRY_DELAY=15
			JOIN_COMMAND_FILE="/home/join-command-%v.txt"
			for ((i=1; i<=MAX_COPY_RETRIES; i++)); do
				echo "Attempt $i to copy join command from GCS" | tee -a $LOG_FILE
				gsutil cp gs://%s/join-command-%v.txt $JOIN_COMMAND_FILE
				if [ -f "$JOIN_COMMAND_FILE" ]; then
					echo "Join command file copied successfully on attempt $i" | tee -a $LOG_FILE
					break
				else
					echo "Join command file not found, retrying in $COPY_RETRY_DELAY seconds..." | tee -a $LOG_FILE
					sleep $COPY_RETRY_DELAY
				fi
			done
			JOIN_COMMAND=$(cat /home/join-command-%v.txt)
			# Retry mechanism with backoff delay
			MAX_RETRIES=5
			RETRY_DELAY=10

			for ((i=1; i<=MAX_RETRIES; i++)); do
				echo "Attempt $i to join the cluster" | tee -a $LOG_FILE
				sudo bash -c "$JOIN_COMMAND"
				STATUS=$?
				if [ $STATUS -eq 0 ]; then
					echo "Successfully joined the cluster on attempt $i" | tee -a $LOG_FILE
					break
				else
					echo "Failed to join the cluster, retrying in $RETRY_DELAY seconds..." | tee -a $LOG_FILE
					sleep $RETRY_DELAY
				fi
			done
			`, workerIndice, bucketName, workerIndice, workerIndice, workerIndice), nil
	}).(pulumi.StringOutput)

	instance, err := compute.NewInstance(ctx, instanceName, &compute.InstanceArgs{
		MachineType: machineType,
		Zone:        region,

		BootDisk: &compute.InstanceBootDiskArgs{
			InitializeParams: &compute.InstanceBootDiskInitializeParamsArgs{
				Image: image,
			},
		},

		Tags: pulumi.StringArray{
			pulumi.String("worker"),
			pulumi.String("kubernetes"),
		},

		NetworkInterfaces: compute.InstanceNetworkInterfaceArray{
			&compute.InstanceNetworkInterfaceArgs{
				Network:    network.Name,
				Subnetwork: subnetwork.Name,
				AccessConfigs: compute.InstanceNetworkInterfaceAccessConfigArray{
					&compute.InstanceNetworkInterfaceAccessConfigArgs{
						NetworkTier: networkTier,
					},
				},
			},
		},
		ServiceAccount: &compute.InstanceServiceAccountArgs{
			Email: service_account.Email,
			Scopes: pulumi.StringArray{
				pulumi.String("https://www.googleapis.com/auth/cloud-platform"),
			},
		},
		MetadataStartupScript: workerCommand,
	}, pulumi.DependsOn([]pulumi.Resource{lastInstance}))
	if err != nil {
		return nil, err
	}

	return instance, nil

}

func GenerateMasterMachine(ctx *pulumi.Context, workerNumber int, instanceName string, network *compute.Network, subnetwork *compute.Subnetwork, bucket *storage.Bucket, publicKey string, service_account *serviceaccount.Account) (*compute.Instance, error) {
	var machineType pulumi.String = "e2-small"
	var region pulumi.String = "us-central1-a"
	var image pulumi.String = "ubuntu-os-cloud/ubuntu-2004-lts"
	var networkTier pulumi.String = "STANDARD"

	masterCommand := pulumi.All(workerNumber, bucket.Name).ApplyT(func(args []interface{}) (string, error) {
		workerNum := args[0].(int)
		bucketName := args[1].(string)

		return fmt.Sprintf(`#!/bin/bash
			LOG_FILE="/var/log/startup_script.log"
			exec > >(tee -a $LOG_FILE) 2>&1
			echo "Starting script execution" | tee -a $LOG_FILE
			sudo apt-get update | tee -a $LOG_FILE
			sudo snap install microk8s --classic --channel=1.31 | tee -a $LOG_FILE
			sudo usermod -aG microk8s $(whoami) | tee -a $LOG_FILE
			newgrp microk8s | tee -a $LOG_FILE
			if sudo microk8s status --wait-ready; then
        echo "microk8s is ready" | tee -a $LOG_FILE
        touch /tmp/microk8s-ready.lock
        break
			else
					echo "Waiting for microk8s to be ready"
					sleep 10
			fi
			sudo microk8s enable dns | tee -a $LOG_FILE
			for i in $(seq 1 %d); do
				JOIN_COMMAND=$(sudo microk8s add-node | grep 'microk8s join' | sed -n '2p')
				if [ -n "$JOIN_COMMAND" ]; then
					echo $JOIN_COMMAND > /home/join-command-$i.txt
					gsutil cp /home/join-command-$i.txt gs://%s/join-command-$i.txt
					echo "Join command for node $i stored and uploaded." | tee -a $LOG_FILE
				fi
			done
			echo "AcceptEnv PULUMI_COMMAND_STDOUT" >> /etc/ssh/sshd_config
			sudo systemctl restart sshd

			echo "Script execution completed" | tee -a $LOG_FILE
			`, workerNum, bucketName), nil
	}).(pulumi.StringOutput)

	instance, err := compute.NewInstance(ctx, instanceName, &compute.InstanceArgs{
		MachineType: machineType,
		Zone:        region,

		BootDisk: &compute.InstanceBootDiskArgs{
			InitializeParams: &compute.InstanceBootDiskInitializeParamsArgs{
				Image: image,
			},
		},

		Tags: pulumi.StringArray{
			pulumi.String("master"),
			pulumi.String("kubernetes"),
		},

		NetworkInterfaces: compute.InstanceNetworkInterfaceArray{
			&compute.InstanceNetworkInterfaceArgs{
				Network:    network.Name,
				Subnetwork: subnetwork.Name,
				AccessConfigs: compute.InstanceNetworkInterfaceAccessConfigArray{
					&compute.InstanceNetworkInterfaceAccessConfigArgs{
						NetworkTier: networkTier,
					},
				},
			},
		},
		ServiceAccount: &compute.InstanceServiceAccountArgs{
			Email: service_account.Email,
			Scopes: pulumi.StringArray{
				pulumi.String("https://www.googleapis.com/auth/cloud-platform"),
			},
		},
		Metadata: pulumi.StringMap{
			"ssh-keys": pulumi.String("pulumi:" + publicKey),
		},
		MetadataStartupScript: masterCommand,
	}, pulumi.DependsOn([]pulumi.Resource{bucket}))
	if err != nil {
		return nil, err
	}

	return instance, nil

}
