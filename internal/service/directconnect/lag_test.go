package directconnect_test

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/go-multierror"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfdirectconnect "github.com/hashicorp/terraform-provider-aws/internal/service/directconnect"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccDirectConnectLag_basic(t *testing.T) {
	var lag directconnect.Lag
	resourceName := "aws_dx_lag.test"
	rName1 := sdkacctest.RandomWithPrefix("tf-acc-test")
	rName2 := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, directconnect.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLagDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxLagConfigBasic(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLagExists(resourceName, &lag),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "directconnect", regexp.MustCompile(`dxlag/.+`)),
					resource.TestCheckResourceAttr(resourceName, "connections_bandwidth", "1Gbps"),
					resource.TestCheckResourceAttr(resourceName, "force_destroy", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "has_logical_redundancy"),
					resource.TestCheckResourceAttrSet(resourceName, "jumbo_frame_capable"),
					resource.TestCheckResourceAttrSet(resourceName, "location"),
					resource.TestCheckResourceAttr(resourceName, "name", rName1),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_account_id"),
					resource.TestCheckResourceAttr(resourceName, "provider_name", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				Config: testAccDxLagConfigBasic(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLagExists(resourceName, &lag),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "directconnect", regexp.MustCompile(`dxlag/.+`)),
					resource.TestCheckResourceAttr(resourceName, "connections_bandwidth", "1Gbps"),
					resource.TestCheckResourceAttr(resourceName, "force_destroy", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "has_logical_redundancy"),
					resource.TestCheckResourceAttrSet(resourceName, "jumbo_frame_capable"),
					resource.TestCheckResourceAttrSet(resourceName, "location"),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_account_id"),
					resource.TestCheckResourceAttr(resourceName, "provider_name", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			// Test import.
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
		},
	})
}

func TestAccDirectConnectLag_providerName(t *testing.T) {
	var lag directconnect.Lag
	resourceName := "aws_dx_lag.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, directconnect.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLagDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxLagConfigProviderName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLagExists(resourceName, &lag),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "directconnect", regexp.MustCompile(`dxlag/.+`)),
					resource.TestCheckResourceAttr(resourceName, "connections_bandwidth", "1Gbps"),
					resource.TestCheckResourceAttr(resourceName, "force_destroy", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "has_logical_redundancy"),
					resource.TestCheckResourceAttrSet(resourceName, "jumbo_frame_capable"),
					resource.TestCheckResourceAttrSet(resourceName, "location"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_account_id"),
					resource.TestCheckResourceAttrSet(resourceName, "provider_name"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			// Test import.
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
		},
	})
}

func TestAccDirectConnectLag_disappears(t *testing.T) {
	var lag directconnect.Lag
	resourceName := "aws_dx_lag.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, directconnect.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLagDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxLagConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLagExists(resourceName, &lag),
					acctest.CheckResourceDisappears(acctest.Provider, tfdirectconnect.ResourceLag(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDirectConnectLag_tags(t *testing.T) {
	var lag directconnect.Lag
	resourceName := "aws_dx_lag.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, directconnect.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLagDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxLagConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLagExists(resourceName, &lag),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			// Test import.
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
			{
				Config: testAccDxLagConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLagExists(resourceName, &lag),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccDxLagConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLagExists(resourceName, &lag),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckLagDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DirectConnectConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_dx_lag" {
			continue
		}

		_, err := tfdirectconnect.FindLagByID(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Direct Connect LAG %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckLagExists(name string, v *directconnect.Lag) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DirectConnectConn

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		lag, err := tfdirectconnect.FindLagByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *lag

		return nil
	}
}

func testAccDxLagConfigBasic(rName string) string {
	return fmt.Sprintf(`
data "aws_dx_locations" "test" {}

resource "aws_dx_lag" "test" {
  name                  = %[1]q
  connections_bandwidth = "1Gbps"
  location              = tolist(data.aws_dx_locations.test.location_codes)[0]
}
`, rName)
}

func testAccDxLagConfigProviderName(rName string) string {
	return fmt.Sprintf(`
data "aws_dx_locations" "test" {}

data "aws_dx_location" "test" {
  location_code = tolist(data.aws_dx_locations.test.location_codes)[0]
}

resource "aws_dx_lag" "test" {
  name                  = %[1]q
  connections_bandwidth = "1Gbps"
  location              = data.aws_dx_location.test.location_code

  provider_name = data.aws_dx_location.test.available_providers[0]
}
`, rName)
}

func testAccDxLagConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
data "aws_dx_locations" "test" {}

resource "aws_dx_lag" "test" {
  name                  = %[1]q
  connections_bandwidth = "1Gbps"
  location              = tolist(data.aws_dx_locations.test.location_codes)[0]
  force_destroy         = true

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccDxLagConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
data "aws_dx_locations" "test" {}

resource "aws_dx_lag" "test" {
  name                  = %[1]q
  connections_bandwidth = "1Gbps"
  location              = tolist(data.aws_dx_locations.test.location_codes)[0]
  force_destroy         = true

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
