---
layout: "bitbucket"
page_title: "Bitbucket: bitbucket_deployment_variable"
sidebar_current: "docs-bitbucket-resource-deployment-variable"
description: |-
  Manage variables for your pipelines deployment environments
---


# bitbucket\_deployment\_variable

This resource allows you to configure deployment variables.

## Example Usage

```hcl
resource "bitbucket_repository" "example" {
    owner = "workspace"
    name = "example-repository"
    pipelines_enabled = true
}
resource "bitbucket_deployment" "example" {
  repository = bitbucket_repository.example.id
  name = "test"
  stage = "Test"
}
resource "bitbucket_deployment_variable" "example" {
  deployment = bitbucket_deployment.test.id
  name = "EXAMPLE_VARIABLE"
  value = "thisisanexample"
  secured = false
}
```

## Argument Reference

* `deployment` - (Required) The deployment ID you want to assign this variable to.
* `name` - (Required) The name of the variable
* `value` - (Required) The stage (Test, Staging, Production)
* `secured` - (Optional) Boolean indicating whether the variable contains sensitive data
* `always_override` - (Optional) If true secured variables will always be overridden, if false remote changes will be ignored but changed from terraform will still be applied
* `uuid` - (Computed) The UUID of the variable
