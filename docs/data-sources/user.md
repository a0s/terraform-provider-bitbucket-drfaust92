---
layout: "bitbucket"
page_title: "Bitbucket: bitbucket_user"
sidebar_current: "docs-bitbucket-data-user"
description: |-
  Provides a data for a Bitbucket user
---

# bitbucket\_user

Provdes a way to fetch data on a user.

## Example Usage

```hcl
# Manage your repository
data "bitbucket_user" "reviewer" {
  username = "gob"
}
```

## Argument Reference

The following arguments are supported:

* `username` - (Required) the username of the user to query.

## Exports

* `uuid` - the uuid that bitbucket users to connect a user to various objects
* `display_name` - the display name that the user wants to use for GDPR
* `nickname` - Account name defined by the owner. Note that "nickname" cannot be used in place of "username" in URLs and queries, as "nickname" is not guaranteed to be unique.
* `account_status` - The status of the account.
* `account_id` - The user's Atlassian account ID.
* `is_staff` - is staff user.
