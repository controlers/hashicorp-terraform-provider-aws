package aws

import (
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsEc2TransitGateway() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsEc2TransitGatewayRead,

		Schema: map[string]*schema.Schema{
			"amazon_side_asn": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"association_default_route_table_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auto_accept_shared_attachments": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"default_route_table_association": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"default_route_table_propagation": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dns_support": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"filter": dataSourceFiltersSchema(),
			"id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"propagation_default_route_table_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchemaComputed(),
			"vpn_ecmp_support": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsEc2TransitGatewayRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &ec2.DescribeTransitGatewaysInput{}

	if v, ok := d.GetOk("filter"); ok {
		input.Filters = buildAwsDataSourceFilters(v.(*schema.Set))
	}

	if v, ok := d.GetOk("id"); ok {
		input.TransitGatewayIds = []*string{aws.String(v.(string))}
	}

	log.Printf("[DEBUG] Reading EC2 Transit Gateways: %s", input)
	output, err := conn.DescribeTransitGateways(input)

	if err != nil {
		return fmt.Errorf("error reading EC2 Transit Gateway: %w", err)
	}

	if output == nil || len(output.TransitGateways) == 0 {
		return errors.New("error reading EC2 Transit Gateway: no results found")
	}

	if len(output.TransitGateways) > 1 {
		return errors.New("error reading EC2 Transit Gateway: multiple results found, try adjusting search criteria")
	}

	transitGateway := output.TransitGateways[0]

	if transitGateway == nil {
		return errors.New("error reading EC2 Transit Gateway: empty result")
	}

	if transitGateway.Options == nil {
		return errors.New("error reading EC2 Transit Gateway: missing options")
	}

	d.Set("amazon_side_asn", transitGateway.Options.AmazonSideAsn)
	d.Set("arn", transitGateway.TransitGatewayArn)
	d.Set("association_default_route_table_id", transitGateway.Options.AssociationDefaultRouteTableId)
	d.Set("auto_accept_shared_attachments", transitGateway.Options.AutoAcceptSharedAttachments)
	d.Set("default_route_table_association", transitGateway.Options.DefaultRouteTableAssociation)
	d.Set("default_route_table_propagation", transitGateway.Options.DefaultRouteTablePropagation)
	d.Set("description", transitGateway.Description)
	d.Set("dns_support", transitGateway.Options.DnsSupport)
	d.Set("owner_id", transitGateway.OwnerId)
	d.Set("propagation_default_route_table_id", transitGateway.Options.PropagationDefaultRouteTableId)

	if err := d.Set("tags", keyvaluetags.Ec2KeyValueTags(transitGateway.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	d.Set("vpn_ecmp_support", transitGateway.Options.VpnEcmpSupport)

	d.SetId(aws.StringValue(transitGateway.TransitGatewayId))

	return nil
}
