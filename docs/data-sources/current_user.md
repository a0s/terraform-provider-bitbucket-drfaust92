---
layout: "bitbucket"
page_title: "Bitbucket: bitbucket_user"
sidebar_current: "docs-bitbucket-data-user"
description: |-
  Provides a data for a Bitbucket user
---

# bitbucket\_user

Provdes a way to fetch data of the current user.

## Example Usage

```hcl
# Manage your repository
data "bitbucket_currnet_user" "example" {}
```

## Argument Reference

There are no arguments available for this data source.

## Attributes Reference

* `username` - The Username.
* `uuid` - the uuid that bitbucket users to connect a user to various objects
* `display_name` - the display name that the user wants to use for GDPR
* `nickname` - Account name defined by the owner. Note that "nickname" cannot be used in place of "username" in URLs and queries, as "nickname" is not guaranteed to be unique.
* `account_status` - The status of the account.
* `account_id` - The user's Atlassian account ID.
* `is_staff` - is staff user.
* `email` - A Set of emails associated to current user. See [Email](#email) below.

### Email

* `is_primary` - Whether is primary email for the user.
* `is_confirmed` - Whether the email is confirmed.
* `email` - The email address.
