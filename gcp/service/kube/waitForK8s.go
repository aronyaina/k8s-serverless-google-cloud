package kube

import (
	"k8s-serverless/utils"

	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func WaitForLockFile(ctx *pulumi.Context, privateKey string, masterIp string) (*remote.Command, error) {
	lockCheckCmd := `
        for i in {1..10}; do
            if [ -f /tmp/k8sready.lock ]; then
                echo "Lock file exists"
                break
            else
                echo "Waiting for lock file"
                sleep 10
            fi
        done
    `

	lockCmdArgs := &remote.CommandArgs{
		Create: pulumi.String(lockCheckCmd),
		Connection: &remote.ConnectionArgs{
			Host:       pulumi.String(masterIp),
			User:       pulumi.String("pulumi"),
			PrivateKey: pulumi.String(privateKey),
		},
	}

	name, err := utils.CreateUniqueString("lockCheckCmd-master-node")
	if err != nil {
		return nil, err
	}
	ready, err := remote.NewCommand(ctx, name, lockCmdArgs)
	if err != nil {
		return nil, err
	}

	return ready, nil
}

//namespace, err := ressources.GenerateNameSpace(ctx, "dev-ns", "dev", provider)
//if err != nil {
//	return err
//}

//ctx.Export("namespaceName", namespace.Metadata.Name())
