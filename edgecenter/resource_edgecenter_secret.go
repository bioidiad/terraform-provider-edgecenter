package edgecenter

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	edgecloud "github.com/Edge-Center/edgecentercloud-go"
	"github.com/Edge-Center/edgecentercloud-go/edgecenter/secret/v1/secrets"
	secretsV2 "github.com/Edge-Center/edgecentercloud-go/edgecenter/secret/v2/secrets"
	"github.com/Edge-Center/edgecentercloud-go/edgecenter/task/v1/tasks"
)

const (
	SecretDeleting        int = 1200
	SecretCreatingTimeout int = 1200
	SecretPoint               = "secrets"
)

func resourceSecret() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSecretCreate,
		ReadContext:   resourceSecretRead,
		DeleteContext: resourceSecretDelete,
		Description:   "Represent secret",
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				projectID, regionID, secretID, err := ImportStringParser(d.Id())
				if err != nil {
					return nil, err
				}
				d.Set("project_id", projectID)
				d.Set("region_id", regionID)
				d.SetId(secretID)

				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				Description:  "The uuid of the project. Either 'project_id' or 'project_name' must be specified.",
				ExactlyOneOf: []string{"project_id", "project_name"},
			},
			"project_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Description:  "The name of the project. Either 'project_id' or 'project_name' must be specified.",
				ExactlyOneOf: []string{"project_id", "project_name"},
			},
			"region_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				Description:  "The uuid of the region. Either 'region_id' or 'region_name' must be specified.",
				ExactlyOneOf: []string{"region_id", "region_name"},
			},
			"region_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Description:  "The name of the region. Either 'region_id' or 'region_name' must be specified.",
				ExactlyOneOf: []string{"region_id", "region_name"},
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the secret.",
			},
			"private_key": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "SSL private key in PEM format",
			},
			"certificate_chain": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "SSL certificate chain of intermediates and root certificates in PEM format",
			},
			"certificate": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "SSL certificate in PEM format",
			},
			"algorithm": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The encryption algorithm used for the secret.",
			},
			"bit_length": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The bit length of the encryption algorithm.",
			},
			"mode": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The mode of the encryption algorithm.",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The current status of the secret.",
			},
			"content_types": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "The content types associated with the secret's payload.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"expiration": {
				Type:        schema.TypeString,
				Description: "Datetime when the secret will expire. The format is 2025-12-28T19:14:44",
				Optional:    true,
				Computed:    true,
				StateFunc: func(val interface{}) string {
					expTime, _ := time.Parse(edgecloud.RFC3339NoZ, val.(string))
					return expTime.Format(edgecloud.RFC3339NoZ)
				},
				ValidateDiagFunc: func(i interface{}, path cty.Path) diag.Diagnostics {
					rawTime := i.(string)
					_, err := time.Parse(edgecloud.RFC3339NoZ, rawTime)
					if err != nil {
						return diag.FromErr(err)
					}
					return nil
				},
			},
			"created": {
				Type:        schema.TypeString,
				Description: "Datetime when the secret was created. The format is 2025-12-28T19:14:44.180394",
				Computed:    true,
			},
		},
	}
}

func resourceSecretCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Println("[DEBUG] Start Secret creating")
	var diags diag.Diagnostics
	config := m.(*Config)
	provider := config.Provider

	client, err := CreateClient(provider, d, SecretPoint, VersionPointV2)
	if err != nil {
		return diag.FromErr(err)
	}

	opts := secretsV2.CreateOpts{
		Name: d.Get("name").(string),
		Payload: secretsV2.PayloadOpts{
			CertificateChain: d.Get("certificate_chain").(string),
			Certificate:      d.Get("certificate").(string),
			PrivateKey:       d.Get("private_key").(string),
		},
	}
	if rawTime := d.Get("expiration").(string); rawTime != "" {
		expiration, err := time.Parse(edgecloud.RFC3339NoZ, rawTime)
		if err != nil {
			return diag.FromErr(err)
		}
		opts.Expiration = &expiration
	}

	results, err := secretsV2.Create(client, opts).Extract()
	if err != nil {
		return diag.FromErr(err)
	}

	taskID := results.Tasks[0]
	log.Printf("[DEBUG] Task id (%s)", taskID)

	clientV1, err := CreateClient(provider, d, SecretPoint, VersionPointV1)
	if err != nil {
		return diag.FromErr(err)
	}
	secretID, err := tasks.WaitTaskAndReturnResult(clientV1, taskID, true, SecretCreatingTimeout, func(task tasks.TaskID) (interface{}, error) {
		taskInfo, err := tasks.Get(clientV1, string(task)).Extract()
		if err != nil {
			return nil, fmt.Errorf("cannot get task with ID: %s. Error: %w", task, err)
		}
		Secret, err := secrets.ExtractSecretIDFromTask(taskInfo)
		if err != nil {
			return nil, fmt.Errorf("cannot retrieve Secret ID from task info: %w", err)
		}
		return Secret, nil
	},
	)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[DEBUG] Secret id (%s)", secretID)

	d.SetId(secretID.(string))

	resourceSecretRead(ctx, d, m)

	log.Printf("[DEBUG] Finish Secret creating (%s)", secretID)

	return diags
}

func resourceSecretRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Println("[DEBUG] Start secret reading")
	var diags diag.Diagnostics
	config := m.(*Config)
	provider := config.Provider
	secretID := d.Id()
	log.Printf("[DEBUG] Secret id = %s", secretID)

	client, err := CreateClient(provider, d, SecretPoint, VersionPointV1)
	if err != nil {
		return diag.FromErr(err)
	}

	secret, err := secrets.Get(client, secretID).Extract()
	if err != nil {
		return diag.Errorf("cannot get secret with ID: %s. Error: %s", secretID, err.Error())
	}
	d.Set("name", secret.Name)
	d.Set("algorithm", secret.Algorithm)
	d.Set("bit_length", secret.BitLength)
	d.Set("mode", secret.Mode)
	d.Set("status", secret.Status)
	d.Set("expiration", secret.Expiration.Format(edgecloud.RFC3339NoZ))
	d.Set("created", secret.CreatedAt.Format(edgecloud.RFC3339MilliNoZ))
	if err := d.Set("content_types", secret.ContentTypes); err != nil {
		return diag.FromErr(err)
	}

	log.Println("[DEBUG] Finish secret reading")

	return diags
}

func resourceSecretDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Println("[DEBUG] Start secret deleting")
	var diags diag.Diagnostics
	config := m.(*Config)
	provider := config.Provider
	secretID := d.Id()
	log.Printf("[DEBUG] Secret id = %s", secretID)

	client, err := CreateClient(provider, d, SecretPoint, VersionPointV1)
	if err != nil {
		return diag.FromErr(err)
	}

	results, err := secrets.Delete(client, secretID).Extract()
	if err != nil {
		return diag.FromErr(err)
	}
	taskID := results.Tasks[0]
	log.Printf("[DEBUG] Task id (%s)", taskID)
	_, err = tasks.WaitTaskAndReturnResult(client, taskID, true, SecretDeleting, func(task tasks.TaskID) (interface{}, error) {
		_, err := secrets.Get(client, secretID).Extract()
		if err == nil {
			return nil, fmt.Errorf("cannot delete secret with ID: %s", secretID)
		}
		return nil, nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	log.Printf("[DEBUG] Finish of secret deleting")

	return diags
}
