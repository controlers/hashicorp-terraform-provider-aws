package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
)

func TestAccAWSSNSTopicPolicy_basic(t *testing.T) {
	attributes := make(map[string]string)
	resourceName := "aws_sns_topic_policy.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, sns.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSSNSTopicPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSNSTopicPolicyBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSNSTopicExists("aws_sns_topic.test", attributes),
					resource.TestCheckResourceAttrPair(resourceName, "arn", "aws_sns_topic.test", "arn"),
					resource.TestMatchResourceAttr(resourceName, "policy",
						regexp.MustCompile(fmt.Sprintf("\"Sid\":\"%[1]s\"", rName))),
					acctest.CheckResourceAttrAccountID(resourceName, "owner"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSSNSTopicPolicy_updated(t *testing.T) {
	attributes := make(map[string]string)
	resourceName := "aws_sns_topic_policy.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, sns.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSSNSTopicPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSNSTopicPolicyBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSNSTopicExists("aws_sns_topic.test", attributes),
					resource.TestMatchResourceAttr(resourceName, "policy",
						regexp.MustCompile(fmt.Sprintf("\"Sid\":\"%[1]s\"", rName))),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSNSTopicPolicyUpdatedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSNSTopicExists("aws_sns_topic.test", attributes),
					resource.TestMatchResourceAttr(resourceName, "policy",
						regexp.MustCompile(fmt.Sprintf("\"Sid\":\"%[1]s\"", rName))),
					resource.TestMatchResourceAttr(resourceName, "policy",
						regexp.MustCompile("SNS:DeleteTopic")),
				),
			},
		},
	})
}

func TestAccAWSSNSTopicPolicy_disappears_topic(t *testing.T) {
	attributes := make(map[string]string)
	topicResourceName := "aws_sns_topic.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, sns.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSSNSTopicPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSNSTopicPolicyBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSNSTopicExists(topicResourceName, attributes),
					acctest.CheckResourceDisappears(acctest.Provider, ResourceTopic(), topicResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSSNSTopicPolicy_disappears(t *testing.T) {
	attributes := make(map[string]string)
	resourceName := "aws_sns_topic_policy.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, sns.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSSNSTopicPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSNSTopicPolicyBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSNSTopicExists("aws_sns_topic.test", attributes),
					acctest.CheckResourceDisappears(acctest.Provider, ResourceTopicPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSSNSTopicPolicyDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SNSConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sns_topic_policy" {
			continue
		}

		// Check if the topic policy exists by fetching its attributes
		params := &sns.GetTopicAttributesInput{
			TopicArn: aws.String(rs.Primary.ID),
		}
		_, err := conn.GetTopicAttributes(params)
		if err != nil {
			if tfawserr.ErrMessageContains(err, sns.ErrCodeNotFoundException, "") {
				return nil
			}
			return err
		}
		return fmt.Errorf("SNS Topic Policy (%s) exists when it should be destroyed", rs.Primary.ID)
	}

	return nil
}

func testAccAWSSNSTopicPolicyBasicConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_sns_topic_policy" "test" {
  arn    = aws_sns_topic.test.arn
  policy = <<POLICY
{
  "Version":"2012-10-17",
  "Id":"default",
  "Statement":[
    {
      "Sid":"%[1]s",
      "Effect":"Allow",
      "Principal":{
        "AWS":"*"
      },
      "Action":[
        "SNS:GetTopicAttributes",
        "SNS:SetTopicAttributes",
        "SNS:AddPermission",
        "SNS:RemovePermission"
      ],
      "Resource":"${aws_sns_topic.test.arn}"
    }
  ]
}
POLICY
}
`, rName)
}

func testAccAWSSNSTopicPolicyUpdatedConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_sns_topic_policy" "test" {
  arn    = aws_sns_topic.test.arn
  policy = <<POLICY
{
  "Version":"2012-10-17",
  "Id":"default",
  "Statement":[
    {
      "Sid":"%[1]s",
      "Effect":"Allow",
      "Principal":{
        "AWS":"*"
      },
      "Action":[
        "SNS:GetTopicAttributes",
        "SNS:SetTopicAttributes",
        "SNS:AddPermission",
        "SNS:RemovePermission",
        "SNS:DeleteTopic"
      ],
      "Resource":"${aws_sns_topic.test.arn}"
    }
  ]
}
POLICY
}
`, rName)
}
