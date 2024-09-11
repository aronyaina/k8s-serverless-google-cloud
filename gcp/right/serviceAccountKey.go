package right

import (
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/serviceaccount"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func GenerateServiceAccountKey(ctx *pulumi.Context, service_account *serviceaccount.Account) (*serviceaccount.Key, error) {
	serviceAccountKey, err := serviceaccount.NewKey(ctx, "my-service-account-key", &serviceaccount.KeyArgs{
		ServiceAccountId: service_account.ID(),
		PrivateKeyType:   pulumi.String("TYPE_GOOGLE_CREDENTIALS_FILE"),
		KeyAlgorithm:     pulumi.String("KEY_ALG_RSA_2048"),
	}, pulumi.Parent(service_account))
	if err != nil {
		return nil, err
	}
	return serviceAccountKey, nil
}
