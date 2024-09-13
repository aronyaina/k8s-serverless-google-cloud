package k8s

import (
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/compute"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func GenerateProvider(ctx *pulumi.Context, machine *compute.Instance, providerName string, microk8sProvider pulumi.String) (*kubernetes.Provider, error) {
	providerName = "provider-" + providerName
	k8sProvider, err := kubernetes.NewProvider(ctx, providerName, &kubernetes.ProviderArgs{
		Kubeconfig: microk8sProvider,
	}, pulumi.DependsOn([]pulumi.Resource{machine}))
	if err != nil {
		return nil, err
	}
	return k8sProvider, err
}
