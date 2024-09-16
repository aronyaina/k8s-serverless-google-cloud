package network

import (
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func GenerateAddress(ctx *pulumi.Context) (*compute.Address, error) {
	address, err := compute.NewAddress(ctx, "address", &compute.AddressArgs{
		Name:        pulumi.String("my-internal-address"),
		AddressType: pulumi.String("INTERNAL"),
		Purpose:     pulumi.String("GCE_ENDPOINT"),
	})

	if err != nil {
		return nil, err
	}
	return address, nil

}
