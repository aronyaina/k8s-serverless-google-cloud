package k8s

import (
	"fmt"

	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func GenerateMasterKubeConfig(ctx *pulumi.Context, masterMachine *compute.Instance) {
	masterIP := masterMachine.NetworkInterfaces.Index(pulumi.Int(0)).ApplyT(func(nics compute.InstanceNetworkInterface) string {
		return *nics.AccessConfigs[0].NatIp
	}).(pulumi.StringOutput)

	masterKubeConfig := masterIP.ApplyT(func(ip string) (string, error) {
		kubeconfig := fmt.Sprintf(`
apiVersion: v1
clusters:
- cluster:
    certificate-authority: /var/snap/microk8s/current/certs/ca.crt
    server: https://%s:16443
  name: microk8s-cluster
contexts:
- context:
    cluster: microk8s-cluster
    user: admin
  name: microk8s-context
current-context: microk8s-context
kind: Config
preferences: {}
users:
- name: admin
  user:
    token: $(sudo microk8s kubectl config view --raw --minify --flatten | grep token)
`, ip)
		return kubeconfig, nil
	}).(pulumi.StringOutput)

	ctx.Export("masterKubeConfig", masterKubeConfig)
}
