package bitbucket

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccUser_basic(t *testing.T) {
	dataSourceName := "data.bitbucket_user.test"
	testUser := os.Getenv("BITBUCKET_USER")

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccBitbucketUserConfig(testUser),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "uuid"),
					resource.TestCheckResourceAttrSet(dataSourceName, "nickname"),
					resource.TestCheckResourceAttrSet(dataSourceName, "display_name"),
					resource.TestCheckResourceAttrSet(dataSourceName, "account_id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "account_status"),
					resource.TestCheckResourceAttr(dataSourceName, "is_staff", "false"),
					resource.TestCheckResourceAttr(dataSourceName, "username", testUser),
				),
			},
		},
	})
}

func testAccBitbucketUserConfig(user string) string {
	return fmt.Sprintf(`
data "bitbucket_user" "test" {
  username = %[1]q
}
`, user)
}
