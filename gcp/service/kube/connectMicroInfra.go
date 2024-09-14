package kube

import (
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/compute"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/storage"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func ConnectMicroInfra(ctx *pulumi.Context, masterMachine *compute.Instance, workerMachine []*compute.Instance, workerNumber int, bucket *storage.Bucket, masterPrivateKey string, workerPrivateKey string, triggers pulumi.Array) error {
	masterExternalIp := masterMachine.NetworkInterfaces.Index(pulumi.Int(0)).AccessConfigs().Index(pulumi.Int(0)).NatIp()
	debugOutputCreation, err := GenerateTokenFromMasterAndUploadIt(ctx, masterExternalIp, workerNumber, bucket, masterPrivateKey, triggers)
	if err != nil {
		return err
	}

	workerTriggers := pulumi.Array{debugOutputCreation}
	for workerIndex, workerMachine := range workerMachine {
		workerExternalIp := workerMachine.NetworkInterfaces.Index(pulumi.Int(0)).AccessConfigs().Index(pulumi.Int(0)).NatIp()
		RetrieveTokenFromBucket(ctx, workerExternalIp, workerIndex, bucket, workerPrivateKey, workerTriggers)
	}

	ctx.Export("TokenGeneration", debugOutputCreation)
	ctx.Export("TokenRetrieval", debugOutputCreation)
	return nil
}
