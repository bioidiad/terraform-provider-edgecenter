package edgecenter

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/Edge-Center/edgecentercloud-go/edgecenter/securitygroup/v1/securitygrouprules"
	"github.com/Edge-Center/edgecentercloud-go/edgecenter/securitygroup/v1/securitygroups"
	"github.com/Edge-Center/edgecentercloud-go/edgecenter/securitygroup/v1/types"
)

const (
	SecurityGroupPoint      = "securitygroups"
	securityGroupRulesPoint = "securitygrouprules"
)

func resourceSecurityGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSecurityGroupCreate,
		ReadContext:   resourceSecurityGroupRead,
		UpdateContext: resourceSecurityGroupUpdate,
		DeleteContext: resourceSecurityGroupDelete,
		Description:   "Represent SecurityGroups(Firewall)",
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				projectID, regionID, sgID, err := ImportStringParser(d.Id())
				if err != nil {
					return nil, err
				}
				d.Set("project_id", projectID)
				d.Set("region_id", regionID)
				d.SetId(sgID)

				return []*schema.ResourceData{d}, nil
			},
		},

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
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the security group.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A detailed description of the security group.",
			},
			"metadata_map": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "A map containing metadata, for example tags.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"metadata_read_only": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: `A list of read-only metadata items, e.g. tags.`,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"value": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"read_only": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			"security_group_rules": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: "Firewall rules control what inbound(ingress) and outbound(egress) traffic is allowed to enter or leave a Instance. At least one 'egress' rule should be set",
				Set:         secGroupUniqueID,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"direction": {
							Type:        schema.TypeString,
							Required:    true,
							Description: fmt.Sprintf("Available value is '%s', '%s'", types.RuleDirectionIngress, types.RuleDirectionEgress),
							ValidateDiagFunc: func(v interface{}, path cty.Path) diag.Diagnostics {
								val := v.(string)
								switch types.RuleDirection(val) {
								case types.RuleDirectionIngress, types.RuleDirectionEgress:
									return nil
								}
								return diag.Errorf("wrong direction '%s', available value is '%s', '%s'", val, types.RuleDirectionIngress, types.RuleDirectionEgress)
							},
						},
						"ethertype": {
							Type:        schema.TypeString,
							Required:    true,
							Description: fmt.Sprintf("Available value is '%s', '%s'", types.EtherTypeIPv4, types.EtherTypeIPv6),
							ValidateDiagFunc: func(v interface{}, path cty.Path) diag.Diagnostics {
								val := v.(string)
								switch types.EtherType(val) {
								case types.EtherTypeIPv4, types.EtherTypeIPv6:
									return nil
								}
								return diag.Errorf("wrong ethertype '%s', available value is '%s', '%s'", val, types.EtherTypeIPv4, types.EtherTypeIPv6)
							},
						},
						"protocol": {
							Type:        schema.TypeString,
							Required:    true,
							Description: fmt.Sprintf("Available value is %s", strings.Join(types.Protocol("").StringList(), ",")),
						},
						"port_range_min": {
							Type:             schema.TypeInt,
							Optional:         true,
							Default:          0,
							ValidateDiagFunc: validatePortRange,
						},
						"port_range_max": {
							Type:             schema.TypeInt,
							Optional:         true,
							Default:          0,
							ValidateDiagFunc: validatePortRange,
						},
						"description": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "",
						},
						"remote_ip_prefix": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "",
						},
						"updated_at": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"created_at": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"last_updated": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The timestamp of the last update (use with update context).",
			},
		},
	}
}

func resourceSecurityGroupCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Println("[DEBUG] Start SecurityGroup creating")

	var valid bool
	vals := d.Get("security_group_rules").(*schema.Set).List()
	for _, val := range vals {
		rule := val.(map[string]interface{})
		if types.RuleDirection(rule["direction"].(string)) == types.RuleDirectionEgress {
			valid = true
			break
		}
	}
	if !valid {
		return diag.Errorf("at least one 'egress' rule should be set")
	}

	var diags diag.Diagnostics
	config := m.(*Config)
	provider := config.Provider

	client, err := CreateClient(provider, d, SecurityGroupPoint, VersionPointV1)
	if err != nil {
		return diag.FromErr(err)
	}

	rawRules := d.Get("security_group_rules").(*schema.Set).List()
	rules := make([]securitygroups.CreateSecurityGroupRuleOpts, len(rawRules))
	for i, r := range rawRules {
		rule := r.(map[string]interface{})

		portRangeMax := rule["port_range_max"].(int)
		portRangeMin := rule["port_range_min"].(int)
		descr := rule["description"].(string)
		remoteIPPrefix := rule["remote_ip_prefix"].(string)

		sgrOpts := securitygroups.CreateSecurityGroupRuleOpts{
			Direction:   types.RuleDirection(rule["direction"].(string)),
			EtherType:   types.EtherType(rule["ethertype"].(string)),
			Protocol:    types.Protocol(rule["protocol"].(string)),
			Description: &descr,
		}

		if remoteIPPrefix != "" {
			sgrOpts.RemoteIPPrefix = &remoteIPPrefix
		}

		if portRangeMax != 0 && portRangeMin != 0 {
			sgrOpts.PortRangeMax = &portRangeMax
			sgrOpts.PortRangeMin = &portRangeMin
		}

		rules[i] = sgrOpts
	}

	createSecurityGroupOpts := &securitygroups.CreateSecurityGroupOpts{}
	createSecurityGroupOpts.Name = d.Get("name").(string)
	createSecurityGroupOpts.SecurityGroupRules = rules

	if metadataRaw, ok := d.GetOk("metadata_map"); ok {
		createSecurityGroupOpts.Metadata = metadataRaw.(map[string]interface{})
	}

	opts := securitygroups.CreateOpts{
		SecurityGroup: *createSecurityGroupOpts,
	}
	descr := d.Get("description").(string)
	if descr != "" {
		opts.SecurityGroup.Description = &descr
	}

	sg, err := securitygroups.Create(client, opts).Extract()
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(sg.ID)

	resourceSecurityGroupRead(ctx, d, m)
	log.Printf("[DEBUG] Finish SecurityGroup creating (%s)", sg.ID)

	return diags
}

func resourceSecurityGroupRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Println("[DEBUG] Start SecurityGroup reading")
	var diags diag.Diagnostics
	config := m.(*Config)
	provider := config.Provider

	client, err := CreateClient(provider, d, SecurityGroupPoint, VersionPointV1)
	if err != nil {
		return diag.FromErr(err)
	}

	sg, err := securitygroups.Get(client, d.Id()).Extract()
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("project_id", sg.ProjectID)
	d.Set("region_id", sg.RegionID)
	d.Set("name", sg.Name)
	d.Set("description", sg.Description)

	metadataMap := make(map[string]string)
	metadataReadOnly := make([]map[string]interface{}, 0, len(sg.Metadata))

	if len(sg.Metadata) > 0 {
		for _, metadataItem := range sg.Metadata {
			metadataMap[metadataItem.Key] = metadataItem.Value
			metadataReadOnly = append(metadataReadOnly, map[string]interface{}{
				"key":       metadataItem.Key,
				"value":     metadataItem.Value,
				"read_only": metadataItem.ReadOnly,
			})
		}
	}

	if err := d.Set("metadata_map", metadataMap); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("metadata_read_only", metadataReadOnly); err != nil {
		return diag.FromErr(err)
	}

	newSgRules := make([]interface{}, len(sg.SecurityGroupRules))
	for i, sgr := range sg.SecurityGroupRules {
		log.Printf("rules: %+v", sgr)
		r := make(map[string]interface{})
		r["id"] = sgr.ID
		r["direction"] = sgr.Direction.String()

		if sgr.EtherType != nil {
			r["ethertype"] = sgr.EtherType.String()
		}

		r["protocol"] = types.ProtocolAny
		if sgr.Protocol != nil {
			r["protocol"] = sgr.Protocol.String()
		}

		r["port_range_max"] = 0
		if sgr.PortRangeMax != nil {
			r["port_range_max"] = *sgr.PortRangeMax
		}
		r["port_range_min"] = 0
		if sgr.PortRangeMin != nil {
			r["port_range_min"] = *sgr.PortRangeMin
		}

		r["description"] = ""
		if sgr.Description != nil {
			r["description"] = *sgr.Description
		}

		r["remote_ip_prefix"] = ""
		if sgr.RemoteIPPrefix != nil {
			r["remote_ip_prefix"] = *sgr.RemoteIPPrefix
		}

		r["updated_at"] = sgr.UpdatedAt.String()
		r["created_at"] = sgr.CreatedAt.String()

		newSgRules[i] = r
	}

	if err := d.Set("security_group_rules", schema.NewSet(secGroupUniqueID, newSgRules)); err != nil {
		return diag.FromErr(err)
	}

	log.Println("[DEBUG] Finish SecurityGroup reading")

	return diags
}

func resourceSecurityGroupUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Println("[DEBUG] Start SecurityGroup updating")
	var valid bool
	vals := d.Get("security_group_rules").(*schema.Set).List()
	for _, val := range vals {
		rule := val.(map[string]interface{})
		if types.RuleDirection(rule["direction"].(string)) == types.RuleDirectionEgress {
			valid = true
			break
		}
	}
	if !valid {
		return diag.Errorf("at least one 'egress' rule should be set")
	}

	config := m.(*Config)
	provider := config.Provider
	clientCreate, err := CreateClient(provider, d, SecurityGroupPoint, VersionPointV1)
	if err != nil {
		return diag.FromErr(err)
	}

	clientUpdateDelete, err := CreateClient(provider, d, securityGroupRulesPoint, VersionPointV1)
	if err != nil {
		return diag.FromErr(err)
	}

	gid := d.Id()

	if d.HasChange("security_group_rules") {
		oldRulesRaw, newRulesRaw := d.GetChange("security_group_rules")
		oldRules := oldRulesRaw.(*schema.Set)
		newRules := newRulesRaw.(*schema.Set)

		changedRule := make(map[string]bool)
		for _, r := range newRules.List() {
			rule := r.(map[string]interface{})
			rid := rule["id"].(string)
			if !oldRules.Contains(r) && rid == "" {
				opts := extractSecurityGroupRuleMap(r, gid)
				_, err := securitygroups.AddRule(clientCreate, gid, opts).Extract()
				if err != nil {
					return diag.FromErr(err)
				}

				continue
			}
			if rid != "" && !oldRules.Contains(r) {
				changedRule[rid] = true
				opts := extractSecurityGroupRuleMap(r, gid)
				_, err := securitygrouprules.Replace(clientUpdateDelete, rid, opts).Extract()
				if err != nil {
					return diag.FromErr(err)
				}
			}
		}

		for _, r := range oldRules.List() {
			rule := r.(map[string]interface{})
			rid := rule["id"].(string)
			if !newRules.Contains(r) && !changedRule[rid] {
				// todo patch lib, should be task instead of DeleteResult
				err := securitygrouprules.Delete(clientUpdateDelete, rid).ExtractErr()
				if err != nil {
					return diag.FromErr(err)
				}
				// todo remove after patch lib
				time.Sleep(time.Second * 2)
				continue
			}
		}
	}

	if d.HasChange("metadata_map") {
		_, nmd := d.GetChange("metadata_map")

		err := securitygroups.MetadataReplace(clientCreate, gid, nmd.(map[string]interface{})).Err
		if err != nil {
			return diag.Errorf("cannot update metadata. Error: %s", err)
		}
	}

	d.Set("last_updated", time.Now().Format(time.RFC850))
	log.Println("[DEBUG] Finish SecurityGroup updating")

	return resourceSecurityGroupRead(ctx, d, m)
}

func resourceSecurityGroupDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Println("[DEBUG] Start SecurityGroup deleting")
	var diags diag.Diagnostics
	config := m.(*Config)
	provider := config.Provider
	sgID := d.Id()

	client, err := CreateClient(provider, d, SecurityGroupPoint, VersionPointV1)
	if err != nil {
		return diag.FromErr(err)
	}

	err = securitygroups.Delete(client, sgID).Err
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	log.Printf("[DEBUG] Finish of SecurityGroup deleting")

	return diags
}
