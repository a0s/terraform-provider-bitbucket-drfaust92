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

// sshKey is the data we need to send to create a new SSH Key for the repository
type SshKey struct {
	UUID    string `json:"uuid,omitempty"`
	Key     string `json:"key,omitempty"`
	Label   string `json:"label,omitempty"`
	Comment string `json:"comment,omitempty"`
}

func resourceSshKey() *schema.Resource {
	return &schema.Resource{
		Create: resourceSshKeysCreate,
		Read:   resourceSshKeysRead,
		Update: resourceSshKeysUpdate,
		Delete: resourceSshKeysDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"user": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"key": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"label": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"uuid": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"comment": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceSshKeysCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*Client)

	sshKey := expandsshKey(d)
	log.Printf("[DEBUG] SSH Key Request: %#v", sshKey)
	bytedata, err := json.Marshal(sshKey)

	if err != nil {
		return err
	}

	user := d.Get("user").(string)
	sshKeyReq, err := client.Post(fmt.Sprintf("2.0/users/%s/ssh-keys", user), bytes.NewBuffer(bytedata))

	if err != nil {
		return err
	}

	body, readerr := ioutil.ReadAll(sshKeyReq.Body)
	if readerr != nil {
		return readerr
	}

	decodeerr := json.Unmarshal(body, &sshKey)
	if decodeerr != nil {
		return decodeerr
	}

	d.SetId(string(fmt.Sprintf("%s/%s", user, sshKey.UUID)))

	return resourceSshKeysRead(d, m)
}

func resourceSshKeysRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*Client)

	user, keyId, err := sshKeyId(d.Id())
	if err != nil {
		return err
	}
	sshKeysReq, _ := client.Get(fmt.Sprintf("2.0/users/%s/ssh-keys/%s", user, keyId))

	if sshKeysReq.StatusCode == 404 {
		log.Printf("[WARN] SSH Key (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if sshKeysReq.Body == nil {
		return fmt.Errorf("error getting SSH Key (%s): empty response", d.Id())
	}

	var sshKey *SshKey
	body, readerr := ioutil.ReadAll(sshKeysReq.Body)
	if readerr != nil {
		return readerr
	}

	log.Printf("[DEBUG] SSH Key Response JSON: %v", string(body))

	decodeerr := json.Unmarshal(body, &sshKey)
	if decodeerr != nil {
		return decodeerr
	}

	log.Printf("[DEBUG] SSH Key Response Decoded: %#v", sshKey)

	d.Set("user", user)
	d.Set("key", d.Get("key").(string))
	d.Set("label", sshKey.Label)
	d.Set("uuid", sshKey.UUID)
	d.Set("comment", sshKey.Comment)

	return nil
}

func resourceSshKeysUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*Client)

	sshKey := expandsshKey(d)
	log.Printf("[DEBUG] SSH Key Request: %#v", sshKey)
	bytedata, err := json.Marshal(sshKey)

	if err != nil {
		return err
	}

	_, err = client.Put(fmt.Sprintf("2.0/users/%s/ssh-keys/%s",
		d.Get("user").(string), d.Get("uuid").(string)), bytes.NewBuffer(bytedata))

	if err != nil {
		return err
	}

	return resourceSshKeysRead(d, m)
}

func resourceSshKeysDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*Client)

	user, keyId, err := sshKeyId(d.Id())
	if err != nil {
		return err
	}

	_, err = client.Delete(fmt.Sprintf("2.0/users/%s/ssh-keys/%s", user, keyId))

	if err != nil {
		return err
	}

	return err
}

func expandsshKey(d *schema.ResourceData) *SshKey {
	key := &SshKey{
		Key:   d.Get("key").(string),
		Label: d.Get("label").(string),
	}

	return key
}

func sshKeyId(id string) (string, string, error) {
	parts := strings.Split(id, "/")

	if len(parts) != 2 {
		return "", "", fmt.Errorf("unexpected format of ID (%q), expected USER-ID/KEY-ID", id)
	}

	return parts[0], parts[1], nil
}
