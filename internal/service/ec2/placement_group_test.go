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

func init() {
	resource.AddTestSweepers("aws_placement_group", &resource.Sweeper{
		Name: "aws_placement_group",
		F:    sweepPlacementGroups,
		Dependencies: []string{
			"aws_autoscaling_group",
			"aws_instance",
			"aws_launch_template",
			"aws_spot_fleet_request",
		},
	})
}

func sweepPlacementGroups(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).EC2Conn
	sweepResources := make([]*sweep.SweepResource, 0)
	var errs *multierror.Error

	input := &ec2.DescribePlacementGroupsInput{}

	// EC2 API provides no NextToken/Marker
	output, err := conn.DescribePlacementGroups(input)

	if err != nil {
		err := fmt.Errorf("error reading EC2 Placement Group: %w", err)
		log.Printf("[ERROR] %s", err)
		errs = multierror.Append(errs, err)
		return errs.ErrorOrNil()
	}

	for _, placementGroup := range output.PlacementGroups {
		r := tfec2.ResourcePlacementGroup()
		d := r.Data(nil)

		d.SetId(aws.StringValue(placementGroup.GroupName))

		sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
	}

	if err = sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping EC2 Placement Group for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping EC2 Placement Group sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

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
