package main

import (
	"k8s-serverless/gcp"

	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		network, err := gcp.GeneratePrivateNetwork(ctx)
		if err != nil {
			return err
		}

		subnet, err := gcp.GenerateSubNetwork(ctx, network)
		if err != nil {
			return err
		}

		machine, err := gcp.GenerateVirtualMachine(ctx, subnet)
		if err != nil {
			return err
		}

		ctx.Export("instanceName", machine.Name)
		ctx.Export("instanceIP", machine.NetworkInterfaces.ApplyT(func(nis []compute.InstanceNetworkInterface) *string {
			return nis[0].AccessConfigs[0].NatIp
		}).(pulumi.StringOutput))
		ctx.Export("networkName", network.Name)
		ctx.Export("subnetName", subnet.Name)
		return nil
	})
}
