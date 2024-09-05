package gcp

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/serviceaccount"
)

func GenerateServiceAccount(ctx *pulumi.Context) (*serviceaccount.Account, error) {

	serviceAccount, err := serviceaccount.NewAccount(ctx, "my-service-account", &serviceaccount.AccountArgs{
		AccountId:   pulumi.String("bucket-service-account"),
		DisplayName: pulumi.String("My Bucket Service Account"),
	})
	if err != nil {
		return nil, err
	}
	return serviceAccount, nil
}
