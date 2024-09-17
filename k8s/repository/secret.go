package repository

import (
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func GenerateSecret(ctx *pulumi.Context, namespace string, name string, secretType string, data pulumi.StringMap, provider *kubernetes.Provider, ressourceDependence []pulumi.Resource) (*corev1.Secret, error) {
	secret, err := corev1.NewSecret(ctx, name, &corev1.SecretArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String(name),
			Namespace: pulumi.String(namespace),
		},
		Type:       pulumi.String(secretType),
		StringData: data,
	}, pulumi.Provider(provider), pulumi.DependsOn(ressourceDependence))
	if err != nil {
		return nil, err
	}
	return secret, nil
}
