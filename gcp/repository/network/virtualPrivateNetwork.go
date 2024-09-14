package network

import (
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func GeneratePrivateNetwork(ctx *pulumi.Context) (*compute.Network, error) {
	var vpcName string = "my-vpc-network"
	var autoCreateSubNetworks pulumi.Bool = false

	network, err := compute.NewNetwork(ctx, vpcName, &compute.NetworkArgs{
		AutoCreateSubnetworks: autoCreateSubNetworks,
	})

	if err != nil {
		return nil, err
	}
	return network, nil
}

func GenerateSubNetwork(ctx *pulumi.Context, network *compute.Network) (*compute.Subnetwork, error) {
	var subnetworkName string = "my-vpc-subnetwork"
	var ipAddr pulumi.String = "10.0.0.0/24"
	var region pulumi.String = "us-central1"
	var purpose pulumi.String = "PRIVATE"

	subnetwork, err := compute.NewSubnetwork(ctx, subnetworkName, &compute.SubnetworkArgs{
		IpCidrRange: ipAddr,
		Name:        pulumi.String(subnetworkName),
		Network:     network.ID(),
		Region:      region,
		Purpose:     purpose,
	}, pulumi.DependsOn([]pulumi.Resource{network}))

	if err != nil {
		return nil, err
	}
	return subnetwork, nil
}
