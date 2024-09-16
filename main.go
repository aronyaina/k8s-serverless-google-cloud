package main

import (
	"k8s-serverless/gcp/service/kube"
	"k8s-serverless/k8s/config"
	"k8s-serverless/k8s/repository"
	"log"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const WORKER_NUMBER = 2

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		masterMachine, workerMachines, masterPrivateKey, workerPrivateKey, bucket, err := kube.GenerateMicroInfra(ctx, WORKER_NUMBER)
		if err != nil {
			log.Println("Error While generating microk8s infra")
			return err
		}

		triggers := pulumi.Array{bucket}

		if err = kube.ConnectMicroInfra(ctx, masterMachine, workerMachines, WORKER_NUMBER, bucket, masterPrivateKey, workerPrivateKey, triggers); err != nil {
			log.Println("Error While connecting microk8s infra")
			return err
		}

		provider, err := config.GenerateProviderFromConfig(ctx, masterMachine, masterPrivateKey)
		if err != nil {
			return err
		}

		namespace, err := repository.GenerateNamespace(ctx, "dev", "k8s-serverless", provider, []pulumi.Resource{provider})
		if err != nil {
			return err
		}

		labels := pulumi.StringMap{"app": pulumi.String("nginx")}
		service, err := repository.GenerateService(ctx, labels, provider, []pulumi.Resource{namespace, provider})
		if err != nil {
			return err
		}
		_, err = repository.GenerateDeployment(ctx, labels, "nginx", provider, []pulumi.Resource{namespace, provider, service})
		if err != nil {
			return err
		}

		return nil
	})
}
