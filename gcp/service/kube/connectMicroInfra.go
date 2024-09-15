package kube

import (
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/compute"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/storage"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func ConnectMicroInfra(ctx *pulumi.Context, masterMachine *compute.Instance, workerMachines []*compute.Instance, workerNumber int, bucket *storage.Bucket, masterPrivateKey string, workerPrivateKey string, triggers pulumi.Array) error {
	masterExternalIp := masterMachine.NetworkInterfaces.Index(pulumi.Int(0)).AccessConfigs().Index(pulumi.Int(0)).NatIp()
	debugOutputCreation, err := GenerateTokenFromMasterAndUploadIt(ctx, masterExternalIp, workerNumber, bucket, masterPrivateKey, triggers)
	if err != nil {
		return err
	}

	workerTriggers := pulumi.Array{debugOutputCreation}

	if workerNumber >= 1 {
		for i := 1; i <= workerNumber; i++ {
			workerExternalIp := workerMachines[i-1].NetworkInterfaces.Index(pulumi.Int(0)).AccessConfigs().Index(pulumi.Int(0)).NatIp()
			RetrieveTokenFromBucket(ctx, workerExternalIp, i, bucket, workerPrivateKey, workerTriggers)
		}
	}

	ctx.Export("TokenGeneration", debugOutputCreation)
	ctx.Export("TokenRetrieval", debugOutputCreation)
	return nil
}
