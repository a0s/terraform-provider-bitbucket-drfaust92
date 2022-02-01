---
layout: "bitbucket"
page_title: "Bitbucket: bitbucket_ssh_key"
sidebar_current: "docs-bitbucket-resource-hook"
description: |-
  Provides a Bitbucket SSH Key
---

# bitbucket\_ssh_key

Provides a Bitbucket SSH Key resource.

This allows you to manage your SSH Keys for a user.

## Example Usage

```hcl
resource "bitbucket_ssh_key" "test" {
  user  = data.bitbucket_current_user.test.uuid
  key   = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIKqP3Cr632C2dNhhgKVcon4ldUSAeKiku2yP9O9/bDtY"
  label = "test-key"
}
```

## Argument Reference

The following arguments are supported:

* `user` - (Required) This can either be the UUID of the account, surrounded by curly-braces, for example: {account UUID}, OR an Atlassian Account ID.
* `key` - (Required) The SSH public key value in OpenSSH format.
* `label` - (Optional) The user-defined label for the SSH key

## Attributes Reference

* `uuid` - The SSH key's UUID value.
* `comment` - The comment parsed from the SSH key (if present)

## Import

SSH Keys can be imported using their `user_id/key-id` ID, e.g.

```sh
terraform import bitbucket_ssh_key.key user-id/key-id
```
