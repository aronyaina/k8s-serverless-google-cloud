package repository

import (
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func GenerateNamespace(ctx *pulumi.Context, name string, environment string, provider *kubernetes.Provider, ressourceDependence []pulumi.Resource) (*corev1.Namespace, error) {
	nameSpaceMetaData := &corev1.NamespaceArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String(name),
			Labels: pulumi.StringMap{
				"environment": pulumi.String(environment),
			},
		},
	}
	myNamespace, err := corev1.NewNamespace(ctx, name, nameSpaceMetaData, pulumi.Provider(provider), pulumi.DependsOn(ressourceDependence))
	if err != nil {
		return nil, err
	}
	ctx.Export("Namespace", myNamespace.Metadata.Name())
	return myNamespace, nil

}
