package civo

import (
	"fmt"
	"github.com/civo/civogo"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"log"
)

// Firewall Rule resource represent you can create and manage all firewall rules
// this resource don't have an update option because the backend don't have the
// support for that, so in this case we use ForceNew for all object in the resource
func resourceFirewallRule() *schema.Resource {
	fmt.Print()
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"firewall_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateName,
			},
			"protocol": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The protocol choice from tcp, udp or icmp (the default if unspecified is tcp)",
				ValidateFunc: validation.StringInSlice([]string{
					"tcp",
					"udp",
					"icmp",
				}, false),
			},
			"start_port": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Description:  "The start of the port range to configure for this rule (or the single port if required)",
				ValidateFunc: validation.NoZeroValues,
			},
			"end_port": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Description:  "The end of the port range (this is optional, by default it will only apply to the single port listed in start_port)",
				ValidateFunc: validation.NoZeroValues,
			},
			"cird": {
				Type:        schema.TypeSet,
				Required:    true,
				ForceNew:    true,
				Description: "The IP address of the other end (i.e. not your instance) to affect, or a valid network CIDR (defaults to being globally applied, i.e. 0.0.0.0/0)",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"direction": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Will this rule affect inbound or outbound traffic (by default this is inbound)",
				ValidateFunc: validation.StringInSlice([]string{
					"inbound",
					"outbound",
				}, false),
			},
			"label": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Description:  "A string that will be the displayed name/reference for this rule (optional)",
				ValidateFunc: validation.StringIsNotEmpty,
			},
		},
		Create: resourceFirewallRuleCreate,
		Read:   resourceFirewallRuleRead,
		Delete: resourceFirewallRuleDelete,
		Importer: &schema.ResourceImporter{
			State: resourceFirewallRuleImport,
		},
	}
}

// function to create a new firewall rule
func resourceFirewallRuleCreate(d *schema.ResourceData, m interface{}) error {
	apiClient := m.(*civogo.Client)

	tfCidr := d.Get("cird").(*schema.Set).List()
	cird := make([]string, len(tfCidr))
	for i, tfCird := range tfCidr {
		cird[i] = tfCird.(string)
	}

	log.Printf("[INFO] configuring a new firewall rule for firewall %s", d.Get("firewall_id").(string))
	config := &civogo.FirewallRuleConfig{
		FirewallID: d.Get("firewall_id").(string),
		Protocol:   d.Get("protocol").(string),
		StartPort:  d.Get("start_port").(string),
		Direction:  d.Get("direction").(string),
		Cidr:       cird,
	}

	if attr, ok := d.GetOk("end_port"); ok {
		config.EndPort = attr.(string)
	}

	if attr, ok := d.GetOk("label"); ok {
		config.Label = attr.(string)
	}

	log.Printf("[INFO] creating a new firewall rule for firewall %s", d.Get("firewall_id").(string))
	firewallRule, err := apiClient.NewFirewallRule(config)
	if err != nil {
		return fmt.Errorf("[ERR] failed to create a new firewall: %s", err)
	}

	d.SetId(firewallRule.ID)

	return resourceFirewallRuleRead(d, m)
}

// function to read a firewall rule
func resourceFirewallRuleRead(d *schema.ResourceData, m interface{}) error {
	apiClient := m.(*civogo.Client)

	log.Printf("[INFO] retriving the firewall rule %s", d.Id())
	resp, err := apiClient.FindFirewallRule(d.Get("firewall_id").(string), d.Id())
	if err != nil {
		if resp != nil {
			d.SetId("")
			return nil
		}

		return fmt.Errorf("[ERR] error retrieving firewall Rule: %s", err)
	}

	d.Set("firewall_id", resp.FirewallID)
	d.Set("protocol", resp.Protocol)
	d.Set("start_port", resp.StartPort)
	d.Set("end_port", resp.EndPort)
	d.Set("cird", resp.Cidr)
	d.Set("direction", resp.Direction)
	d.Set("label", resp.Label)

	return nil
}

// function to delete a firewall rule
func resourceFirewallRuleDelete(d *schema.ResourceData, m interface{}) error {
	apiClient := m.(*civogo.Client)

	log.Printf("[INFO] retriving the firewall rule %s", d.Id())
	_, err := apiClient.DeleteFirewallRule(d.Get("firewall_id").(string), d.Id())
	if err != nil {
		return fmt.Errorf("[ERR] an error occurred while tring to delete firewall rule %s", d.Id())
	}
	return nil
}

// custom import to able to add a firewall rule to the terraform
func resourceFirewallRuleImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	apiClient := m.(*civogo.Client)

	firewallId, firewallRuleId, err := resourceCommonParseId(d.Id())
	if err != nil {
		return nil, err
	}

	log.Printf("[INFO] retriving the firewall rule %s", firewallRuleId)
	resp, err := apiClient.FindFirewallRule(firewallId, firewallRuleId)
	if err != nil {
		if resp != nil {
			return nil, err
		}
	}

	d.SetId(resp.ID)
	d.Set("firewall_id", resp.FirewallID)
	d.Set("protocol", resp.Protocol)
	d.Set("start_port", resp.StartPort)
	d.Set("end_port", resp.EndPort)
	d.Set("cird", resp.Cidr)
	d.Set("direction", resp.Direction)
	d.Set("label", resp.Label)

	return []*schema.ResourceData{d}, nil
}
