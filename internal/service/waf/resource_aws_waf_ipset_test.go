package aws

import (
	"fmt"
	"log"
	"net"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/hashicorp/go-multierror"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/waf/lister"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func init() {
	resource.AddTestSweepers("aws_waf_ipset", &resource.Sweeper{
		Name: "aws_waf_ipset",
		F:    testSweepWafIPSet,
		Dependencies: []string{
			"aws_waf_rate_based_rule",
			"aws_waf_rule",
			"aws_waf_rule_group",
		},
	})
}

func testSweepWafIPSet(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).WAFConn
	sweepResources := make([]*sweep.SweepResource, 0)
	var errs *multierror.Error
	var g multierror.Group
	var mutex = &sync.Mutex{}

	input := &waf.ListIPSetsInput{}

	err = lister.ListIPSetsPages(conn, input, func(page *waf.ListIPSetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, ipSet := range page.IPSets {
			r := ResourceIPSet()
			d := r.Data(nil)

			id := aws.StringValue(ipSet.IPSetId)
			d.SetId(id)

			// read concurrently and gather errors
			g.Go(func() error {
				// Need to Read first to fill in ip_set_descriptors attribute
				err := r.Read(d, client)

				if err != nil {
					sweeperErr := fmt.Errorf("error reading WAF IP Set (%s): %w", id, err)
					log.Printf("[ERROR] %s", sweeperErr)
					return sweeperErr
				}

				// In case it was already deleted
				if d.Id() == "" {
					return nil
				}

				mutex.Lock()
				defer mutex.Unlock()
				sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))

				return nil
			})
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing WAF IP Set for %s: %w", region, err))
	}

	if err = g.Wait().ErrorOrNil(); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error concurrently reading WAF IP Sets: %w", err))
	}

	if err = sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping WAF IP Set for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping WAF IP Set sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func TestAccAWSWafIPSet_basic(t *testing.T) {
	var v waf.IPSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_waf_ipset.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSWaf(t) },
		ErrorCheck:   acctest.ErrorCheck(t, waf.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSWafIPSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafIPSetConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafIPSetExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ip_set_descriptors.*", map[string]string{
						"type":  "IPV4",
						"value": "192.0.7.0/24",
					}),
					acctest.MatchResourceAttrGlobalARN(resourceName, "arn", "waf", regexp.MustCompile("ipset/.+$")),
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

func TestAccAWSWafIPSet_disappears(t *testing.T) {
	var v waf.IPSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_waf_ipset.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSWaf(t) },
		ErrorCheck:   acctest.ErrorCheck(t, waf.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSWafIPSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafIPSetConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafIPSetExists(resourceName, &v),
					testAccCheckAWSWafIPSetDisappears(&v),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSWafIPSet_changeNameForceNew(t *testing.T) {
	var before, after waf.IPSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	uName := sdkacctest.RandomWithPrefix("tf-acc-test-updated")
	resourceName := "aws_waf_ipset.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSWaf(t) },
		ErrorCheck:   acctest.ErrorCheck(t, waf.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSWafIPSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafIPSetConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafIPSetExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ip_set_descriptors.*", map[string]string{
						"type":  "IPV4",
						"value": "192.0.7.0/24",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSWafIPSetConfigChangeName(uName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafIPSetExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "name", uName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ip_set_descriptors.*", map[string]string{
						"type":  "IPV4",
						"value": "192.0.7.0/24",
					}),
				),
			},
		},
	})
}

func TestAccAWSWafIPSet_changeDescriptors(t *testing.T) {
	var before, after waf.IPSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_waf_ipset.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSWaf(t) },
		ErrorCheck:   acctest.ErrorCheck(t, waf.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSWafIPSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafIPSetConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafIPSetExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "ip_set_descriptors.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ip_set_descriptors.*", map[string]string{
						"type":  "IPV4",
						"value": "192.0.7.0/24",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSWafIPSetConfigChangeIPSetDescriptors(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafIPSetExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "ip_set_descriptors.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ip_set_descriptors.*", map[string]string{
						"type":  "IPV4",
						"value": "192.0.8.0/24",
					}),
				),
			},
		},
	})
}

func TestAccAWSWafIPSet_noDescriptors(t *testing.T) {
	var ipset waf.IPSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_waf_ipset.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSWaf(t) },
		ErrorCheck:   acctest.ErrorCheck(t, waf.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSWafIPSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafIPSetConfig_noDescriptors(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafIPSetExists(resourceName, &ipset),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "ip_set_descriptors.#", "0"),
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

func TestAccAWSWafIPSet_IpSetDescriptors_1000UpdateLimit(t *testing.T) {
	var ipset waf.IPSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_waf_ipset.test"

	incrementIP := func(ip net.IP) {
		for j := len(ip) - 1; j >= 0; j-- {
			ip[j]++
			if ip[j] > 0 {
				break
			}
		}
	}

	// Generate 2048 IPs
	ip, ipnet, err := net.ParseCIDR("10.0.0.0/21")
	if err != nil {
		t.Fatal(err)
	}
	ipSetDescriptors := make([]string, 0, 2048)
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); incrementIP(ip) {
		ipSetDescriptors = append(ipSetDescriptors, fmt.Sprintf("ip_set_descriptors {\ntype=\"IPV4\"\nvalue=\"%s/32\"\n}", ip))
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSWaf(t) },
		ErrorCheck:   acctest.ErrorCheck(t, waf.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSWafIPSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafIPSetConfig_IpSetDescriptors(rName, strings.Join(ipSetDescriptors, "\n")),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafIPSetExists(resourceName, &ipset),
					resource.TestCheckResourceAttr(resourceName, "ip_set_descriptors.#", "2048"),
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

func TestDiffWafIpSetDescriptors(t *testing.T) {
	testCases := []struct {
		Old             []interface{}
		New             []interface{}
		ExpectedUpdates [][]*waf.IPSetUpdate
	}{
		{
			// Change
			Old: []interface{}{
				map[string]interface{}{"type": "IPV4", "value": "192.0.7.0/24"},
			},
			New: []interface{}{
				map[string]interface{}{"type": "IPV4", "value": "192.0.8.0/24"},
			},
			ExpectedUpdates: [][]*waf.IPSetUpdate{
				{
					{
						Action: aws.String(waf.ChangeActionDelete),
						IPSetDescriptor: &waf.IPSetDescriptor{
							Type:  aws.String("IPV4"),
							Value: aws.String("192.0.7.0/24"),
						},
					},
					{
						Action: aws.String(waf.ChangeActionInsert),
						IPSetDescriptor: &waf.IPSetDescriptor{
							Type:  aws.String("IPV4"),
							Value: aws.String("192.0.8.0/24"),
						},
					},
				},
			},
		},
		{
			// Fresh IPSet
			Old: []interface{}{},
			New: []interface{}{
				map[string]interface{}{"type": "IPV4", "value": "10.0.1.0/24"},
				map[string]interface{}{"type": "IPV4", "value": "10.0.2.0/24"},
				map[string]interface{}{"type": "IPV4", "value": "10.0.3.0/24"},
			},
			ExpectedUpdates: [][]*waf.IPSetUpdate{
				{
					{
						Action: aws.String(waf.ChangeActionInsert),
						IPSetDescriptor: &waf.IPSetDescriptor{
							Type:  aws.String("IPV4"),
							Value: aws.String("10.0.1.0/24"),
						},
					},
					{
						Action: aws.String(waf.ChangeActionInsert),
						IPSetDescriptor: &waf.IPSetDescriptor{
							Type:  aws.String("IPV4"),
							Value: aws.String("10.0.2.0/24"),
						},
					},
					{
						Action: aws.String(waf.ChangeActionInsert),
						IPSetDescriptor: &waf.IPSetDescriptor{
							Type:  aws.String("IPV4"),
							Value: aws.String("10.0.3.0/24"),
						},
					},
				},
			},
		},
		{
			// Deletion
			Old: []interface{}{
				map[string]interface{}{"type": "IPV4", "value": "192.0.7.0/24"},
				map[string]interface{}{"type": "IPV4", "value": "192.0.8.0/24"},
			},
			New: []interface{}{},
			ExpectedUpdates: [][]*waf.IPSetUpdate{
				{
					{
						Action: aws.String(waf.ChangeActionDelete),
						IPSetDescriptor: &waf.IPSetDescriptor{
							Type:  aws.String(waf.IPSetDescriptorTypeIpv4),
							Value: aws.String("192.0.7.0/24"),
						},
					},
					{
						Action: aws.String(waf.ChangeActionDelete),
						IPSetDescriptor: &waf.IPSetDescriptor{
							Type:  aws.String(waf.IPSetDescriptorTypeIpv4),
							Value: aws.String("192.0.8.0/24"),
						},
					},
				},
			},
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			updates := diffIPSetDescriptors(tc.Old, tc.New)
			if !reflect.DeepEqual(updates, tc.ExpectedUpdates) {
				t.Fatalf("IPSet updates don't match.\nGiven: %s\nExpected: %s",
					updates, tc.ExpectedUpdates)
			}
		})
	}
}

func TestAccAWSWafIPSet_ipv6(t *testing.T) {
	var v waf.IPSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_waf_ipset.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSWaf(t) },
		ErrorCheck:   acctest.ErrorCheck(t, waf.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSWafIPSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafIPSetIPV6Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafIPSetExists(resourceName, &v),
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

func testAccCheckAWSWafIPSetDisappears(v *waf.IPSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFConn

		wr := newWafRetryer(conn)
		_, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
			req := &waf.UpdateIPSetInput{
				ChangeToken: token,
				IPSetId:     v.IPSetId,
			}

			for _, IPSetDescriptor := range v.IPSetDescriptors {
				IPSetUpdate := &waf.IPSetUpdate{
					Action: aws.String("DELETE"),
					IPSetDescriptor: &waf.IPSetDescriptor{
						Type:  IPSetDescriptor.Type,
						Value: IPSetDescriptor.Value,
					},
				}
				req.Updates = append(req.Updates, IPSetUpdate)
			}

			return conn.UpdateIPSet(req)
		})
		if err != nil {
			return fmt.Errorf("Error Updating WAF IPSet: %s", err)
		}

		_, err = wr.RetryWithToken(func(token *string) (interface{}, error) {
			opts := &waf.DeleteIPSetInput{
				ChangeToken: token,
				IPSetId:     v.IPSetId,
			}
			return conn.DeleteIPSet(opts)
		})
		if err != nil {
			return fmt.Errorf("Error Deleting WAF IPSet: %s", err)
		}
		return nil
	}
}

func testAccCheckAWSWafIPSetDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_waf_ipset" {
			continue
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFConn
		resp, err := conn.GetIPSet(
			&waf.GetIPSetInput{
				IPSetId: aws.String(rs.Primary.ID),
			})

		if err == nil {
			if *resp.IPSet.IPSetId == rs.Primary.ID {
				return fmt.Errorf("WAF IPSet %s still exists", rs.Primary.ID)
			}
		}

		// Return nil if the IPSet is already destroyed
		if tfawserr.ErrMessageContains(err, waf.ErrCodeNonexistentItemException, "") {
			return nil
		}

		return err
	}

	return nil
}

func testAccCheckAWSWafIPSetExists(n string, v *waf.IPSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No WAF IPSet ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFConn
		resp, err := conn.GetIPSet(&waf.GetIPSetInput{
			IPSetId: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if *resp.IPSet.IPSetId == rs.Primary.ID {
			*v = *resp.IPSet
			return nil
		}

		return fmt.Errorf("WAF IPSet (%s) not found", rs.Primary.ID)
	}
}

func testAccAWSWafIPSetConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_ipset" "test" {
  name = %[1]q

  ip_set_descriptors {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }
}
`, name)
}

func testAccAWSWafIPSetConfigChangeName(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_ipset" "test" {
  name = %[1]q

  ip_set_descriptors {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }
}
`, name)
}

func testAccAWSWafIPSetConfigChangeIPSetDescriptors(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_ipset" "test" {
  name = %[1]q

  ip_set_descriptors {
    type  = "IPV4"
    value = "192.0.8.0/24"
  }
}
`, name)
}

func testAccAWSWafIPSetConfig_IpSetDescriptors(name, ipSetDescriptors string) string {
	return fmt.Sprintf(`
resource "aws_waf_ipset" "test" {
  name = %[1]q

  %s
}
`, name, ipSetDescriptors)
}

func testAccAWSWafIPSetConfig_noDescriptors(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_ipset" "test" {
  name = %[1]q
}
`, name)
}

func testAccAWSWafIPSetIPV6Config(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_ipset" "test" {
  name = %[1]q

  ip_set_descriptors {
    type  = "IPV6"
    value = "1234:5678:9abc:6811:0000:0000:0000:0000/64"
  }
}
`, name)
}