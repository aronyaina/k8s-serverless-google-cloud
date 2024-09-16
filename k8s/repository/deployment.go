package repository

import (
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	appsv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apps/v1"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func GenerateDeployment(ctx *pulumi.Context, appLabels pulumi.StringMap, appName string, provider *kubernetes.Provider, ressourceDependence []pulumi.Resource) (*appsv1.Deployment, error) {
	deployment, err := appsv1.NewDeployment(ctx, appName, &appsv1.DeploymentArgs{
		Metadata: metav1.ObjectMetaArgs{
			Labels:    appLabels,
			Namespace: pulumi.String("dev"),
		},
		Spec: &appsv1.DeploymentSpecArgs{
			Selector: &metav1.LabelSelectorArgs{
				MatchLabels: appLabels,
			},
			Replicas: pulumi.Int(5),
			Template: &corev1.PodTemplateSpecArgs{
				Metadata: &metav1.ObjectMetaArgs{
					Labels: appLabels,
				},
				Spec: &corev1.PodSpecArgs{
					Containers: corev1.ContainerArray{
						&corev1.ContainerArgs{
							Name:  pulumi.String(appName),
							Image: pulumi.String("nginx"),
						},
					},
				},
			},
		},
	}, pulumi.Provider(provider), pulumi.DependsOn(ressourceDependence))
	if err != nil {
		return nil, err
	}
	ctx.Export("deploymentName", deployment.Metadata.Name())
	return deployment, nil
}
