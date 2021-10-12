package waiter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/amplify"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/amplify/finder"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	tfamplify "github.com/hashicorp/terraform-provider-aws/internal/service/amplify"
	tfamplify "github.com/hashicorp/terraform-provider-aws/internal/service/amplify"
	tfamplify "github.com/hashicorp/terraform-provider-aws/internal/service/amplify"
	tfamplify "github.com/hashicorp/terraform-provider-aws/internal/service/amplify"
	tfamplify "github.com/hashicorp/terraform-provider-aws/internal/service/amplify"
)

func statusDomainAssociation(conn *amplify.Amplify, appID, domainName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		domainAssociation, err := tfamplify.FindDomainAssociationByAppIDAndDomainName(conn, appID, domainName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return domainAssociation, aws.StringValue(domainAssociation.DomainStatus), nil
	}
}
