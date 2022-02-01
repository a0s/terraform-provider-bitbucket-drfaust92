package bitbucket

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccBitbucketDefaultReviewers_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-test")
	owner := os.Getenv("BITBUCKET_TEAM")
	testUser := os.Getenv("BITBUCKET_USERNAME")
	resourceName := "bitbucket_default_reviewers.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBitbucketDefaultReviewersDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBitbucketDefaultReviewersConfig(owner, testUser, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBitbucketDefaultReviewersExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "repository", "bitbucket_repository.test", "name"),
					resource.TestCheckResourceAttr(resourceName, "reviewers.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "reviewers.*", "data.bitbucket_current_user.test", "uuid"),
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

func testAccBitbucketDefaultReviewersConfig(owner, user, rName string) string {
	return fmt.Sprintf(`
data "bitbucket_current_user" "test" {}

resource "bitbucket_repository" "test" {
  owner = %[1]q
  name  = %[3]q
}

resource "bitbucket_default_reviewers" "test" {
  owner      = %[1]q
  repository = bitbucket_repository.test.name
  reviewers  = [data.bitbucket_current_user.test.uuid]
}
`, owner, user, rName)
}

func testAccCheckBitbucketDefaultReviewersDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*Client)
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bitbucket_default_reviewers" {
			continue
		}
		response, _ := client.Get(fmt.Sprintf("2.0/repositories/%s/%s/default-reviewers", rs.Primary.Attributes["owner"], rs.Primary.Attributes["repository"]))

		if response.StatusCode != 404 {
			return fmt.Errorf("Defaults Reviewer still exists")
		}
	}
	return nil
}

func testAccCheckBitbucketDefaultReviewersExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No default reviewers ID is set")
		}

		return nil
	}
}
