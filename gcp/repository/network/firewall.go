package network

import (
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func GenerateFirewall(ctx *pulumi.Context, network *compute.Network) error {
	protocolsAndPorts := map[string][]string{
		"tcp": {
			"22",
			"53",
			"80",
			"443",
			"2379",
			"10250",
			"10257",
			"10259",
			"16443",
			"25000",
			"30100",
		},
		"icmp": {},
	}

	for protocol, ports := range protocolsAndPorts {
		_, err := compute.NewFirewall(ctx, "firewall-"+string(protocol), &compute.FirewallArgs{
			Network: network.ID(), // Ensure the network ID is set correctly
			Allows: compute.FirewallAllowArray{
				&compute.FirewallAllowArgs{
					Protocol: pulumi.String(protocol),
					Ports:    pulumi.StringArray(pulumi.ToStringArray(ports)),
				},
			},
			Direction: pulumi.String("INGRESS"), // Specify the direction of the traffic
			Priority:  pulumi.Int(1000),
			SourceRanges: pulumi.StringArray{
				pulumi.String("0.0.0.0/0"), // Allow from any source IP range
			},
		})

		if err != nil {
			return err
		}
	}

	return nil
}
