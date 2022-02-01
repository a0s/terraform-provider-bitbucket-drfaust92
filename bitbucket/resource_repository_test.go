package bitbucket

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccBitbucketRepository_basic(t *testing.T) {
	var repo Repository

	rName := acctest.RandomWithPrefix("tf-test")
	testUser := os.Getenv("BITBUCKET_TEAM")
	resourceName := "bitbucket_repository.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBitbucketRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBitbucketRepoConfig(testUser, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBitbucketRepositoryExists(resourceName, &repo),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "owner", testUser),
					resource.TestCheckResourceAttr(resourceName, "scm", "git"),
					resource.TestCheckResourceAttr(resourceName, "has_wiki", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "uuid"),
					resource.TestCheckResourceAttr(resourceName, "fork_policy", "allow_forks"),
					resource.TestCheckResourceAttr(resourceName, "language", ""),
					resource.TestCheckResourceAttr(resourceName, "has_issues", "false"),
					resource.TestCheckResourceAttr(resourceName, "slug", rName),
					resource.TestCheckResourceAttr(resourceName, "is_private", "true"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
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

func TestAccBitbucketRepository_avatar(t *testing.T) {
	var repo Repository

	rName := acctest.RandomWithPrefix("tf-test")
	testUser := os.Getenv("BITBUCKET_TEAM")
	resourceName := "bitbucket_repository.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBitbucketRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBitbucketRepoAvatarConfig(testUser, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBitbucketRepositoryExists(resourceName, &repo),
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

func TestAccBitbucketRepository_camelcase(t *testing.T) {
	var repo Repository

	rName := acctest.RandomWithPrefix("tf-test")
	rName2 := acctest.RandomWithPrefix("tf-test-2")
	testUser := os.Getenv("BITBUCKET_TEAM")
	resourceName := "bitbucket_repository.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBitbucketRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBitbucketRepoSlugConfig(testUser, rName, rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBitbucketRepositoryExists(resourceName, &repo),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "owner", testUser),
					resource.TestCheckResourceAttr(resourceName, "scm", "git"),
					resource.TestCheckResourceAttr(resourceName, "has_wiki", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "uuid"),
					resource.TestCheckResourceAttr(resourceName, "fork_policy", "allow_forks"),
					resource.TestCheckResourceAttr(resourceName, "language", ""),
					resource.TestCheckResourceAttr(resourceName, "has_issues", "false"),
					resource.TestCheckResourceAttr(resourceName, "slug", rName2),
					resource.TestCheckResourceAttr(resourceName, "is_private", "true"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
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

func testAccBitbucketRepoConfig(testUser, rName string) string {
	return fmt.Sprintf(`
resource "bitbucket_repository" "test" {
  owner = %[1]q
  name  = %[2]q
}
`, testUser, rName)
}

func testAccBitbucketRepoAvatarConfig(testUser, rName string) string {
	return fmt.Sprintf(`
resource "bitbucket_repository" "test" {
  owner = %[1]q
  name  = %[2]q

  link {
    avatar {
      href = "https://d301sr5gafysq2.cloudfront.net/dfb18959be9c/img/repo-avatars/python.png"
	}
  }  
}
`, testUser, rName)
}

func testAccBitbucketRepoSlugConfig(testUser, rName, rName2 string) string {
	return fmt.Sprintf(`
resource "bitbucket_repository" "test" {
  owner = %[1]q
  name  = %[2]q
  slug  = %[3]q
}
`, testUser, rName, rName2)
}

func testAccCheckBitbucketRepositoryDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*Client)
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bitbucket_repository" {
			continue
		}
		response, err := client.Get(fmt.Sprintf("2.0/repositories/%s/%s", rs.Primary.Attributes["owner"], rs.Primary.Attributes["name"]))

		if err == nil {
			return fmt.Errorf("The resource was found should have errored")
		}

		if response.StatusCode != 404 {
			return fmt.Errorf("Repository still exists")
		}
	}
	return nil
}

func testAccCheckBitbucketRepositoryExists(n string, repository *Repository) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No repository ID is set")
		}
		return nil
	}
}
