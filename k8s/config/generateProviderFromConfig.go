package config

import (
	"fmt"
	"k8s-serverless/utils"

	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/compute"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func GenerateProviderFromConfig(ctx *pulumi.Context, masterMachine *compute.Instance, privateKey string) (*kubernetes.Provider, error) {
	kubeConfig, err := GenerateMasterKubeConfig(ctx, masterMachine, privateKey)
	if err != nil {
		return nil, err
	}

	k8sConfig := pulumi.Output(kubeConfig).ApplyT(func(result interface{}) (string, error) {
		if resultStr, ok := result.(string); ok {
			return resultStr, nil
		}
		return "", fmt.Errorf("unexpected type for kubeconfig result")
	}).(pulumi.StringOutput)

	providerName, err := utils.CreateUniqueString("provider")
	if err != nil {
		return nil, err
	}

	provider, err := kubernetes.NewProvider(ctx, providerName, &kubernetes.ProviderArgs{
		Kubeconfig: k8sConfig,
	})

	if err != nil {
		return nil, err
	}
	ctx.Export("Provider", provider)
	ctx.Export("Kubeconfig", k8sConfig)

	return provider, err

}
