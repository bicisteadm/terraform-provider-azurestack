package azurerm

import (
	"fmt"
	"log"

	"github.com/Azure/azure-sdk-for-go/services/mysql/mgmt/2017-04-30-preview/mysql"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func resourceArmMySqlFirewallRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceArmMySqlFirewallRuleCreateUpdate,
		Read:   resourceArmMySqlFirewallRuleRead,
		Update: resourceArmMySqlFirewallRuleCreateUpdate,
		Delete: resourceArmMySqlFirewallRuleDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"resource_group_name": resourceGroupNameSchema(),

			"server_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"start_ip_address": {
				Type:     schema.TypeString,
				Required: true,
			},

			"end_ip_address": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceArmMySqlFirewallRuleCreateUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).mysqlFirewallRulesClient
	ctx := meta.(*ArmClient).StopContext

	log.Printf("[INFO] preparing arguments for AzureRM MySQL Firewall Rule creation.")

	name := d.Get("name").(string)
	resourceGroup := d.Get("resource_group_name").(string)
	serverName := d.Get("server_name").(string)
	startIPAddress := d.Get("start_ip_address").(string)
	endIPAddress := d.Get("end_ip_address").(string)

	properties := mysql.FirewallRule{
		FirewallRuleProperties: &mysql.FirewallRuleProperties{
			StartIPAddress: utils.String(startIPAddress),
			EndIPAddress:   utils.String(endIPAddress),
		},
	}

	future, err := client.CreateOrUpdate(ctx, resourceGroup, serverName, name, properties)
	if err != nil {
		return err
	}

	err = future.WaitForCompletion(ctx, client.Client)
	if err != nil {
		return err
	}

	read, err := client.Get(ctx, resourceGroup, serverName, name)
	if err != nil {
		return err
	}
	if read.ID == nil {
		return fmt.Errorf("Cannot read MySQL Firewall Rule %q (Gesource Group %q) ID", name, resourceGroup)
	}

	d.SetId(*read.ID)

	return resourceArmMySqlFirewallRuleRead(d, meta)
}

func resourceArmMySqlFirewallRuleRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).mysqlFirewallRulesClient
	ctx := meta.(*ArmClient).StopContext

	id, err := parseAzureResourceID(d.Id())
	if err != nil {
		return err
	}
	resourceGroup := id.ResourceGroup
	serverName := id.Path["servers"]
	name := id.Path["firewallRules"]

	resp, err := client.Get(ctx, resourceGroup, serverName, name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error making Read request on Azure MySQL Firewall Rule %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	d.Set("name", resp.Name)
	d.Set("resource_group_name", resourceGroup)
	d.Set("server_name", serverName)
	d.Set("start_ip_address", resp.StartIPAddress)
	d.Set("end_ip_address", resp.EndIPAddress)

	return nil
}

func resourceArmMySqlFirewallRuleDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).mysqlFirewallRulesClient
	ctx := meta.(*ArmClient).StopContext

	id, err := parseAzureResourceID(d.Id())
	if err != nil {
		return err
	}
	resourceGroup := id.ResourceGroup
	serverName := id.Path["servers"]
	name := id.Path["firewallRules"]

	future, err := client.Delete(ctx, resourceGroup, serverName, name)
	if err != nil {
		return err
	}

	err = future.WaitForCompletion(ctx, client.Client)
	if err != nil {
		return err
	}

	return nil
}