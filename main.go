package main

import (
	"fmt"
	gcpinfra "k8s-serverless/gcp/infra"
	gcpservice "k8s-serverless/gcp/service"
	k8sconfig "k8s-serverless/k8s/config"
	"log"
	"strconv"

	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/compute"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const VM_NUMBER = 1

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		privateKey, publicKey, err := gcpservice.GenerateSshIfItDoesntExist()
		serviceAccount, bucket, err := gcpservice.GenerateServiceLinkedToBucket(ctx)
		network, subnet, err := gcpservice.GenerateNetworkInfra(ctx)

		masterMachine, err := gcpinfra.GenerateMasterMachine(ctx, VM_NUMBER, "master-node", network, subnet, bucket, serviceAccount, publicKey)
		if err != nil {
			log.Println("error while creating master machine")
			return err
		}

		var allMachines []*compute.Instance
		var last *compute.Instance = masterMachine

		if VM_NUMBER >= 1 {
			for i := 1; i <= VM_NUMBER; i++ {
				machine, err := gcpinfra.GenerateWorkerMachine(ctx, i, last, "worker-node-"+strconv.Itoa(i), network, subnet, bucket, serviceAccount)

				if err != nil {
					return err
				}
				last = machine
				allMachines = append(allMachines, machine)
			}
		}

		kubeConfig, err := k8sconfig.GenerateMasterKubeConfig(ctx, masterMachine, privateKey)
		if err != nil {
			return err
		}

		k8sConfig := pulumi.Output(kubeConfig).ApplyT(func(result interface{}) (string, error) {
			if resultStr, ok := result.(string); ok {
				return resultStr, nil
			}
			return "", fmt.Errorf("unexpected type for kubeconfig result")
		}).(pulumi.StringOutput)

		provider, err := kubernetes.NewProvider(ctx, "master-1", &kubernetes.ProviderArgs{
			Kubeconfig: k8sConfig,
		})

		if err != nil {
			return err
		}
		ctx.Export("Provider", provider)
		ctx.Export("Kubeconfig", k8sConfig)

		for i, machine := range allMachines {
			ctx.Export("instanceName"+strconv.Itoa(i+1), machine.Name)
			ctx.Export("instanceExternalIP"+strconv.Itoa(i+1), machine.NetworkInterfaces.Index(pulumi.Int(0)).AccessConfigs().Index(pulumi.Int(0)).NatIp())
			ctx.Export("instanceInternalIP"+strconv.Itoa(i+1), machine.NetworkInterfaces.Index(pulumi.Int(0)).NetworkIp())
		}
		ctx.Export("masterName", masterMachine.Name)
		ctx.Export("masterExternalIP", masterMachine.NetworkInterfaces.Index(pulumi.Int(0)).AccessConfigs().Index(pulumi.Int(0)).NatIp())
		ctx.Export("masterInternalIP", masterMachine.NetworkInterfaces.Index(pulumi.Int(0)).NetworkIp())

		return nil
	})
}
