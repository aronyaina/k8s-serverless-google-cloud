package repository

import (
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func GenerateService(ctx *pulumi.Context, namespace string, name string, cluster_port int, target_port int, node_port int, appLabels pulumi.StringMap, provider *kubernetes.Provider, ressourceDependence []pulumi.Resource) (*corev1.Service, error) {
	service, err := corev1.NewService(ctx, name, &corev1.ServiceArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Labels:    appLabels,
			Namespace: pulumi.String(namespace),
		},
		Spec: &corev1.ServiceSpecArgs{
			Type:     pulumi.String("NodePort"),
			Selector: appLabels,
			Ports: corev1.ServicePortArray{
				&corev1.ServicePortArgs{
					Port:       pulumi.Int(cluster_port),
					TargetPort: pulumi.Int(target_port),
					NodePort:   pulumi.Int(node_port),
				},
			},
		},
	}, pulumi.Provider(provider), pulumi.DependsOn(ressourceDependence))
	if err != nil {
		return nil, err
	}
	ctx.Export("serviceName", service.Metadata.Name())
	return service, nil
}
