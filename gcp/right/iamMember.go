package right

import (
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/serviceaccount"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func GenerateIamMember(ctx *pulumi.Context, serviceAccount *serviceaccount.Account) (*serviceaccount.IAMMember, error) {
	iamMember, err := serviceaccount.NewIAMMember(ctx, "my-iam-member", &serviceaccount.IAMMemberArgs{
		Role:             pulumi.String("roles/editor"),
		Member:           pulumi.Sprintf("serviceAccount:%s", serviceAccount.Email),
		ServiceAccountId: serviceAccount.ID(),
	})
	if err != nil {
		return nil, err
	}

	return iamMember, nil

}
