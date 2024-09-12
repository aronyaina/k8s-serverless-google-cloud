package main

import (
	"k8s-serverless/connexion"
	"k8s-serverless/gcp/access"
	"k8s-serverless/gcp/infra"
	"k8s-serverless/gcp/network"
	"log"
	"strconv"

	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const VM_NUMBER = 0

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		privateKeyPath := "./private-key.pem"
		privateKey, publicKey, err := connexion.GenerateSSHKeyPair(privateKeyPath)
		if err != nil {
			log.Println("Failed to generate SSH key pair:", err)
			return err
		}

		network1, err := network.GeneratePrivateNetwork(ctx)
		if err != nil {
			log.Println("error While creating network")
			return err
		}

		err = network.GenerateFirewall(ctx, network1)
		if err != nil {
			log.Println("error While creating firewall")
			return err
		}

		subnet, err := network.GenerateSubNetwork(ctx, network1)
		if err != nil {
			log.Println("error while creating subnet")
			return err
		}

		serviceAccount, err := access.GenerateServiceAccount(ctx)
		if err != nil {
			log.Println("error while creating service account")
			return err
		}

		bucket, err := infra.GenerateBucket(ctx, "my-bucket-329102")
		if err != nil {
			log.Println("error while creating bucket")
			return err
		}

		_, err = access.GenerateIamMember(ctx, serviceAccount)
		if err != nil {
			log.Println("error while creating iam member")
			return err
		}

		_, err = access.GenerateIamBindingOfBucket(ctx, bucket, serviceAccount)
		if err != nil {
			log.Println("error while creating iam binding")
			return err
		}

		masterMachine, err := infra.GenerateMasterMachine(ctx, VM_NUMBER, "master-node", network1, subnet, bucket, publicKey, serviceAccount)
		if err != nil {
			log.Println("error while creating master machine")
			return err
		}

		//masterExternalIp := masterMachine.NetworkInterfaces.Index(pulumi.Int(0)).AccessConfigs().Index(pulumi.Int(0)).NatIp()

		//connectionArgs := masterExternalIp.ApplyT(func(ip *string) (remote.ConnectionArgs, error) {
		//	if ip == nil {
		//		return remote.ConnectionArgs{}, fmt.Errorf("masterExternalIp is nil")
		//	}

		//	// Return the connection args
		//	return remote.ConnectionArgs{
		//		Host:       pulumi.String(*ip),
		//		User:       pulumi.String("pulumi"),
		//		PrivateKey: pulumi.String(privateKey),
		//	}, nil
		//}).(pulumi.Output).ApplyT(func(ca interface{}) remote.ConnectionArgs {
		//	return ca.(remote.ConnectionArgs)
		//}).(pulumi.Output)

		//getKubeconfig, err := remote.NewCommand(ctx, "getKubeconfig", &remote.CommandArgs{
		//	Create: pulumi.String("microk8s config"),
		//	Connection: connectionArgs.ApplyT(func(ca remote.ConnectionArgs) *remote.ConnectionArgs {
		//		return &ca
		//	}).(remote.ConnectionInput),
		//})

		//TODO: not working
		//kubeconfig := masterExternalIp.ApplyT(func(ip *string) (interface{}, error) {
		//	// Make sure the IP is not nil before using it
		//	if ip == nil {
		//		return nil, fmt.Errorf("masterExternalIp is nil")
		//	}
		//	hostnameCmdArgs := &remote.CommandArgs{
		//		Create: pulumi.String("sudo microk8s config"),
		//		Connection: &remote.ConnectionArgs{
		//			Host:       pulumi.String(*ip), // Convert *string to pulumi.String
		//			User:       pulumi.String("pulumi"),
		//			PrivateKey: pulumi.String(privateKey),
		//		},
		//	}
		//	// Run the command with the extracted IP
		//	hostnameCmd, err := remote.NewCommand(ctx, "hostnameCmd", hostnameCmdArgs)
		//	if err != nil {
		//		return nil, err
		//	}
		//	fmt.Println("hostname stdout", hostnameCmd.Stdout)
		//	hostnameCmd.Stdout.ApplyT(func(result interface{}) (interface{}, error) {
		//		if results, ok := result.([]interface{}); ok {
		//			for _, res := range results {
		//				if hostname, ok := res.(string); ok {
		//					fmt.Printf("Hostname: %s\n", hostname)
		//				}
		//			}
		//		} else {
		//			return nil, fmt.Errorf("unexpected type for Stdout result")
		//		}
		//		return nil, nil
		//	})
		//	return hostnameCmd, nil
		//})
		//if err != nil {
		//	return err
		//}
		//hostnameCmd, err := remote.NewCommand(ctx, "hostnameCmd", hostnameCmdArgs)
		//if err != nil {
		//	return nil, err
		//}
		//hostnameCmd.Stdout.ApplyT(func(result interface{}) (interface{}, error) {
		//	// Safely assert the result as []interface{} and handle each part
		//	if results, ok := result.([]interface{}); ok {
		//		for _, res := range results {
		//			if hostname, ok := res.(string); ok {
		//				fmt.Printf("Hostname: %s\n", hostname)
		//			}
		//		}
		//	} else {
		//		return nil, fmt.Errorf("unexpected type for Stdout result")
		//	}
		//	return nil, nil
		//})

		var allMachines []*compute.Instance
		var last *compute.Instance = masterMachine

		if VM_NUMBER >= 1 {
			for i := 1; i <= VM_NUMBER; i++ {
				machine, err := infra.GenerateWorkerMachine(ctx, i, last, "worker-node-"+strconv.Itoa(i), network1, subnet, bucket, serviceAccount)

				if err != nil {
					return err
				}

				// Ensure the next worker depends on the current one
				last = machine
				allMachines = append(allMachines, machine)
			}
		}

		//provider, err := k8s.GenerateProvider(ctx, machine)
		//if err != nil {
		//	return err
		//}

		//ctx.Export("kubeconfig", kubeconfig)

		ctx.Export("serviceAccountName", serviceAccount.Name)
		ctx.Export("BucketName", bucket.Name)

		for i, machine := range allMachines {
			ctx.Export("instanceName"+strconv.Itoa(i+1), machine.Name) // Add an index to make export names unique
			ctx.Export("instanceExternalIP"+strconv.Itoa(i+1), machine.NetworkInterfaces.Index(pulumi.Int(0)).AccessConfigs().Index(pulumi.Int(0)).NatIp())
			ctx.Export("instanceInternalIP"+strconv.Itoa(i+1), machine.NetworkInterfaces.Index(pulumi.Int(0)).NetworkIp())
		}
		ctx.Export("masterName", masterMachine.Name) // Add an index to make export names unique
		ctx.Export("masterExternalIP", masterMachine.NetworkInterfaces.Index(pulumi.Int(0)).AccessConfigs().Index(pulumi.Int(0)).NatIp())
		ctx.Export("masterInternalIP", masterMachine.NetworkInterfaces.Index(pulumi.Int(0)).NetworkIp())

		ctx.Export("masterMachinePrivateKey", pulumi.String(privateKey))
		ctx.Export("masterMachinePublicKey", pulumi.String(publicKey))

		ctx.Export("networkName", network1.Name)
		ctx.Export("subnetName", subnet.Name)
		return nil
	})
}
