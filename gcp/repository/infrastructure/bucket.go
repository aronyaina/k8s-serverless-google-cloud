package infrastructure

import (
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/storage"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func GenerateBucket(ctx *pulumi.Context, name string) (bucket *storage.Bucket, error error) {
	bucket, err := storage.NewBucket(ctx, name, &storage.BucketArgs{
		Name:         pulumi.String(name),
		Location:     pulumi.String("EU"),
		ForceDestroy: pulumi.Bool(true),
	})

	if err != nil {
		return nil, err
	}
	return bucket, nil
}
