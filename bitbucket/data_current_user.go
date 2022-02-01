package bitbucket

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type PaginatedUserEmails struct {
	Values []UserEmail `json:"values,omitempty"`
	Page   int         `json:"page,omitempty"`
	Size   int         `json:"size,omitempty"`
	Next   string      `json:"next,omitempty"`
}

type UserEmail struct {
	Email       string `json:"email"`
	IsPrimary   bool   `json:"is_primary"`
	IsConfirmed bool   `json:"is_confirmed"`
}

func dataCurrentUser() *schema.Resource {
	return &schema.Resource{
		Read: dataReadCurrentUser,

		Schema: map[string]*schema.Schema{
			"username": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"display_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"uuid": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"nickname": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"account_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"is_staff": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"email": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"is_confirmed": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"is_primary": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"email": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataReadCurrentUser(d *schema.ResourceData, m interface{}) error {
	c := m.(*Client)

	curUser, err := c.Get("2.0/user")
	if err != nil {
		return err
	}

	if curUser.StatusCode == http.StatusNotFound {
		return fmt.Errorf("user not found")
	}

	if curUser.StatusCode >= http.StatusInternalServerError {
		return fmt.Errorf("internal server error fetching user")
	}

	curUserEmails, err := c.Get("2.0/user/emails")
	if err != nil {
		return err
	}

	body, readerr := ioutil.ReadAll(curUser.Body)
	if readerr != nil {
		return readerr
	}

	log.Printf("[DEBUG] Current User Response JSON: %v", string(body))

	emailBody, readerr := ioutil.ReadAll(curUserEmails.Body)
	if readerr != nil {
		return readerr
	}

	log.Printf("[DEBUG] Current User Emails Response JSON: %v", string(emailBody))

	var u apiUser

	decodeerr := json.Unmarshal(body, &u)
	if decodeerr != nil {
		return decodeerr
	}

	log.Printf("[DEBUG] Current User Response Decoded: %#v", u)

	var emails PaginatedUserEmails

	decodeerr = json.Unmarshal(emailBody, &emails)
	if decodeerr != nil {
		return decodeerr
	}

	log.Printf("[DEBUG] Current User Emails Response Decoded: %#v", emails)

	d.SetId(u.UUID)
	d.Set("uuid", u.UUID)
	d.Set("username", u.Username)
	d.Set("nickname", u.Nickname)
	d.Set("display_name", u.DisplayName)
	d.Set("account_id", u.AccountId)
	d.Set("account_status", u.AccountStatus)
	d.Set("is_staff", u.IsStaff)
	d.Set("email", flattenUserEmails(emails.Values))

	return nil
}

func flattenUserEmails(userEmails []UserEmail) []interface{} {
	if len(userEmails) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, btRaw := range userEmails {
		log.Printf("[DEBUG] User Email Response Decoded: %#v", btRaw)

		branchType := map[string]interface{}{
			"email":        btRaw.Email,
			"is_confirmed": btRaw.IsConfirmed,
			"is_primary":   btRaw.IsPrimary,
		}

		tfList = append(tfList, branchType)
	}

	return tfList
}
