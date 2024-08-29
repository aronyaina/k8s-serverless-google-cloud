package main

import (
	"k8s-serverless/gcp"
	"strconv"

	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const VM_NUMBER = 2

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		network, err := gcp.GeneratePrivateNetwork(ctx)
		if err != nil {
			return err
		}

		//err = gcp.GenerateFirewall(ctx, network)
		//if err != nil {
		//	return err
		//}

		subnet, err := gcp.GenerateSubNetwork(ctx, network)
		if err != nil {
			return err
		}

		var allMachines []*compute.Instance
		for i := 1; i <= VM_NUMBER; i++ {
			machine, err := gcp.GenerateVirtualMachine(ctx, "vm-"+strconv.Itoa(i), network, subnet)

			if err != nil {
				return err
			}
			allMachines = append(allMachines, machine)
		}

		//provider, err := k8s.GenerateProvider(ctx, machine)
		//if err != nil {
		//	return err
		//}

		//ctx.Export("ProviderID", provider.ID())

		for i, machine := range allMachines {
			ctx.Export("instanceName"+strconv.Itoa(i+1), machine.Name) // Add an index to make export names unique
			ctx.Export("instanceExternalIP"+strconv.Itoa(i+1), machine.NetworkInterfaces.Index(pulumi.Int(0)).AccessConfigs().Index(pulumi.Int(0)).NatIp())
			ctx.Export("instanceInternalIP"+strconv.Itoa(i+1), machine.NetworkInterfaces.Index(pulumi.Int(0)).NetworkIp())
		}
		ctx.Export("networkName", network.Name)
		ctx.Export("subnetName", subnet.Name)
		return nil
	})
}
