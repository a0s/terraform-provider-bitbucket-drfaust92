package bitbucket

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccBitbucketProject_basic(t *testing.T) {
	var project Project

	resourceName := "bitbucket_project.test"
	testTeam := os.Getenv("BITBUCKET_TEAM")
	rName := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBitbucketProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBitbucketProjectConfig(testTeam, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBitbucketProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "has_publicly_visible_repos", "false"),
					resource.TestCheckResourceAttr(resourceName, "key", "AAAAAA"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "owner", testTeam),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "is_private", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "uuid"),
					resource.TestCheckResourceAttr(resourceName, "link.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "link.0.avatar.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "link.0.avatar.0.href"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccBitbucketProject_avatar(t *testing.T) {
	var project Project

	resourceName := "bitbucket_project.test"
	testTeam := os.Getenv("BITBUCKET_TEAM")
	rName := acctest.RandomWithPrefix("tf-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBitbucketProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBitbucketProjectAvatarConfig(testTeam, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBitbucketProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "link.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "link.0.avatar.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "link.0.avatar.0.href"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccBitbucketProjectConfig(team, rName string) string {
	return fmt.Sprintf(`
resource "bitbucket_project" "test" {
  owner = %[1]q
  name  = %[2]q
  key   = "AAAAAA"
}
`, team, rName)
}

func testAccBitbucketProjectAvatarConfig(team, rName string) string {
	return fmt.Sprintf(`
resource "bitbucket_project" "test" {
  owner = %[1]q
  name  = %[2]q
  key   = "BBBBB"

  link {
    avatar {
      href = "https://d301sr5gafysq2.cloudfront.net/dfb18959be9c/img/repo-avatars/python.png"
	}
  }
}
`, team, rName)
}

func testAccCheckBitbucketProjectDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*Client)
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bitbucket_project" {
			continue
		}
		response, err := client.Get(fmt.Sprintf("2.0/workspaces/%s/projects/%s", rs.Primary.Attributes["owner"], rs.Primary.Attributes["name"]))

		if err == nil {
			return fmt.Errorf("The resource was found should have errored")
		}

		if response.StatusCode != 404 {
			return fmt.Errorf("Project still exists")
		}
	}
	return nil
}

func testAccCheckBitbucketProjectExists(n string, project *Project) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No project ID is set")
		}
		return nil
	}
}
