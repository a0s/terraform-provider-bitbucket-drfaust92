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

// Project is the project data we need to send to create a project on the bitbucket api
type Project struct {
	Key                     string        `json:"key,omitempty"`
	IsPrivate               bool          `json:"is_private,omitempty"`
	Owner                   string        `json:"owner.username,omitempty"`
	Description             string        `json:"description,omitempty"`
	Name                    string        `json:"name,omitempty"`
	UUID                    string        `json:"uuid,omitempty"`
	HasPubliclyVisibleRepos bool          `json:"has_publicly_visible_repos,omitempty"`
	ProjectLinks            *ProjectLinks `json:"links,omitempty"`
}

type ProjectLinks struct {
	Avatar Link `json:"avatar,omitempty"`
}

type Link struct {
	Href string `json:"href,omitempty"`
	Name string `json:"name,omitempty"`
}

func resourceProject() *schema.Resource {
	return &schema.Resource{
		Create: resourceProjectCreate,
		Update: resourceProjectUpdate,
		Read:   resourceProjectRead,
		Delete: resourceProjectDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"key": {
				Type:     schema.TypeString,
				Required: true,
			},
			"is_private": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"owner": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"has_publicly_visible_repos": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"uuid": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"link": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"avatar": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"href": {
										Type:     schema.TypeString,
										Optional: true,
										DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
											return strings.HasPrefix(old, "https://bitbucket.org/account/user")
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func newProjectFromResource(d *schema.ResourceData) *Project {
	project := &Project{
		Name:        d.Get("name").(string),
		IsPrivate:   d.Get("is_private").(bool),
		Description: d.Get("description").(string),
		Key:         d.Get("key").(string),
	}

	if v, ok := d.GetOk("link"); ok && len(v.([]interface{})) > 0 && v.([]interface{}) != nil {
		project.ProjectLinks = expandProjectLinks(v.([]interface{}))
	}

	return project
}

func resourceProjectUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*Client)
	project := newProjectFromResource(d)

	var jsonbuffer []byte

	jsonpayload := bytes.NewBuffer(jsonbuffer)
	enc := json.NewEncoder(jsonpayload)
	enc.Encode(project)

	var projectKey string
	projectKey = d.Get("key").(string)
	if projectKey == "" {
		projectKey = d.Get("key").(string)
	}

	_, err := client.Put(fmt.Sprintf("2.0/workspaces/%s/projects/%s",
		d.Get("owner").(string),
		projectKey,
	), jsonpayload)

	if err != nil {
		return err
	}

	return resourceProjectRead(d, m)
}

func resourceProjectCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*Client)
	project := newProjectFromResource(d)

	bytedata, err := json.Marshal(project)

	if err != nil {
		return err
	}

	var projectKey string
	projectKey = d.Get("key").(string)
	if projectKey == "" {
		projectKey = d.Get("key").(string)
	}

	owner := d.Get("owner").(string)
	if owner == "" {
		return fmt.Errorf("owner must not be a empty string")
	}

	_, err = client.Post(fmt.Sprintf("2.0/workspaces/%s/projects/",
		d.Get("owner").(string),
	), bytes.NewBuffer(bytedata))

	if err != nil {
		return err
	}

	d.SetId(string(fmt.Sprintf("%s/%s", d.Get("owner").(string), projectKey)))

	return resourceProjectRead(d, m)
}

func resourceProjectRead(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id != "" {
		idparts := strings.Split(id, "/")
		if len(idparts) == 2 {
			d.Set("owner", idparts[0])
			d.Set("key", idparts[1])
		} else {
			return fmt.Errorf("incorrect ID format, should match `owner/key`")
		}
	}

	var projectKey string
	projectKey = d.Get("key").(string)
	if projectKey == "" {
		projectKey = d.Get("key").(string)
	}

	client := m.(*Client)
	projectReq, _ := client.Get(fmt.Sprintf("2.0/workspaces/%s/projects/%s",
		d.Get("owner").(string),
		projectKey,
	))

	if projectReq.StatusCode == 404 {
		log.Printf("[WARN] Project (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if projectReq.StatusCode == 200 {

		var project Project

		body, readerr := ioutil.ReadAll(projectReq.Body)
		if readerr != nil {
			return readerr
		}

		log.Printf("[DEBUG] Project Response JSON: %v", string(body))

		decodeerr := json.Unmarshal(body, &project)
		if decodeerr != nil {
			return decodeerr
		}

		log.Printf("[DEBUG] Project Response Decoded: %#v", project)

		d.Set("key", project.Key)
		d.Set("is_private", project.IsPrivate)
		d.Set("name", project.Name)
		d.Set("description", project.Description)
		d.Set("has_publicly_visible_repos", project.HasPubliclyVisibleRepos)
		d.Set("uuid", project.UUID)
		d.Set("link", flattenProjectLinks(project.ProjectLinks))
	}

	return nil
}

func resourceProjectDelete(d *schema.ResourceData, m interface{}) error {

	var projectKey string
	projectKey = d.Get("key").(string)
	if projectKey == "" {
		projectKey = d.Get("key").(string)
	}

	client := m.(*Client)
	_, err := client.Delete(fmt.Sprintf("2.0/workspaces/%s/projects/%s",
		d.Get("owner").(string),
		projectKey,
	))

	return err
}

func expandProjectLinks(l []interface{}) *ProjectLinks {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	rp := &ProjectLinks{}

	if v, ok := tfMap["avatar"].([]interface{}); ok && len(v) > 0 {
		rp.Avatar = expandProjectLink(v)
	}

	return rp
}

func expandProjectLink(l []interface{}) Link {

	tfMap, _ := l[0].(map[string]interface{})

	rp := Link{}

	if v, ok := tfMap["href"].(string); ok {
		rp.Href = v
	}

	return rp
}

func flattenProjectLinks(rp *ProjectLinks) []interface{} {
	if rp == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"avatar": flattenProjectLink(rp.Avatar),
	}

	return []interface{}{m}
}

func flattenProjectLink(rp Link) []interface{} {
	m := map[string]interface{}{
		"href": rp.Href,
	}

	return []interface{}{m}
}
