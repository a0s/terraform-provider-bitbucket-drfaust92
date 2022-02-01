package bitbucket

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// Hook is the hook you want to add to a bitbucket repository
type Hook struct {
	UUID                 string   `json:"uuid,omitempty"`
	URL                  string   `json:"url,omitempty"`
	Description          string   `json:"description,omitempty"`
	Active               bool     `json:"active"`
	SkipCertVerification bool     `json:"skip_cert_verification"`
	Events               []string `json:"events,omitempty"`
}

func resourceHook() *schema.Resource {
	return &schema.Resource{
		Create: resourceHookCreate,
		Read:   resourceHookRead,
		Update: resourceHookUpdate,
		Delete: resourceHookDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), "/")
				if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
					return nil, fmt.Errorf("unexpected format of ID (%q), expected OWNER/REPO/HOOK-ID", d.Id())
				}
				d.SetId(idParts[2])
				d.Set("owner", idParts[0])
				d.Set("repository", idParts[1])
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"owner": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"repository": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"active": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"url": {
				Type:     schema.TypeString,
				Required: true,
			},
			"uuid": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Required: true,
			},
			"events": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{
						"pullrequest:unapproved",
						"issue:comment_created",
						"repo:imported",
						"repo:created",
						"repo:commit_comment_created",
						"pullrequest:approved",
						"pullrequest:comment_updated",
						"issue:updated",
						"project:updated",
						"repo:deleted",
						"pullrequest:changes_request_created",
						"pullrequest:comment_created",
						"repo:commit_status_updated",
						"pullrequest:updated",
						"issue:created",
						"repo:fork",
						"pullrequest:comment_deleted",
						"repo:commit_status_created",
						"repo:updated",
						"pullrequest:rejected",
						"pullrequest:fulfilled",
						"pullrequest:created",
						"pullrequest:changes_request_removed",
						"repo:transfer",
						"repo:push",
					}, false),
				},
			},
			"skip_cert_verification": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
		},
	}
}

func createHook(d *schema.ResourceData) *Hook {

	events := make([]string, 0, len(d.Get("events").(*schema.Set).List()))

	for _, item := range d.Get("events").(*schema.Set).List() {
		events = append(events, item.(string))
	}

	hook := &Hook{
		URL:                  d.Get("url").(string),
		Description:          d.Get("description").(string),
		Active:               d.Get("active").(bool),
		SkipCertVerification: d.Get("skip_cert_verification").(bool),
		Events:               events,
	}

	return hook
}

func resourceHookCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*Client)
	hook := createHook(d)

	payload, err := json.Marshal(hook)
	if err != nil {
		return err
	}

	hookReq, err := client.Post(fmt.Sprintf("2.0/repositories/%s/%s/hooks",
		d.Get("owner").(string),
		d.Get("repository").(string),
	), bytes.NewBuffer(payload))

	if err != nil {
		return err
	}

	body, readerr := ioutil.ReadAll(hookReq.Body)
	if readerr != nil {
		return readerr
	}

	decodeerr := json.Unmarshal(body, &hook)
	if decodeerr != nil {
		return decodeerr
	}

	d.SetId(hook.UUID)

	return resourceHookRead(d, m)
}
func resourceHookRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*Client)

	hookReq, err := client.Get(fmt.Sprintf("2.0/repositories/%s/%s/hooks/%s",
		d.Get("owner").(string),
		d.Get("repository").(string),
		url.PathEscape(d.Id()),
	))

	if hookReq.StatusCode == 404 {
		log.Printf("[WARN] Repository Hook (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return err
	}

	log.Printf("ID: %s", url.PathEscape(d.Id()))

	if hookReq.StatusCode == 200 {
		var hook Hook

		body, readerr := ioutil.ReadAll(hookReq.Body)
		if readerr != nil {
			return readerr
		}

		decodeerr := json.Unmarshal(body, &hook)
		if decodeerr != nil {
			return decodeerr
		}

		d.Set("uuid", hook.UUID)
		d.Set("description", hook.Description)
		d.Set("active", hook.Active)
		d.Set("url", hook.URL)
		d.Set("skip_cert_verification", hook.SkipCertVerification)
		d.Set("events", hook.Events)
	}

	return nil
}

func resourceHookUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*Client)
	hook := createHook(d)
	payload, err := json.Marshal(hook)
	if err != nil {
		return err
	}

	_, err = client.Put(fmt.Sprintf("2.0/repositories/%s/%s/hooks/%s",
		d.Get("owner").(string),
		d.Get("repository").(string),
		url.PathEscape(d.Id()),
	), bytes.NewBuffer(payload))

	if err != nil {
		return err
	}

	return resourceHookRead(d, m)
}

func resourceHookDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*Client)
	_, err := client.Delete(fmt.Sprintf("2.0/repositories/%s/%s/hooks/%s",
		d.Get("owner").(string),
		d.Get("repository").(string),
		url.PathEscape(d.Id()),
	))

	return err

}
