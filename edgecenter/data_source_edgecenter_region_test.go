//go:build cloud
// +build cloud

package edgecenter

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/Edge-Center/edgecentercloud-go/edgecenter/region/v1/regions"
)

func TestAccRegionDataSource(t *testing.T) {
	cfg, err := createTestConfig()
	if err != nil {
		t.Fatal(err)
	}

	client, err := CreateTestClient(cfg.Provider, regionPoint, versionPointV1)
	if err != nil {
		t.Fatal(err)
	}

	rs, err := regions.ListAll(client)
	if err != nil {
		t.Fatal(err)
	}

	if len(rs) == 0 {
		t.Fatal("regions not found")
	}

	region := rs[0]

	fullName := "data.edgecenter_region.acctest"
	tpl := func(name string) string {
		return fmt.Sprintf(`
			data "edgecenter_region" "acctest" {
              name = "%s"
			}
		`, name)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: tpl(region.DisplayName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceExists(fullName),
					resource.TestCheckResourceAttr(fullName, "name", region.DisplayName),
					resource.TestCheckResourceAttr(fullName, "id", strconv.Itoa(region.ID)),
				),
			},
		},
	})
}
