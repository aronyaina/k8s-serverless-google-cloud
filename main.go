package main

import (
	"k8s-serverless/gcp/service/kube"
	"k8s-serverless/k8s/config"
	"k8s-serverless/k8s/ressource"
	"log"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const WORKER_NUMBER = 1

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		masterMachine, workerMachines, masterPrivateKey, workerPrivateKey, bucket, err := kube.GenerateMicroInfra(ctx, WORKER_NUMBER)
		if err != nil {
			log.Println("Error While generating microk8s infra")
			return err
		}

		triggers := pulumi.Array{masterMachine, bucket}

		if err = kube.ConnectMicroInfra(ctx, masterMachine, workerMachines, WORKER_NUMBER, bucket, masterPrivateKey, workerPrivateKey, triggers); err != nil {
			log.Println("Error While connecting microk8s infra")
			return err
		}

		provider, err := config.GenerateProviderFromConfig(ctx, masterMachine, masterPrivateKey)
		if err != nil {
			return err
		}

		namespace, err := ressource.GenerateNameSpace(ctx, "k8s-serverless", "dev", provider)
		if err != nil {
			return err
		}

		ctx.Export("Namespace", namespace)

		return nil
	})
}
