package main

import (
	"k8s-serverless/gcp"
	"log"
	"strconv"

	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const VM_NUMBER = 2

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		network, err := gcp.GeneratePrivateNetwork(ctx)
		if err != nil {
			log.Println("error While creating network")
			return err
		}

		err = gcp.GenerateFirewall(ctx, network)
		if err != nil {
			log.Println("error While creating firewall")
			return err
		}

		subnet, err := gcp.GenerateSubNetwork(ctx, network)
		if err != nil {
			log.Println("error while creating subnet")
			return err
		}

		serviceAccount, err := gcp.GenerateServiceAccount(ctx)
		if err != nil {
			log.Println("error while creating service account")
			return err
		}

		bucket, err := gcp.GenerateBucket(ctx, "my-bucket-329102")
		if err != nil {
			log.Println("error while creating bucket")
			return err
		}

		_, err = gcp.GenerateIamMember(ctx, serviceAccount)
		if err != nil {
			log.Println("error while creating iam member")
			return err
		}

		_, err = gcp.GenerateIamBindingOfBucket(ctx, bucket, serviceAccount)
		if err != nil {
			log.Println("error while creating iam binding")
			return err
		}

		masterMachine, err := gcp.GenerateMasterMachine(ctx, "master-node", network, subnet, bucket, serviceAccount)
		if err != nil {
			log.Println("error while creating master machine")
			return err
		}

		var allMachines []*compute.Instance
		for i := 1; i <= VM_NUMBER; i++ {
			machine, err := gcp.GenerateWorkerMachine(ctx, masterMachine, "worker-node-"+strconv.Itoa(i), network, subnet, bucket, serviceAccount)

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

		ctx.Export("serviceAccountName", serviceAccount.Name)
		ctx.Export("BucketName", bucket.Name)

		for i, machine := range allMachines {
			ctx.Export("instanceName"+strconv.Itoa(i+1), machine.Name) // Add an index to make export names unique
			ctx.Export("instanceExternalIP"+strconv.Itoa(i+1), machine.NetworkInterfaces.Index(pulumi.Int(0)).AccessConfigs().Index(pulumi.Int(0)).NatIp())
			ctx.Export("instanceInternalIP"+strconv.Itoa(i+1), machine.NetworkInterfaces.Index(pulumi.Int(0)).NetworkIp())
		}

		ctx.Export("tokenBucket", bucket.Name)
		ctx.Export("masterName", masterMachine.Name) // Add an index to make export names unique
		ctx.Export("masterExternalIP", masterMachine.NetworkInterfaces.Index(pulumi.Int(0)).AccessConfigs().Index(pulumi.Int(0)).NatIp())
		ctx.Export("masterInternalIP", masterMachine.NetworkInterfaces.Index(pulumi.Int(0)).NetworkIp())

		ctx.Export("networkName", network.Name)
		ctx.Export("subnetName", subnet.Name)
		return nil
	})
}
