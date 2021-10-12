package ec2_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	multierror "github.com/hashicorp/go-multierror"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)





func TestAccEC2PlacementGroup_basic(t *testing.T) {
	var pg ec2.PlacementGroup
	resourceName := "aws_placement_group.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckPlacementGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPlacementGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlacementGroupExists(resourceName, &pg),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "strategy", "cluster"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "ec2", fmt.Sprintf("placement-group/%s", rName)),
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

func TestAccEC2PlacementGroup_tags(t *testing.T) {
	var pg ec2.PlacementGroup
	resourceName := "aws_placement_group.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckPlacementGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPlacementGroupTags1Config(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlacementGroupExists(resourceName, &pg),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPlacementGroupTags2Config(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlacementGroupExists(resourceName, &pg),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccPlacementGroupTags1Config(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlacementGroupExists(resourceName, &pg),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2")),
			},
		},
	})
}

func TestAccEC2PlacementGroup_disappears(t *testing.T) {
	var pg ec2.PlacementGroup
	resourceName := "aws_placement_group.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckPlacementGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPlacementGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlacementGroupExists(resourceName, &pg),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourcePlacementGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckPlacementGroupDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_placement_group" {
			continue
		}

		_, err := conn.DescribePlacementGroups(&ec2.DescribePlacementGroupsInput{
			GroupNames: []*string{aws.String(rs.Primary.Attributes["name"])},
		})

		if tfawserr.ErrMessageContains(err, "InvalidPlacementGroup.Unknown", "") {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("still exists")
	}
	return nil
}

func testAccCheckPlacementGroupExists(n string, pg *ec2.PlacementGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Placement Group ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn
		resp, err := conn.DescribePlacementGroups(&ec2.DescribePlacementGroupsInput{
			GroupNames: []*string{aws.String(rs.Primary.ID)},
		})

		if err != nil {
			return fmt.Errorf("Placement Group error: %v", err)
		}

		*pg = *resp.PlacementGroups[0]

		return nil
	}
}

func testAccPlacementGroupConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_placement_group" "test" {
  name     = %q
  strategy = "cluster"
}
`, rName)
}

func testAccPlacementGroupTags1Config(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_placement_group" "test" {
  name     = %[1]q
  strategy = "cluster"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccPlacementGroupTags2Config(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_placement_group" "test" {
  name     = %[1]q
  strategy = "cluster"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
