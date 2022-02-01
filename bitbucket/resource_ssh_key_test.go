package bitbucket

import (
	"fmt"
	"net/url"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccBitbucketSshKey_basic(t *testing.T) {
	var sshKey SshKey
	resourceName := "bitbucket_ssh_key.test"

	userEmail := os.Getenv("BITBUCKET_USERNAME")
	publicKey, _, err := RandSSHKeyPairSize(2048, userEmail)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBitbucketSshKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBitbucketSshKeyConfig(publicKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBitbucketSshKeyExists(resourceName, &sshKey),
					resource.TestCheckResourceAttrPair(resourceName, "user", "data.bitbucket_current_user.test", "uuid"),
					resource.TestCheckResourceAttr(resourceName, "key", publicKey),
					resource.TestCheckResourceAttr(resourceName, "comment", userEmail),
					resource.TestCheckResourceAttrSet(resourceName, "uuid"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"key"},
			},
		},
	})
}

func TestAccBitbucketSshKey_label(t *testing.T) {
	var sshKey SshKey
	resourceName := "bitbucket_ssh_key.test"

	rName := acctest.RandomWithPrefix("tf-test")
	rName2 := acctest.RandomWithPrefix("tf-test")
	userEmail := os.Getenv("BITBUCKET_USERNAME")
	publicKey, _, err := RandSSHKeyPairSize(2048, userEmail)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBitbucketSshKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBitbucketSshKeyLabelConfig(publicKey, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBitbucketSshKeyExists(resourceName, &sshKey),
					resource.TestCheckResourceAttrPair(resourceName, "user", "data.bitbucket_current_user.test", "uuid"),
					resource.TestCheckResourceAttr(resourceName, "key", publicKey),
					resource.TestCheckResourceAttr(resourceName, "label", rName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"key"},
			},
			{
				Config: testAccBitbucketSshKeyLabelConfig(publicKey, rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBitbucketSshKeyExists(resourceName, &sshKey),
					resource.TestCheckResourceAttrPair(resourceName, "user", "data.bitbucket_current_user.test", "uuid"),
					resource.TestCheckResourceAttr(resourceName, "key", publicKey),
					resource.TestCheckResourceAttr(resourceName, "label", rName2),
				),
			},
		},
	})
}

func testAccCheckBitbucketSshKeyDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*Client)
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bitbucket_ssh_key" {
			continue
		}

		response, err := client.Get(fmt.Sprintf("2.0/repositories/%s/%s/ssh_keys/%s", rs.Primary.Attributes["owner"], rs.Primary.Attributes["repository"], url.PathEscape(rs.Primary.Attributes["uuid"])))

		if err == nil {
			return fmt.Errorf("The resource was found should have errored")
		}

		if response.StatusCode != 404 {
			return fmt.Errorf("Ssh Key still exists")
		}

	}
	return nil
}

func testAccCheckBitbucketSshKeyExists(n string, sshKey *SshKey) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Ssh Key ID is set")
		}
		return nil
	}
}

func testAccBitbucketSshKeyConfig(pubkey string) string {
	return fmt.Sprintf(`
data "bitbucket_current_user" "test" {}

resource "bitbucket_ssh_key" "test" {
  user = data.bitbucket_current_user.test.uuid
  key  = %[1]q
}
`, pubkey)
}

func testAccBitbucketSshKeyLabelConfig(pubkey, label string) string {
	return fmt.Sprintf(`
data "bitbucket_current_user" "test" {}

resource "bitbucket_ssh_key" "test" {
  user  = data.bitbucket_current_user.test.uuid
  key   = %[1]q
  label = %[2]q
}
`, pubkey, label)
}
