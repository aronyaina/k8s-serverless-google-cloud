package gcp

import (
	"fmt"

	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/compute"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/serviceaccount"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/storage"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func GenerateWorkerMachine(ctx *pulumi.Context, masterNode *compute.Instance, instanceName string, network *compute.Network, subnetwork *compute.Subnetwork, bucket *storage.Bucket, service_account *serviceaccount.Account) (*compute.Instance, error) {
	var machineType pulumi.String = "e2-micro"
	var region pulumi.String = "us-central1-a"
	var image pulumi.String = "ubuntu-os-cloud/ubuntu-2004-lts"
	var networkTier pulumi.String = "STANDARD"

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
				
				while ! sudo microk8s status --wait-ready; do
					echo "Waiting for MicroK8s to be ready..."
					sleep 5
			  done

				sudo microk8s enable dns

				gsutil cp gs://%s/join-command.txt /home/$USER/join-command.txt
				JOIN_COMMAND=$(cat /tmp/join-command.txt)
				sudo $JOIN_COMMAND
			`, bucket.Name)),
	}, pulumi.DependsOn([]pulumi.Resource{masterNode}))
	if err != nil {
		return nil, err
	}

	return instance, nil

}

func GenerateMasterMachine(ctx *pulumi.Context, instanceName string, network *compute.Network, subnetwork *compute.Subnetwork, bucket *storage.Bucket, service_account *serviceaccount.Account) (*compute.Instance, error) {
	var machineType pulumi.String = "e2-micro"
	var region pulumi.String = "us-central1-a"
	var image pulumi.String = "ubuntu-os-cloud/ubuntu-2004-lts"
	var networkTier pulumi.String = "STANDARD"

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
		MetadataStartupScript: pulumi.String(fmt.Sprintf(`#!/bin/bash
				sudo apt-get update
				sudo snap install microk8s --classic --channel=1.31
				sudo usermod -aG microk8s $USER
				newgrp microk8s
				# Wait until MicroK8s is ready
				while ! sudo microk8s status --wait-ready; do
						echo "Waiting for MicroK8s to be ready..."
						sleep 5
				done
				sudo microk8s enable dns
				JOIN_COMMAND=$(sudo microk8s add-node | grep 'microk8s join' | sed -n '2p')

				echo $JOIN_COMMAND > /home/$USER/join-command.txt
				
				# Copy the join command to the GCS bucket
				gsutil cp /home/$USER/join-command.txt gs://%s/join-command.txt
				`, bucket.Name)),
	}, pulumi.DependsOn([]pulumi.Resource{bucket}))
	if err != nil {
		return nil, err
	}

	return instance, nil

}
