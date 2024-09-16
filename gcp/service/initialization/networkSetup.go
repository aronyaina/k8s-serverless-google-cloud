package initialization

import (
	"k8s-serverless/gcp/repository/network"
	"log"

	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func GenerateNetworkInfra(ctx *pulumi.Context) (*compute.Network, *compute.Subnetwork, error) {
	network1, err := network.GeneratePrivateNetwork(ctx)
	if err != nil {
		log.Println("error While creating network")
		return nil, nil, err
	}

	err = network.GenerateFirewall(ctx, network1)
	if err != nil {
		log.Println("error While creating firewall")
		return nil, nil, err
	}

	subnet, err := network.GenerateSubNetwork(ctx, network1)
	if err != nil {
		log.Println("error while creating subnet")
		return nil, nil, err
	}

	ctx.Export("networkName", network1.Name)
	ctx.Export("subnetName", subnet.Name)
	return network1, subnet, err
}
