resource "edgecenter_loadbalancer" "lb" {
  region_id  = var.region_id
  project_id = var.project_id
  // other fields
}

# Example 1
data "edgecenter_lbpool" "pool1" {
  region_id       = var.region_id
  project_id      = var.project_id
  name            = "test-lbpool"
  loadbalancer_id = edgecenter_loadbalancer.lb.id
}

output "pool1" {
  value = data.edgecenter_lbpool.pool1
}

# Example 2
data "edgecenter_lbpool" "pool2" {
  region_id       = var.region_id
  project_id      = var.project_id
  id              = "00000000-0000-0000-0000-000000000000"
  loadbalancer_id = edgecenter_loadbalancer.lb.id
}

output "pool2" {
  value = data.edgecenter_lbpool.pool2
}
