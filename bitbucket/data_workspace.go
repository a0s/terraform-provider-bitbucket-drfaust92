package bitbucket

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type Workspace struct {
	Slug      string `json:"slug"`
	IsPrivate bool   `json:"is_private"`
	Name      string `json:"name"`
	UUID      string `json:"uuid"`
}

func dataWorkspace() *schema.Resource {
	return &schema.Resource{
		Read: dataReadWorkspace,

		Schema: map[string]*schema.Schema{
			"workspace": {
				Type:     schema.TypeString,
				Required: true,
			},
			"slug": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"is_private": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataReadWorkspace(d *schema.ResourceData, m interface{}) error {
	c := m.(*Client)

	workspace := d.Get("workspace").(string)
	workspaceReq, err := c.Get(fmt.Sprintf("2.0/workspaces/%s", workspace))
	if err != nil {
		return err
	}

	if workspaceReq.StatusCode == http.StatusNotFound {
		return fmt.Errorf("user not found")
	}

	if workspaceReq.StatusCode >= http.StatusInternalServerError {
		return fmt.Errorf("internal server error fetching user")
	}

	body, readerr := ioutil.ReadAll(workspaceReq.Body)
	if readerr != nil {
		return readerr
	}

	log.Printf("[DEBUG] Workspace Response JSON: %v", string(body))

	var work Workspace

	decodeerr := json.Unmarshal(body, &work)
	if decodeerr != nil {
		return decodeerr
	}

	log.Printf("[DEBUG] Workspace Response Decoded: %#v", work)

	d.SetId(work.UUID)
	d.Set("workspace", workspace)
	d.Set("name", work.Name)
	d.Set("slug", work.Slug)
	d.Set("is_private", work.IsPrivate)

	return nil
}
