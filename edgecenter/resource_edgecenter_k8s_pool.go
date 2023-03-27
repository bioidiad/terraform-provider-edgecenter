package edgecenter

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	edgecloud "github.com/Edge-Center/edgecentercloud-go"
	"github.com/Edge-Center/edgecentercloud-go/edgecenter/k8s/v1/clusters"
	"github.com/Edge-Center/edgecentercloud-go/edgecenter/k8s/v1/pools"
	"github.com/Edge-Center/edgecentercloud-go/edgecenter/task/v1/tasks"
	"github.com/Edge-Center/edgecentercloud-go/edgecenter/volume/v1/volumes"
)

func resourceK8sPool() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceK8sPoolCreate,
		ReadContext:   resourceK8sPoolRead,
		UpdateContext: resourceK8sPoolUpdate,
		DeleteContext: resourceK8sPoolDelete,
		Description:   "Represent k8s cluster's pool.",
		Timeouts: &schema.ResourceTimeout{
			Create: &k8sCreateTimeout,
			Update: &k8sCreateTimeout,
		},
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				projectID, regionID, poolID, clusterID, err := ImportStringParserExtended(d.Id())
				if err != nil {
					return nil, err
				}
				d.Set("project_id", projectID)
				d.Set("region_id", regionID)
				d.Set("cluster_id", clusterID)
				d.SetId(poolID)

				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:     schema.TypeInt,
				Optional: true,
				ExactlyOneOf: []string{
					"project_id",
					"project_name",
				},
			},
			"region_id": {
				Type:     schema.TypeInt,
				Optional: true,
				ExactlyOneOf: []string{
					"region_id",
					"region_name",
				},
			},
			"project_name": {
				Type:     schema.TypeString,
				Optional: true,
				ExactlyOneOf: []string{
					"project_id",
					"project_name",
				},
			},
			"region_name": {
				Type:     schema.TypeString,
				Optional: true,
				ExactlyOneOf: []string{
					"region_id",
					"region_name",
				},
			},
			"cluster_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"flavor_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"min_node_count": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"max_node_count": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"node_count": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"docker_volume_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Available value is 'standard', 'ssd_hiiops', 'cold', 'ultra'.",
			},
			"docker_volume_size": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"stack_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_updated": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceK8sPoolCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Println("[DEBUG] Start K8s pool creating")
	var diags diag.Diagnostics
	config := m.(*Config)
	provider := config.Provider

	client, err := CreateClient(provider, d, K8sPoint, VersionPointV1)
	if err != nil {
		return diag.FromErr(err)
	}

	opts := pools.CreateOpts{
		Name:         d.Get("name").(string),
		FlavorID:     d.Get("flavor_id").(string),
		NodeCount:    d.Get("node_count").(int),
		MinNodeCount: d.Get("min_node_count").(int),
		MaxNodeCount: d.Get("max_node_count").(int),
	}

	dockerVolumeSize := d.Get("docker_volume_size").(int)
	if dockerVolumeSize != 0 {
		opts.DockerVolumeSize = dockerVolumeSize
	}

	dockerVolumeType := d.Get("docker_volume_type").(string)
	if dockerVolumeType != "" {
		opts.DockerVolumeType = volumes.VolumeType(dockerVolumeType)
	}

	clusterID := d.Get("cluster_id").(string)
	results, err := pools.Create(client, clusterID, opts).Extract()
	if err != nil {
		return diag.FromErr(err)
	}

	taskID := results.Tasks[0]
	log.Printf("[DEBUG] Task id (%s)", taskID)
	poolID, err := tasks.WaitTaskAndReturnResult(client, taskID, true, K8sCreateTimeout, func(task tasks.TaskID) (interface{}, error) {
		taskInfo, err := tasks.Get(client, string(task)).Extract()
		if err != nil {
			return nil, fmt.Errorf("cannot get task with ID: %s. Error: %w", task, err)
		}
		poolID, err := pools.ExtractClusterPoolIDFromTask(taskInfo)
		if err != nil {
			return nil, fmt.Errorf("cannot retrieve k8s pool ID from task info: %w", err)
		}
		return poolID, nil
	},
	)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(poolID.(string))
	resourceK8sPoolRead(ctx, d, m)

	log.Printf("[DEBUG] Finish K8s pool creating (%s)", poolID)

	return diags
}

func resourceK8sPoolRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Println("[DEBUG] Start K8s pool reading")
	var diags diag.Diagnostics
	config := m.(*Config)
	provider := config.Provider

	client, err := CreateClient(provider, d, K8sPoint, VersionPointV1)
	if err != nil {
		return diag.FromErr(err)
	}

	clusterID := d.Get("cluster_id").(string)
	poolID := d.Id()

	pool, err := pools.Get(client, clusterID, poolID).Extract()
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("name", pool.Name)
	d.Set("cluster_id", pool.ClusterID)
	d.Set("flavor_id", pool.FlavorID)
	d.Set("min_node_count", pool.MinNodeCount)
	d.Set("max_node_count", pool.MaxNodeCount)
	d.Set("node_count", pool.NodeCount)
	d.Set("docker_volume_type", pool.DockerVolumeType.String())
	d.Set("docker_volume_size", pool.DockerVolumeSize)
	d.Set("stack_id", pool.StackID)
	d.Set("created_at", pool.CreatedAt.Format(time.RFC850))

	log.Println("[DEBUG] Finish K8s pool reading")

	return diags
}

func resourceK8sPoolUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Println("[DEBUG] Start K8s updating")
	config := m.(*Config)
	provider := config.Provider

	client, err := CreateClient(provider, d, K8sPoint, VersionPointV1)
	if err != nil {
		return diag.FromErr(err)
	}

	poolID := d.Id()
	clusterID := d.Get("cluster_id").(string)

	if d.HasChanges("name", "min_node_count", "max_node_count") {
		updateOpts := pools.UpdateOpts{
			Name:         d.Get("name").(string),
			MinNodeCount: d.Get("min_node_count").(int),
			MaxNodeCount: d.Get("max_node_count").(int),
		}
		results, err := pools.Update(client, clusterID, poolID, updateOpts).Extract()
		if err != nil {
			return diag.FromErr(err)
		}

		taskID := results.Tasks[0]
		_, err = tasks.WaitTaskAndReturnResult(client, taskID, true, K8sCreateTimeout, func(task tasks.TaskID) (interface{}, error) {
			_, err := pools.Get(client, clusterID, poolID).Extract()
			if err != nil {
				return nil, fmt.Errorf("cannot get pool with ID: %s. Error: %w", poolID, err)
			}
			return nil, nil
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("node_count") {
		resizeOpts := clusters.ResizeOpts{
			NodeCount: d.Get("node_count").(int),
		}
		results, err := clusters.Resize(client, clusterID, poolID, resizeOpts).Extract()
		if err != nil {
			return diag.FromErr(err)
		}

		taskID := results.Tasks[0]
		_, err = tasks.WaitTaskAndReturnResult(client, taskID, true, K8sCreateTimeout, func(task tasks.TaskID) (interface{}, error) {
			_, err := pools.Get(client, clusterID, poolID).Extract()
			if err != nil {
				return nil, fmt.Errorf("cannot get pool with ID: %s. Error: %w", poolID, err)
			}
			return nil, nil
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceK8sPoolRead(ctx, d, m)
}

func resourceK8sPoolDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Println("[DEBUG] Start K8s deleting")
	var diags diag.Diagnostics
	config := m.(*Config)
	provider := config.Provider

	client, err := CreateClient(provider, d, K8sPoint, VersionPointV1)
	if err != nil {
		return diag.FromErr(err)
	}

	id := d.Id()
	clusterID := d.Get("cluster_id").(string)
	results, err := pools.Delete(client, clusterID, id).Extract()
	if err != nil {
		return diag.FromErr(err)
	}

	taskID := results.Tasks[0]
	_, err = tasks.WaitTaskAndReturnResult(client, taskID, true, K8sCreateTimeout, func(task tasks.TaskID) (interface{}, error) {
		_, err := pools.Get(client, clusterID, id).Extract()
		if err == nil {
			return nil, fmt.Errorf("cannot delete k8s cluster pool with ID: %s", id)
		}
		var errDefault404 edgecloud.ErrDefault404
		if errors.As(err, &errDefault404) {
			return nil, nil
		}
		return nil, fmt.Errorf("extracting Pool resource error: %w", err)
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	log.Printf("[DEBUG] Finish of K8s pool deleting")

	return diags
}
