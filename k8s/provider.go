package k8s

import (
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/compute"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func GenerateProvider(ctx *pulumi.Context, machine *compute.Instance) (*kubernetes.Provider, error) {

	var providerName string = "gcp-k8s-provider"
	var microk8sProvider pulumi.String = "/var/snap/microk8s/current/credentials/kubelet.config"

	k8sProvider, err := kubernetes.NewProvider(ctx, providerName, &kubernetes.ProviderArgs{
		Kubeconfig: microk8sProvider,
	}, pulumi.DependsOn([]pulumi.Resource{machine}))
	if err != nil {
		return nil, err
	}
	return k8sProvider, err
}
