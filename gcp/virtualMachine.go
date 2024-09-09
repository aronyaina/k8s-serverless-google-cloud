package gcp

import (
	"fmt"

	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/compute"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/serviceaccount"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/storage"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func GenerateWorkerMachine(ctx *pulumi.Context, workerIndice int, masterNode *compute.Instance, instanceName string, network *compute.Network, subnetwork *compute.Subnetwork, bucket *storage.Bucket, service_account *serviceaccount.Account) (*compute.Instance, error) {
	var machineType pulumi.String = "e2-micro"
	var region pulumi.String = "us-central1-a"
	var image pulumi.String = "ubuntu-os-cloud/ubuntu-2004-lts"
	var networkTier pulumi.String = "STANDARD"

	worker_command := pulumi.All(workerNumber, bucket.Name).ApplyT(func(args []interface{}) (string, error) {
		workerNum := args[0].(int)
		bucketName := args[1].(string)

		return fmt.Sprintf(`#!/bin/bash
			LOG_FILE="/var/log/startup_script.log"
			exec > >(tee -a $LOG_FILE) 2>&1

			echo "Starting script execution" | tee -a $LOG_FILE
			sudo apt-get update | tee -a $LOG_FILE
			sudo snap install microk8s --classic --channel=1.31 | tee -a $LOG_FILE
			sudo usermod -aG microk8s $USER | tee -a $LOG_FILE
			newgrp microk8s | tee -a $LOG_FILE
			while ! sudo microk8s status --wait-ready; do
				echo "waiting for microk8s to be ready" | tee -a $LOG_FILE
				sleep 5
			done
			sudo microk8s enable dns | tee -a $LOG_FILE
			for i in $(seq 1 $((%d))); do
				JOIN_COMMAND=$(sudo microk8s add-node | grep 'microk8s join' | sed -n '2p')
				if [ -n "$JOIN_COMMAND" ]; then
					echo $JOIN_COMMAND > /home/join-command-$i.txt
					gsutil cp /home/join-command-$i.txt gs://%s/join-command-$i.txt
					echo "Join command for node $i stored and uploaded." | tee -a $LOG_FILE
				fi
			done
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
		MetadataStartupScript: pulumi.String(fmt.Sprintf(`#!/bin/bash
				sudo apt-get update
				sudo snap install microk8s --classic --channel=1.31
				sudo usermod -aG microk8s $USER
				newgrp microk8s
				while !sudo microk8s status --wait-ready; do
					echo "waiting for microk8s to be ready"
					sleep 5
			  done
				sudo microk8s enable dns
				gsutil cp gs://%s/join-command-%d.txt /home/join-command.txt
				JOIN_COMMAND=$(cat /home/join-command.txt)
				sudo $JOIN_COMMAND
			`, bucket.Name, workerIndice)),
	}, pulumi.DependsOn([]pulumi.Resource{masterNode}))
	if err != nil {
		return nil, err
	}

	return instance, nil

}

func GenerateMasterMachine(ctx *pulumi.Context, workerNumber int, instanceName string, network *compute.Network, subnetwork *compute.Subnetwork, bucket *storage.Bucket, service_account *serviceaccount.Account) (*compute.Instance, error) {
	var machineType pulumi.String = "e2-small"
	var region pulumi.String = "us-central1-a"
	var image pulumi.String = "ubuntu-os-cloud/ubuntu-2004-lts"
	var networkTier pulumi.String = "STANDARD"
	//var command string = fmt.Sprintf(`#!/bin/bash
	//			sudo apt-get update
	//			sudo snap install microk8s --classic --channel=1.31
	//			sudo usermod -aG microk8s $USER
	//			newgrp microk8s
	//			while ! sudo microk8s status --wait-ready; do
	//					echo "waiting for microk8s to be ready"
	//					sleep 5
	//			done
	//			sudo microk8s enable dns
	//			for i in $(seq 1 $((%d))); do
	//				JOIN_COMMAND=$(sudo microk8s add-node | grep 'microk8s join' | sed -n '2p')
	//				if [ -n "$JOIN_COMMAND" ]; then
	//					echo $JOIN_COMMAND > /home/join-command-$i.txt
	//					gsutil cp /home/join-command-$i.txt gs://%s/join-command-$i.txt
	//				fi
	//			done
	//			`, workerNumber, bucket.Name)

	master_command := pulumi.All(workerNumber, bucket.Name).ApplyT(func(args []interface{}) (string, error) {
		workerNum := args[0].(int)
		bucketName := args[1].(string)

		return fmt.Sprintf(`#!/bin/bash
			LOG_FILE="/var/log/startup_script.log"
			exec > >(tee -a $LOG_FILE) 2>&1

			echo "Starting script execution" | tee -a $LOG_FILE
			sudo apt-get update | tee -a $LOG_FILE
			sudo snap install microk8s --classic --channel=1.31 | tee -a $LOG_FILE
			sudo usermod -aG microk8s $USER | tee -a $LOG_FILE
			newgrp microk8s | tee -a $LOG_FILE
			while ! sudo microk8s status --wait-ready; do
				echo "waiting for microk8s to be ready" | tee -a $LOG_FILE
				sleep 5
			done
			sudo microk8s enable dns | tee -a $LOG_FILE
			for i in $(seq 1 $((%d))); do
				JOIN_COMMAND=$(sudo microk8s add-node | grep 'microk8s join' | sed -n '2p')
				if [ -n "$JOIN_COMMAND" ]; then
					echo $JOIN_COMMAND > /home/join-command-$i.txt
					gsutil cp /home/join-command-$i.txt gs://%s/join-command-$i.txt
					echo "Join command for node $i stored and uploaded." | tee -a $LOG_FILE
				fi
			done
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
		MetadataStartupScript: master_command,
	}, pulumi.DependsOn([]pulumi.Resource{bucket}))
	if err != nil {
		return nil, err
	}

	return instance, nil

}
