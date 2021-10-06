package autoscalingplans

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscalingplans"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_autoscalingplans_scaling_plan", &resource.Sweeper{
		Name: "aws_autoscalingplans_scaling_plan",
		F:    sweepScalingPlans,
	})
}

func sweepScalingPlans(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).AutoScalingPlansConn
	input := &autoscalingplans.DescribeScalingPlansInput{}
	var sweeperErrs *multierror.Error

	for {
		output, err := conn.DescribeScalingPlans(input)
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Auto Scaling Scaling Plans sweep for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
		}
		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Auto Scaling Scaling Plans: %w", err))
			return sweeperErrs.ErrorOrNil()
		}

		for _, scalingPlan := range output.ScalingPlans {
			scalingPlanName := aws.StringValue(scalingPlan.ScalingPlanName)
			scalingPlanVersion := int(aws.Int64Value(scalingPlan.ScalingPlanVersion))

			r := ResourceScalingPlan()
			d := r.Data(nil)
			d.SetId("????????????????") // ID not used in Delete.
			d.Set("name", scalingPlanName)
			d.Set("scaling_plan_version", scalingPlanVersion)
			err = r.Delete(d, client)

			if err != nil {
				log.Printf("[ERROR] %s", err)
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}
		input.NextToken = output.NextToken
	}

	return sweeperErrs.ErrorOrNil()
}
