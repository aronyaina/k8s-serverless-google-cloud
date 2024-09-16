package initialization

import (
	"k8s-serverless/gcp/repository/access"
	"k8s-serverless/gcp/repository/infrastructure"
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

	bucket, err := infrastructure.GenerateBucket(ctx, "gs-token")
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
