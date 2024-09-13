package k8s

import (
	"github.com/pulumi/pulumi-command/sdk/go/command/remote"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func waitForLockFile(ctx *pulumi.Context, privateKey string, masterIp string) (*remote.Command, error) {
	lockCheckCmd := `
        for i in {1..10}; do
            if [ -f /tmp/microk8s-ready.lock ]; then
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

	ready, err := remote.NewCommand(ctx, "lockCheckCmd", lockCmdArgs)
	if err != nil {
		return nil, err
	}

	return ready, nil
}
