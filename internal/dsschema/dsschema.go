// Package dsschema derives a data source's schema from the resource it reads.
//
// A data source reports what the resource manages, so its attributes are the
// resource's with every one of them computed: nothing is set, validated,
// defaulted or constrained, since nothing is configured. Writing that schema
// out a second time by hand only lets the two drift apart.
package dsschema

import (
	"fmt"
	"sort"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Options tailors the derived schema.
type Options struct {
	// Drop names the attributes to leave out -- secrets, which the API never
	// returns, and their write-only companions. A nested attribute is named by
	// its path through the blocks, "auth.secret".
	Drop []string
	// Descriptions replaces a description, by the same kind of path, where
	// the resource's speaks of configuring the attribute: "Removing it..." or
	// "Fixed at creation" says nothing true of a value that is only read.
	Descriptions map[string]string
}

// FromResource returns a computed copy of a resource's schema. A path in
// Drop or Descriptions that names no attribute panics: it is a mistake made
// when the provider is written, and a dropped secret that quietly stays would
// be the worst way to find it.
func FromResource(resource map[string]*schema.Schema, opts Options) map[string]*schema.Schema {
	unused := map[string]bool{}
	for _, path := range opts.Drop {
		unused[path] = true
	}
	for path := range opts.Descriptions {
		unused[path] = true
	}
	out := computed(resource, "", opts, unused)
	if len(unused) > 0 {
		paths := make([]string, 0, len(unused))
		for path := range unused {
			paths = append(paths, path)
		}
		sort.Strings(paths)
		panic(fmt.Sprintf("dsschema: no attribute at %v", paths))
	}
	return out
}

func computed(in map[string]*schema.Schema, prefix string, opts Options, unused map[string]bool) map[string]*schema.Schema {
	out := make(map[string]*schema.Schema, len(in))
	for name, s := range in {
		path := prefix + name
		if contains(opts.Drop, path) {
			delete(unused, path)
			continue
		}
		c := &schema.Schema{
			Type:        s.Type,
			Description: s.Description,
			Computed:    true,
			Sensitive:   s.Sensitive,
			Set:         s.Set,
		}
		if description, ok := opts.Descriptions[path]; ok {
			c.Description = description
			delete(unused, path)
		}
		switch elem := s.Elem.(type) {
		case *schema.Resource:
			c.Elem = &schema.Resource{Schema: computed(elem.Schema, path+".", opts, unused)}
		case *schema.Schema:
			c.Elem = element(elem)
		}
		out[name] = c
	}
	return out
}

// element copies a collection's element type, and that of any collection it
// holds in turn, without what only a configured value needs.
func element(s *schema.Schema) *schema.Schema {
	c := &schema.Schema{Type: s.Type}
	if inner, ok := s.Elem.(*schema.Schema); ok {
		c.Elem = element(inner)
	}
	return c
}

func contains(paths []string, path string) bool {
	for _, p := range paths {
		if p == path {
			return true
		}
	}
	return false
}

// Lookup turns an attribute of a derived schema into what the configuration
// names the object by: optional, and still computed when the object is found
// by something else.
func Lookup(s map[string]*schema.Schema, name, description string, conflictsWith ...string) {
	attr := s[name]
	attr.Optional = true
	attr.Computed = true
	if description != "" {
		attr.Description = description
	}
	attr.ConflictsWith = conflictsWith
}
