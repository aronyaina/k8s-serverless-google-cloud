package infrastructure

import (
	"fmt"

	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/compute"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/serviceaccount"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/storage"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const MACHINE_TYPE pulumi.String = "e2-small"
const REGION pulumi.String = "us-central1-a"
const SYSTEM_IMAGE pulumi.String = "ubuntu-os-cloud/ubuntu-2004-lts"
const NETWORK_TIER pulumi.String = "STANDARD"

func GenerateWorkerMachine(ctx *pulumi.Context, workerIndex int, lastInstance *compute.Instance, instanceName string, network *compute.Network, subnetwork *compute.Subnetwork, bucket *storage.Bucket, service_account *serviceaccount.Account, publicKey string) (*compute.Instance, error) {
	//TODO: Implement the k8s script
	workerCommand := pulumi.String(fmt.Sprintf(`#!/bin/bash
			LOG_FILE="/var/log/startup_script.log"
			exec > >(tee -a $LOG_FILE) 2>&1
			sudo apt-get update | tee -a $LOG_FILE
			sudo snap install microk8s --classic --channel=1.31 | tee -a $LOG_FILE
			sudo usermod -aG microk8s $(whoami) | tee -a $LOG_FILE
			sudo usermod -aG microk8s pulumi | tee -a $LOG_FILE
			newgrp microk8s 
			sudo microk8s enable dns | tee -a $LOG_FILE
			echo "AcceptEnv PULUMI_COMMAND_STDOUT" >> /etc/ssh/sshd_config
			sudo systemctl restart sshd
			if sudo microk8s status --wait-ready; then
        echo "microk8s is ready" | tee -a $LOG_FILE
        touch /tmp/k8sready.lock
			else
					echo "Waiting for microk8s to be ready"
					sleep 10
			fi
			`))

	instance, err := compute.NewInstance(ctx, instanceName, &compute.InstanceArgs{
		MachineType: MACHINE_TYPE,
		Zone:        REGION,

		BootDisk: &compute.InstanceBootDiskArgs{
			InitializeParams: &compute.InstanceBootDiskInitializeParamsArgs{
				Image: SYSTEM_IMAGE,
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
						NetworkTier: NETWORK_TIER,
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
		MetadataStartupScript: workerCommand,
	}, pulumi.DependsOn([]pulumi.Resource{lastInstance}))

	if err != nil {
		return nil, err
	}

	return instance, nil

}

func GenerateMasterMachine(ctx *pulumi.Context, workerNumber int, instanceName string, network *compute.Network, subnetwork *compute.Subnetwork, bucket *storage.Bucket, service_account *serviceaccount.Account, publicKey string) (*compute.Instance, error) {
	masterCommand := pulumi.String(fmt.Sprintf(`#!/bin/bash
			LOG_FILE="/var/log/startup_script.log"
			exec > >(tee -a $LOG_FILE) 2>&1
			echo "Starting script execution" | tee -a $LOG_FILE
			sudo apt-get update | tee -a $LOG_FILE
			sudo snap install microk8s --classic --channel=1.31 | tee -a $LOG_FILE
			sudo usermod -aG microk8s $(whoami) | tee -a $LOG_FILE
			sudo usermod -aG microk8s pulumi | tee -a $LOG_FILE
			newgrp microk8s | tee -a $LOG_FILE
			sudo microk8s enable dns | tee -a $LOG_FILE
			sudo microk8s enable ha-cluster
			echo "AcceptEnv PULUMI_COMMAND_STDOUT" >> /etc/ssh/sshd_config
			sudo systemctl restart sshd
			if sudo microk8s status --wait-ready; then
        echo "microk8s is ready" | tee -a $LOG_FILE
        touch /tmp/k8sready.lock
			else
					echo "Waiting for microk8s to be ready"
					sleep 10
			fi
			`))

	instance, err := compute.NewInstance(ctx, instanceName, &compute.InstanceArgs{
		MachineType: MACHINE_TYPE,
		Zone:        REGION,

		BootDisk: &compute.InstanceBootDiskArgs{
			InitializeParams: &compute.InstanceBootDiskInitializeParamsArgs{
				Image: SYSTEM_IMAGE,
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
						NetworkTier: NETWORK_TIER,
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

	ctx.Export("masterName", instance.Name)
	ctx.Export("masterExternalIP", instance.NetworkInterfaces.Index(pulumi.Int(0)).AccessConfigs().Index(pulumi.Int(0)).NatIp())
	ctx.Export("masterInternalIP", instance.NetworkInterfaces.Index(pulumi.Int(0)).NetworkIp())
	return instance, nil

}
