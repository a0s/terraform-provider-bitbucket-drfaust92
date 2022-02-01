package bitbucket

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type UserGroup struct {
	Name string `json:"name,omitempty"`
	Slug string `json:"slug,omitempty"`
}

func resourceGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceGroupsCreate,
		Read:   resourceGroupsRead,
		Update: resourceGroupsUpdate,
		Delete: resourceGroupsDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"workspace": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				// ForceNew: true,
			},
			"slug": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceGroupsCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*Client)

	group := expandGroup(d)
	log.Printf("[DEBUG] Group Request: %#v", group)

	workspace := d.Get("workspace").(string)
	body := []byte(fmt.Sprintf("name=%s", group.Name))
	groupReq, err := client.PostNonJson(fmt.Sprintf("1.0/groups/%s", workspace), bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	body, readerr := ioutil.ReadAll(groupReq.Body)
	if readerr != nil {
		return readerr
	}

	log.Printf("[DEBUG] Group Req Response JSON: %v", string(body))

	decodeerr := json.Unmarshal(body, &group)
	if decodeerr != nil {
		return decodeerr
	}

	log.Printf("[DEBUG] Group Req Response Decoded: %#v", group)

	d.SetId(string(fmt.Sprintf("%s/%s", workspace, group.Slug)))

	return resourceGroupsRead(d, m)
}

func resourceGroupsRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*Client)

	workspace, slug, err := groupId(d.Id())
	if err != nil {
		return err
	}
	groupsReq, _ := client.Get(fmt.Sprintf("1.0/groups/%s/", workspace))

	if groupsReq.StatusCode == 404 {
		log.Printf("[WARN] Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if groupsReq.Body == nil {
		return fmt.Errorf("error getting Group (%s): empty response", d.Id())
	}

	var groups []*UserGroup
	body, readerr := ioutil.ReadAll(groupsReq.Body)
	if readerr != nil {
		return readerr
	}

	log.Printf("[DEBUG] Groups Response JSON: %v", string(body))

	decodeerr := json.Unmarshal(body, &groups)
	if decodeerr != nil {
		return decodeerr
	}

	log.Printf("[DEBUG] Groups Response Decoded: %#v", groups)

	if len(groups) == 0 {
		return fmt.Errorf("error getting Group (%s): empty response", d.Id())
	}

	var group *UserGroup
	for _, grp := range groups {
		if grp.Slug == slug {
			group = grp
			break
		}
	}

	log.Printf("[DEBUG] Group Response Decoded: %#v", groups)

	d.Set("workspace", workspace)
	d.Set("slug", group.Slug)
	d.Set("name", group.Name)

	return nil
}

func resourceGroupsUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*Client)

	group := expandGroup(d)
	log.Printf("[DEBUG] Group Request: %#v", group)
	bytedata, err := json.Marshal(group)

	if err != nil {
		return err
	}

	_, err = client.Put(fmt.Sprintf("1.0/groups/%s/%s/",
		d.Get("workspace").(string), d.Get("slug").(string)), bytes.NewBuffer(bytedata))

	if err != nil {
		return err
	}

	return resourceGroupsRead(d, m)
}

func resourceGroupsDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*Client)

	workspace, slug, err := groupId(d.Id())
	if err != nil {
		return err
	}

	_, err = client.Delete(fmt.Sprintf("1.0/groups/%s/%s", workspace, slug))

	if err != nil {
		return err
	}

	return err
}

func expandGroup(d *schema.ResourceData) *UserGroup {
	key := &UserGroup{
		Name: d.Get("name").(string),
	}

	return key
}

func groupId(id string) (string, string, error) {
	parts := strings.Split(id, "/")

	if len(parts) != 2 {
		return "", "", fmt.Errorf("unexpected format of ID (%q), expected WORKSPACE-ID/GROUP-SLUG-ID", id)
	}

	return parts[0], parts[1], nil
}
