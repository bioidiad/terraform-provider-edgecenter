package edgecenter

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/Edge-Center/edgecentercloud-go/edgecenter/reservedfixedip/v1/reservedfixedips"
)

func dataSourceReservedFixedIP() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceReservedFixedIPRead,
		Description: "Represent reserved ips",
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				Description:  "The uuid of the project. Either 'project_id' or 'project_name' must be specified.",
				ExactlyOneOf: []string{"project_id", "project_name"},
			},
			"project_name": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "The name of the project. Either 'project_id' or 'project_name' must be specified.",
				ExactlyOneOf: []string{"project_id", "project_name"},
			},
			"region_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				Description:  "The uuid of the region. Either 'region_id' or 'region_name' must be specified.",
				ExactlyOneOf: []string{"region_id", "region_name"},
			},
			"region_name": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "The name of the region. Either 'region_id' or 'region_name' must be specified.",
				ExactlyOneOf: []string{"region_id", "region_name"},
			},
			"fixed_ip_address": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The IP address that is associated with the reserved IP.",
				ValidateDiagFunc: func(val interface{}, key cty.Path) diag.Diagnostics {
					v := val.(string)
					ip := net.ParseIP(v)
					if ip != nil {
						return diag.Diagnostics{}
					}

					return diag.FromErr(fmt.Errorf("%q must be a valid ip, got: %s", key, v))
				},
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The current status of the reserved fixed IP.",
			},
			"subnet_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the subnet from which the fixed IP should be reserved.",
			},
			"network_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the network to which the reserved fixed IP is associated.",
			},
			"is_vip": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Flag to determine if the reserved fixed IP should be treated as a Virtual IP (VIP).",
			},
			"port_id": {
				Type:        schema.TypeString,
				Description: "ID of the port_id underlying the reserved fixed IP",
				Computed:    true,
			},
			"allowed_address_pairs": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Group of IP addresses that share the current IP as VIP.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip_address": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"mac_address": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceReservedFixedIPRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Println("[DEBUG] Start ReservedFixedIP reading")
	var diags diag.Diagnostics
	config := m.(*Config)
	provider := config.Provider

	client, err := CreateClient(provider, d, ReservedFixedIPsPoint, VersionPointV1)
	if err != nil {
		return diag.FromErr(err)
	}

	ipAddr := d.Get("fixed_ip_address").(string)
	ips, err := reservedfixedips.ListAll(client, reservedfixedips.ListOpts{})
	if err != nil {
		return diag.FromErr(err)
	}

	var found bool
	var reservedFixedIP reservedfixedips.ReservedFixedIP
	for _, ip := range ips {
		if ip.FixedIPAddress.String() == ipAddr {
			reservedFixedIP = ip
			found = true
			break
		}
	}

	if !found {
		return diag.Errorf("reserved fixed ip %s not found", ipAddr)
	}

	// should we use PortID as id?
	d.SetId(reservedFixedIP.PortID)
	d.Set("project_id", reservedFixedIP.ProjectID)
	d.Set("region_id", reservedFixedIP.RegionID)
	d.Set("status", reservedFixedIP.Status)
	d.Set("fixed_ip_address", reservedFixedIP.FixedIPAddress.String())
	d.Set("subnet_id", reservedFixedIP.SubnetID)
	d.Set("network_id", reservedFixedIP.NetworkID)
	d.Set("is_vip", reservedFixedIP.IsVip)
	d.Set("port_id", reservedFixedIP.PortID)

	allowedPairs := make([]map[string]interface{}, len(reservedFixedIP.AllowedAddressPairs))
	for i, p := range reservedFixedIP.AllowedAddressPairs {
		pair := make(map[string]interface{})

		pair["ip_address"] = p.IPAddress
		pair["mac_address"] = p.MacAddress

		allowedPairs[i] = pair
	}
	if err := d.Set("allowed_address_pairs", allowedPairs); err != nil {
		return diag.FromErr(err)
	}

	log.Println("[DEBUG] Finish ReservedFixedIP reading")

	return diags
}
