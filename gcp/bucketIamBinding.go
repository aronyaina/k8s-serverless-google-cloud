package gcp

import (
	"fmt"

	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/serviceaccount"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/storage"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func GenerateIamBindingOfBucket(ctx *pulumi.Context, bucket *storage.Bucket, serviceAccount *serviceaccount.Account) (*storage.BucketIAMBinding, error) {
	serviceAccountEmail := serviceAccount.Email.ApplyT(func(email string) string {
		return fmt.Sprintf("serviceAccount:%s", email)
	}).(pulumi.StringOutput)
	aimBinding, err := storage.NewBucketIAMBinding(ctx, "bucketIAMBinding", &storage.BucketIAMBindingArgs{
		Bucket: bucket.Name,
		Role:   pulumi.String("roles/storage.admin"),
		Members: pulumi.StringArray{
			serviceAccountEmail,
		},
	}, pulumi.DependsOn([]pulumi.Resource{bucket, serviceAccount}))
	if err != nil {
		return nil, err
	}
	return aimBinding, nil
}
