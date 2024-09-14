package main

import (
	"k8s-serverless/gcp/service/kube"
	"k8s-serverless/k8s/config"
	"k8s-serverless/k8s/ressource"
	"log"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		masterMachine, _, masterPrivateKey, _, err := kube.GenerateMicroInfra(ctx, 1)
		if err != nil {
			log.Println("Error While generating microk8s infra")
			return err
		}
		provider, err := config.GenerateProviderFromConfig(ctx, masterMachine, masterPrivateKey)

		namespace, err := ressource.GenerateNameSpace(ctx, "k8s-serverless", "dev", provider)

		ctx.Export("Namespace", namespace)

		return nil
	})
}
