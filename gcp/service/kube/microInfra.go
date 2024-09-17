package kube

import (
	gcpinfra "k8s-serverless/gcp/repository/infrastructure"
	"k8s-serverless/gcp/service/initialization"
	"log"
	"strconv"

	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/compute"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/storage"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func GenerateMicroInfra(ctx *pulumi.Context, vmNumber int) (master_machine *compute.Instance, worker_machines []*compute.Instance, master_private_key string, worker_private_key string, bucket *storage.Bucket, err error) {
	masterPrivateKey, masterPublicKey, err := initialization.GenerateSshIfItDoesntExist("master")
	if err != nil {
		log.Println("Error while creating ssh for master")
		return nil, nil, "", "", nil, err
	}
	workerPrivateKey, workerPublicKey, err := initialization.GenerateSshIfItDoesntExist("worker")
	if err != nil {
		log.Println("Error while creating ssh for worker")
		return nil, nil, "", "", nil, err
	}
	serviceAccount, bucket, err := initialization.GenerateServiceLinkedToBucket(ctx)
	if err != nil {
		log.Println("Error while creating bucket linked to service")
		return nil, nil, "", "", nil, err
	}

	network, subnet, err := initialization.GenerateNetworkInfra(ctx)
	if err != nil {
		log.Println("Error while creating network infrastructure")
		return nil, nil, "", "", nil, err
	}

	masterMachine, err := gcpinfra.GenerateMasterMachine(ctx, vmNumber, "master-node-1", network, subnet, bucket, serviceAccount, masterPublicKey)
	if err != nil {
		log.Println("Error while creating master machine")
		return nil, nil, "", "", nil, err
	}

	var allMachines []*compute.Instance
	var last *compute.Instance = masterMachine

	if vmNumber >= 1 {
		for i := 1; i <= vmNumber; i++ {
			machine, err := gcpinfra.GenerateWorkerMachine(ctx, i, last, "worker-node-"+strconv.Itoa(i), network, subnet, bucket, serviceAccount, workerPublicKey)

			if err != nil {
				return nil, nil, "", "", nil, err
			}
			last = machine
			allMachines = append(allMachines, machine)
		}
	}
	for i, machine := range allMachines {
		ctx.Export("instanceName"+strconv.Itoa(i+1), machine.Name)
		ctx.Export("instanceExternalIP"+strconv.Itoa(i+1), machine.NetworkInterfaces.Index(pulumi.Int(0)).AccessConfigs().Index(pulumi.Int(0)).NatIp())
		ctx.Export("instanceInternalIP"+strconv.Itoa(i+1), machine.NetworkInterfaces.Index(pulumi.Int(0)).NetworkIp())
	}

	return masterMachine, allMachines, masterPrivateKey, workerPrivateKey, bucket, nil
}

func ConnectMicroInfra(ctx *pulumi.Context, masterMachine *compute.Instance, workerMachines []*compute.Instance, workerNumber int, bucket *storage.Bucket, masterPrivateKey string, workerPrivateKey string, triggers pulumi.Array) error {

	masterExternalIp := masterMachine.NetworkInterfaces.Index(pulumi.Int(0)).AccessConfigs().Index(pulumi.Int(0)).NatIp()
	debugOutputCreation, err := GenerateTokenFromMasterAndUploadIt(ctx, "master-token-upload", masterExternalIp, workerNumber, bucket, masterPrivateKey, triggers)
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
