---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "edgecenter_k8s Data Source - edgecenter"
subcategory: ""
description: |-
  Represent k8s cluster with one default pool.
---

# edgecenter_k8s (Data Source)

Represent k8s cluster with one default pool.

## Example Usage

```terraform
provider "edgecenter" {
  permanent_api_token = "251$d3361.............1b35f26d8"
}

data "edgecenter_k8s" "cluster" {
  project_id = 1
  region_id  = 1
  cluster_id = "dc3a3ea9-86ae-47ad-a8e8-79df0ce04839"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `cluster_id` (String) The uuid of the Kubernetes cluster.

### Optional

- `project_id` (Number) The uuid of the project. Either 'project_id' or 'project_name' must be specified.
- `project_name` (String) The name of the project. Either 'project_id' or 'project_name' must be specified.
- `region_id` (Number) The uuid of the region. Either 'region_id' or 'region_name' must be specified.
- `region_name` (String) The name of the region. Either 'region_id' or 'region_name' must be specified.

### Read-Only

- `api_address` (String) API endpoint address for the Kubernetes cluster.
- `auto_healing_enabled` (Boolean) Indicates whether auto-healing is enabled for the Kubernetes cluster.
- `certificate_authority_data` (String) The certificate_authority_data field from the Kubernetes cluster config.
- `cluster_template_id` (String) Template identifier from which the Kubernetes cluster was instantiated.
- `container_version` (String) The container runtime version used in the Kubernetes cluster.
- `created_at` (String) The timestamp when the Kubernetes cluster was created.
- `discovery_url` (String) URL used for node discovery within the Kubernetes cluster.
- `faults` (Map of String)
- `fixed_network` (String) Fixed network (uuid) associated with the Kubernetes cluster.
- `fixed_subnet` (String) Subnet (uuid) associated with the fixed network.
- `health_status` (String) Overall health status of the Kubernetes cluster.
- `health_status_reason` (Map of String)
- `id` (String) The ID of this resource.
- `keypair` (String)
- `master_addresses` (List of String) List of IP addresses for master nodes in the Kubernetes cluster.
- `master_flavor_id` (String) Identifier for the master node flavor in the Kubernetes cluster.
- `master_lb_floating_ip_enabled` (Boolean) Flag indicating if the master LoadBalancer should have a floating IP.
- `name` (String) The name of the Kubernetes cluster.
- `node_addresses` (List of String) List of IP addresses for worker nodes in the Kubernetes cluster.
- `node_count` (Number) Total number of nodes in the Kubernetes cluster.
- `pool` (List of Object) Configuration details of the node pool in the Kubernetes cluster. (see [below for nested schema](#nestedatt--pool))
- `status` (String) The current status of the Kubernetes cluster.
- `status_reason` (String) The reason for the current status of the Kubernetes cluster, if ERROR.
- `updated_at` (String) The timestamp when the Kubernetes cluster was updated.
- `user_id` (String) User identifier associated with the Kubernetes cluster.
- `version` (String) The version of the Kubernetes cluster.

<a id="nestedatt--pool"></a>
### Nested Schema for `pool`

Read-Only:

- `created_at` (String)
- `docker_volume_size` (Number)
- `docker_volume_type` (String)
- `flavor_id` (String)
- `max_node_count` (Number)
- `min_node_count` (Number)
- `name` (String)
- `node_count` (Number)
- `stack_id` (String)
- `uuid` (String)


