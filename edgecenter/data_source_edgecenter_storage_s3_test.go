//go:build !cloud
// +build !cloud

package edgecenter

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/Edge-Center/edgecenter-storage-sdk-go/swagger/client/storage"
)

func TestStorageS3DataSource(t *testing.T) {
	random := time.Now().Nanosecond()
	resourceName := fmt.Sprintf("edgecenter_storage_s3.terraformtest%d_s3", random)
	dataSourceName := fmt.Sprintf("data.edgecenter_storage_s3.terraformtest%d_s3_data", random)

	templateCreate := func() string {
		return fmt.Sprintf(`
resource "edgecenter_storage_s3" "terraformtest%d_s3" {
  name = "terraformtest%d"
  location = "s-ed1"
}
		`, random, random)
	}

	templateRead := func() string {
		return fmt.Sprintf(`
data "edgecenter_storage_s3" "terraformtest%d_s3_data" {
  name = "terraformtest%d"
}
		`, random, random)
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckVars(t, EC_USERNAME_VAR, EC_PASSWORD_VAR, EC_STORAGE_URL_VAR)
		},
		CheckDestroy: func(s *terraform.State) error {
			config := testAccProvider.Meta().(*Config)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			for _, rs := range s.RootModule().Resources {
				if rs.Type != "edgecenter_storage_s3" {
					continue
				}
				opts := []func(opt *storage.StorageListHTTPV2Params){
					func(opt *storage.StorageListHTTPV2Params) { opt.Context = ctx },
					func(opt *storage.StorageListHTTPV2Params) { opt.ID = &rs.Primary.ID },
				}
				storages, err := config.StorageClient.StoragesList(opts...)
				if err != nil {
					return fmt.Errorf("find storage: %w", err)
				}
				if len(storages) == 0 {
					return nil
				}
				if storages[0].ProvisioningStatus == "ok" {
					return fmt.Errorf("storage #%s wasn't deleted correctrly", rs.Primary.ID)
				}
			}
			return nil
		},
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: templateCreate(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, StorageSchemaLocation, "s-ed1"),
				),
			},
			{
				Config: templateRead(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceExists(dataSourceName),
					resource.TestCheckResourceAttr(dataSourceName, StorageSchemaLocation, "s-ed1"),
				),
			},
		},
	})
}
