package service

import (
	"k8s-serverless/gcp/access"
	"k8s-serverless/gcp/infra"
	"log"

	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/serviceaccount"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/storage"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func GenerateServiceLinkedToBucket(ctx *pulumi.Context) (*serviceaccount.Account, *storage.Bucket, error) {
	serviceAccount, err := access.GenerateServiceAccount(ctx)
	if err != nil {
		log.Println("error while creating service account")
		return nil, nil, err
	}

	_, err = access.GenerateIamMember(ctx, serviceAccount)
	if err != nil {
		log.Println("error while creating iam member")
		return nil, nil, err
	}

	bucket, err := infra.GenerateBucket(ctx, "bucket-storage-for-token")
	if err != nil {
		log.Println("error while creating bucket")
		return nil, nil, err
	}

	_, err = access.GenerateIamBindingOfBucket(ctx, bucket, serviceAccount)
	if err != nil {
		log.Println("error while creating iam binding")
		return nil, nil, err
	}

	ctx.Export("serviceAccountName", serviceAccount.Name)
	ctx.Export("BucketName", bucket.Name)
	return serviceAccount, bucket, nil

}
