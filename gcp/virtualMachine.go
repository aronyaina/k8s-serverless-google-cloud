package gcp

import (
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func GenerateVirtualMachine(ctx *pulumi.Context, instanceName string, network *compute.Network, subnetwork *compute.Subnetwork) (*compute.Instance, error) {
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
	})
	if err != nil {
		return nil, err
	}

	return instance, nil

}
