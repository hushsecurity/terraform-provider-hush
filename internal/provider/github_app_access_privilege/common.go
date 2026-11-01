package github_app_access_privilege

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	idDesc           = "The unique identifier of the GitHub App access privilege"
	nameDesc         = "The name of the GitHub App access privilege"
	descriptionDesc  = "The description of the GitHub App access privilege"
	repositoriesDesc = "The repositories tokens are narrowed to, by bare name (`my-repo`, not `owner/my-repo`): each 1-100 letters, digits, `.`, `-` or `_`, not `.` or `..`, listed once, at most 500. Empty means every repository the installation can access"
	permissionsDesc  = "The GitHub App permissions tokens are narrowed to, as a non-empty map of permission name to level (`read`, `write` or `admin`), e.g. `{ contents = \"read\" }`. Each must be granted to the App's installation"
	typeDesc         = "The type of access privilege"

	maxRepositories = 500
)

var (
	repositoryRe     = regexp.MustCompile(`^[A-Za-z0-9._-]{1,100}$`)
	permissionNameRe = regexp.MustCompile(`^[a-z][a-z_]*$`)
	permissionLevels = []string{"read", "write", "admin"}
)

func validateRepository(v any, path cty.Path) diag.Diagnostics {
	name, _ := v.(string)
	if strings.Contains(name, "/") {
		return diag.Errorf(
			"repository %q: use the bare repository name, without the owner", name)
	}
	if name == "." || name == ".." {
		return diag.Errorf("repository %q: not a repository name", name)
	}
	if !repositoryRe.MatchString(name) {
		return diag.Errorf(
			"repository %q: must be 1-100 letters, digits, '.', '-' or '_'", name)
	}
	return nil
}

// customizeDiff checks what the per-attribute validators cannot: an empty
// permissions map and duplicate repositories (SDKv2 runs no ValidateFunc on a
// TypeList itself), plus the permission entries, so that an entry whose value
// is unknown at plan time is skipped instead of rejected.
func customizeDiff(_ context.Context, d *schema.ResourceDiff, _ any) error {
	config := d.GetRawConfig()
	if config.IsNull() || !config.IsKnown() {
		return nil
	}
	if err := checkPermissions(config.GetAttr("permissions")); err != nil {
		return err
	}
	return checkRepositories(config.GetAttr("repositories"))
}

// checkPermissions mirrors the Hush API: at least one permission, each a
// permission name with a level of read, write or admin. The API checks the
// names against GitHub's permission list and the level each one accepts.
func checkPermissions(perms cty.Value) error {
	if perms.IsNull() || !perms.IsKnown() {
		return nil
	}
	if perms.LengthInt() == 0 {
		return fmt.Errorf("permissions: name at least one permission")
	}
	var errs []error
	for it := perms.ElementIterator(); it.Next(); {
		k, v := it.Element()
		name := k.AsString()
		if !permissionNameRe.MatchString(name) {
			errs = append(errs, fmt.Errorf(
				"permission %q: not a GitHub App permission name", name))
		}
		if v.IsNull() || !v.IsKnown() {
			continue
		}
		if l := v.AsString(); !slices.Contains(permissionLevels, l) {
			errs = append(errs, fmt.Errorf(
				"permission %q: level must be one of %s, not %q",
				name, strings.Join(permissionLevels, ", "), l))
		}
	}
	return errors.Join(errs...)
}

// checkRepositories rejects a repository listed twice, as the Hush API does.
// Unknown elements are skipped.
func checkRepositories(repos cty.Value) error {
	if repos.IsNull() || !repos.IsKnown() {
		return nil
	}
	seen := make(map[string]bool)
	for it := repos.ElementIterator(); it.Next(); {
		_, v := it.Element()
		if v.IsNull() || !v.IsKnown() {
			continue
		}
		name := v.AsString()
		if seen[name] {
			return fmt.Errorf("repository %q: listed more than once", name)
		}
		seen[name] = true
	}
	return nil
}

func ResourceSchema() map[string]*schema.Schema {
	s := DataSourceSchema()

	s["id"] = &schema.Schema{
		Description: idDesc,
		Type:        schema.TypeString,
		Computed:    true,
	}
	s["name"] = &schema.Schema{
		Description:  nameDesc,
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.StringLenBetween(1, 255),
	}
	s["description"] = &schema.Schema{
		Description:  descriptionDesc,
		Type:         schema.TypeString,
		Optional:     true,
		ValidateFunc: validation.StringLenBetween(0, 1000),
	}
	s["repositories"] = &schema.Schema{
		Description: repositoriesDesc,
		Type:        schema.TypeList,
		Optional:    true,
		MaxItems:    maxRepositories,
		Elem: &schema.Schema{
			Type:             schema.TypeString,
			ValidateDiagFunc: validateRepository,
		},
	}
	s["permissions"] = &schema.Schema{
		Description: permissionsDesc,
		Type:        schema.TypeMap,
		Required:    true,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	}

	return s
}

func DataSourceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"id": {
			Description: idDesc,
			Type:        schema.TypeString,
			Required:    true,
		},
		"name": {
			Description: nameDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"description": {
			Description: descriptionDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"repositories": {
			Description: repositoriesDesc,
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"permissions": {
			Description: permissionsDesc,
			Type:        schema.TypeMap,
			Computed:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"type": {
			Description: typeDesc,
			Type:        schema.TypeString,
			Computed:    true,
		},
	}
}
