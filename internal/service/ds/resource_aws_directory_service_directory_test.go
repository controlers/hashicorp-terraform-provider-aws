package aws

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directoryservice"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/go-multierror"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/directoryservice/lister"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	tfds "github.com/hashicorp/terraform-provider-aws/internal/service/ds"
	tfds "github.com/hashicorp/terraform-provider-aws/internal/service/ds"
)

func init() {
	resource.AddTestSweepers("aws_directory_service_directory", &resource.Sweeper{
		Name: "aws_directory_service_directory",
		F:    testSweepDirectoryServiceDirectories,
		Dependencies: []string{
			"aws_db_instance",
			"aws_ec2_client_vpn_endpoint",
			"aws_fsx_windows_file_system",
			"aws_transfer_server",
			"aws_workspaces_directory",
		},
	})
}

func testSweepDirectoryServiceDirectories(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).DirectoryServiceConn

	var sweeperErrs *multierror.Error

	input := &directoryservice.DescribeDirectoriesInput{}

	err = tfds.DescribeDirectoriesPagesWithContext(context.TODO(), conn, input, func(page *directoryservice.DescribeDirectoriesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, directory := range page.DirectoryDescriptions {
			id := aws.StringValue(directory.DirectoryId)

			r := ResourceDirectory()
			d := r.Data(nil)
			d.SetId(id)

			err := r.Delete(d, client)

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting Directory Service Directory (%s): %w", id, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Directory Service Directory sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil()
	}

	if err != nil {
		sweeperErr := fmt.Errorf("error listing Directory Service Directories: %w", err)
		log.Printf("[ERROR] %s", sweeperErr)
		sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSDirectoryServiceDirectory_basic(t *testing.T) {
	var ds directoryservice.DirectoryDescription
	resourceName := "aws_directory_service_directory.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckDirectoryService(t)
			acctest.PreCheckDirectoryServiceSimpleDirectory(t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, directoryservice.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDirectoryServiceDirectoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryServiceDirectoryConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceDirectoryExists(resourceName, &ds),
					resource.TestCheckResourceAttrSet(resourceName, "security_group_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"password",
				},
			},
		},
	})
}

func TestAccAWSDirectoryServiceDirectory_tags(t *testing.T) {
	var ds directoryservice.DirectoryDescription
	resourceName := "aws_directory_service_directory.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckDirectoryService(t)
			acctest.PreCheckDirectoryServiceSimpleDirectory(t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, directoryservice.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDirectoryServiceDirectoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryServiceDirectoryTagsConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceDirectoryExists(resourceName, &ds),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.foo", "test"),
					resource.TestCheckResourceAttr(resourceName, "tags.project", "test"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"password",
				},
			},
			{
				Config: testAccDirectoryServiceDirectoryUpdateTagsConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceDirectoryExists(resourceName, &ds),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.foo", "test"),
					resource.TestCheckResourceAttr(resourceName, "tags.project", "test2"),
					resource.TestCheckResourceAttr(resourceName, "tags.fizz", "buzz"),
				),
			},
			{
				Config: testAccDirectoryServiceDirectoryRemoveTagsConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceDirectoryExists(resourceName, &ds),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.foo", "test"),
				),
			},
		},
	})
}

func TestAccAWSDirectoryServiceDirectory_microsoft(t *testing.T) {
	var ds directoryservice.DirectoryDescription
	resourceName := "aws_directory_service_directory.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckDirectoryService(t) },
		ErrorCheck:   acctest.ErrorCheck(t, directoryservice.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDirectoryServiceDirectoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryServiceDirectoryConfig_microsoft,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceDirectoryExists(resourceName, &ds),
					resource.TestCheckResourceAttr(resourceName, "edition", directoryservice.DirectoryEditionEnterprise),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"password",
				},
			},
		},
	})
}

func TestAccAWSDirectoryServiceDirectory_microsoftStandard(t *testing.T) {
	var ds directoryservice.DirectoryDescription
	resourceName := "aws_directory_service_directory.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckDirectoryService(t) },
		ErrorCheck:   acctest.ErrorCheck(t, directoryservice.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDirectoryServiceDirectoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryServiceDirectoryConfig_microsoftStandard,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceDirectoryExists(resourceName, &ds),
					resource.TestCheckResourceAttr(resourceName, "edition", directoryservice.DirectoryEditionStandard),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"password",
				},
			},
		},
	})
}

func TestAccAWSDirectoryServiceDirectory_connector(t *testing.T) {
	var ds directoryservice.DirectoryDescription
	resourceName := "aws_directory_service_directory.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckDirectoryService(t)
			acctest.PreCheckDirectoryServiceSimpleDirectory(t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, directoryservice.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDirectoryServiceDirectoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryServiceDirectoryConfig_connector,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceDirectoryExists(resourceName, &ds),
					resource.TestCheckResourceAttrSet(resourceName, "security_group_id"),
					resource.TestCheckResourceAttrSet(resourceName, "connect_settings.0.connect_ips.#"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"password",
				},
			},
		},
	})
}

func TestAccAWSDirectoryServiceDirectory_withAliasAndSso(t *testing.T) {
	var ds directoryservice.DirectoryDescription
	alias := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_directory_service_directory.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckDirectoryService(t)
			acctest.PreCheckDirectoryServiceSimpleDirectory(t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, directoryservice.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDirectoryServiceDirectoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryServiceDirectoryConfig_withAlias(alias),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceDirectoryExists(resourceName, &ds),
					testAccCheckServiceDirectoryAlias(resourceName, alias),
					testAccCheckServiceDirectorySso(resourceName, false),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"password",
				},
			},
			{
				Config: testAccDirectoryServiceDirectoryConfig_withSso(alias),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceDirectoryExists(resourceName, &ds),
					testAccCheckServiceDirectoryAlias(resourceName, alias),
					testAccCheckServiceDirectorySso(resourceName, true),
				),
			},
			{
				Config: testAccDirectoryServiceDirectoryConfig_withSso_modified(alias),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceDirectoryExists(resourceName, &ds),
					testAccCheckServiceDirectoryAlias(resourceName, alias),
					testAccCheckServiceDirectorySso(resourceName, false),
				),
			},
		},
	})
}

func testAccCheckDirectoryServiceDirectoryDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DirectoryServiceConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_directory_service_directory" {
			continue
		}

		input := directoryservice.DescribeDirectoriesInput{
			DirectoryIds: []*string{aws.String(rs.Primary.ID)},
		}
		out, err := conn.DescribeDirectories(&input)

		if tfawserr.ErrMessageContains(err, directoryservice.ErrCodeEntityDoesNotExistException, "") {
			continue
		}

		if err != nil {
			return err
		}

		if out != nil && len(out.DirectoryDescriptions) > 0 {
			return fmt.Errorf("Expected AWS Directory Service Directory to be gone, but was still found")
		}
	}

	return nil
}

func TestAccAWSDirectoryServiceDirectory_disappears(t *testing.T) {
	var ds directoryservice.DirectoryDescription
	resourceName := "aws_directory_service_directory.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckDirectoryService(t)
			acctest.PreCheckDirectoryServiceSimpleDirectory(t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, directoryservice.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDirectoryServiceDirectoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryServiceDirectoryConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceDirectoryExists(resourceName, &ds),
					acctest.CheckResourceDisappears(acctest.Provider, ResourceDirectory(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckServiceDirectoryExists(name string, ds *directoryservice.DirectoryDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DirectoryServiceConn
		out, err := conn.DescribeDirectories(&directoryservice.DescribeDirectoriesInput{
			DirectoryIds: []*string{aws.String(rs.Primary.ID)},
		})

		if err != nil {
			return err
		}

		if len(out.DirectoryDescriptions) < 1 {
			return fmt.Errorf("No DS directory found")
		}

		if *out.DirectoryDescriptions[0].DirectoryId != rs.Primary.ID {
			return fmt.Errorf("DS directory ID mismatch - existing: %q, state: %q",
				*out.DirectoryDescriptions[0].DirectoryId, rs.Primary.ID)
		}

		*ds = *out.DirectoryDescriptions[0]

		return nil
	}
}

func testAccCheckServiceDirectoryAlias(name, alias string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DirectoryServiceConn
		out, err := conn.DescribeDirectories(&directoryservice.DescribeDirectoriesInput{
			DirectoryIds: []*string{aws.String(rs.Primary.ID)},
		})

		if err != nil {
			return err
		}

		if *out.DirectoryDescriptions[0].Alias != alias {
			return fmt.Errorf("DS directory Alias mismatch - actual: %q, expected: %q",
				*out.DirectoryDescriptions[0].Alias, alias)
		}

		return nil
	}
}

func testAccCheckServiceDirectorySso(name string, ssoEnabled bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DirectoryServiceConn
		out, err := conn.DescribeDirectories(&directoryservice.DescribeDirectoriesInput{
			DirectoryIds: []*string{aws.String(rs.Primary.ID)},
		})

		if err != nil {
			return err
		}

		if *out.DirectoryDescriptions[0].SsoEnabled != ssoEnabled {
			return fmt.Errorf("DS directory SSO mismatch - actual: %t, expected: %t",
				*out.DirectoryDescriptions[0].SsoEnabled, ssoEnabled)
		}

		return nil
	}
}

const testAccDirectoryServiceDirectoryConfigBase = `
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
  tags = {
    Name = "terraform-testacc-directory-service-directory-tags"
  }
}

resource "aws_subnet" "test1" {
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.1.0/24"
  tags = {
    Name = "tf-acc-directory-service-directory-foo"
  }
}

resource "aws_subnet" "test2" {
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[1]
  cidr_block        = "10.0.2.0/24"
  tags = {
    Name = "tf-acc-directory-service-directory-test"
  }
}
`

const testAccDirectoryServiceDirectoryConfig = testAccDirectoryServiceDirectoryConfigBase + `
resource "aws_directory_service_directory" "test" {
  name     = "corp.notexample.com"
  password = "SuperSecretPassw0rd"
  size     = "Small"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = [aws_subnet.test1.id, aws_subnet.test2.id]
  }
}
`

const testAccDirectoryServiceDirectoryTagsConfig = testAccDirectoryServiceDirectoryConfigBase + `
resource "aws_directory_service_directory" "test" {
  name     = "corp.notexample.com"
  password = "SuperSecretPassw0rd"
  size     = "Small"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = [aws_subnet.test1.id, aws_subnet.test2.id]
  }

  tags = {
    foo     = "test"
    project = "test"
  }
}
`

const testAccDirectoryServiceDirectoryUpdateTagsConfig = testAccDirectoryServiceDirectoryConfigBase + `
resource "aws_directory_service_directory" "test" {
  name     = "corp.notexample.com"
  password = "SuperSecretPassw0rd"
  size     = "Small"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = [aws_subnet.test1.id, aws_subnet.test2.id]
  }

  tags = {
    foo     = "test"
    project = "test2"
    fizz    = "buzz"
  }
}
`

const testAccDirectoryServiceDirectoryRemoveTagsConfig = testAccDirectoryServiceDirectoryConfigBase + `
resource "aws_directory_service_directory" "test" {
  name     = "corp.notexample.com"
  password = "SuperSecretPassw0rd"
  size     = "Small"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = [aws_subnet.test1.id, aws_subnet.test2.id]
  }

  tags = {
    foo = "test"
  }
}
`

const testAccDirectoryServiceDirectoryConfig_connector = testAccDirectoryServiceDirectoryConfigBase + `
resource "aws_directory_service_directory" "base" {
  name     = "corp.notexample.com"
  password = "SuperSecretPassw0rd"
  size     = "Small"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = [aws_subnet.test1.id, aws_subnet.test2.id]
  }
}

resource "aws_directory_service_directory" "test" {
  name     = "corp.notexample.com"
  password = "SuperSecretPassw0rd"
  size     = "Small"
  type     = "ADConnector"

  connect_settings {
    customer_dns_ips  = aws_directory_service_directory.base.dns_ip_addresses
    customer_username = "Administrator"
    vpc_id            = aws_vpc.test.id
    subnet_ids        = [aws_subnet.test1.id, aws_subnet.test2.id]
  }
}
`

const testAccDirectoryServiceDirectoryConfig_microsoft = testAccDirectoryServiceDirectoryConfigBase + `
resource "aws_directory_service_directory" "test" {
  name     = "corp.notexample.com"
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = [aws_subnet.test1.id, aws_subnet.test2.id]
  }
}
`

const testAccDirectoryServiceDirectoryConfig_microsoftStandard = testAccDirectoryServiceDirectoryConfigBase + `
resource "aws_directory_service_directory" "test" {
  name     = "corp.notexample.com"
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"
  edition  = "Standard"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = [aws_subnet.test1.id, aws_subnet.test2.id]
  }
}
`

func testAccDirectoryServiceDirectoryConfig_withAlias(alias string) string {
	return testAccDirectoryServiceDirectoryConfigBase + fmt.Sprintf(`
resource "aws_directory_service_directory" "test" {
  name     = "corp.notexample.com"
  password = "SuperSecretPassw0rd"
  size     = "Small"
  alias    = %[1]q

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = [aws_subnet.test1.id, aws_subnet.test2.id]
  }
}
`, alias)
}

func testAccDirectoryServiceDirectoryConfig_withSso(alias string) string {
	return testAccDirectoryServiceDirectoryConfigBase + fmt.Sprintf(`
resource "aws_directory_service_directory" "test" {
  name       = "corp.notexample.com"
  password   = "SuperSecretPassw0rd"
  size       = "Small"
  alias      = %[1]q
  enable_sso = true

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = [aws_subnet.test1.id, aws_subnet.test2.id]
  }
}
`, alias)
}

func testAccDirectoryServiceDirectoryConfig_withSso_modified(alias string) string {
	return testAccDirectoryServiceDirectoryConfigBase + fmt.Sprintf(`
resource "aws_directory_service_directory" "test" {
  name       = "corp.notexample.com"
  password   = "SuperSecretPassw0rd"
  size       = "Small"
  alias      = %[1]q
  enable_sso = false

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = [aws_subnet.test1.id, aws_subnet.test2.id]
  }
}
`, alias)
}