#!/usr/bin/env bash
# Type-checks every example under examples/ against the locally built provider.
#
# tfplugindocs embeds these files verbatim into docs/, which ships to the
# Terraform registry, so an example that does not validate is a broken snippet
# handed to customers. `terraform fmt` only checks layout; this checks schemas.
#
# Examples reference each other (most credentials point at
# hush_deployment.example.id), so they are assembled into one module rather than
# validated a directory at a time. The flip side: two example files declaring
# the same resource address (say, a second `resource "hush_deployment"
# "example"`) fail here as a duplicate, so give each example's resources names
# that are unique across examples/.
#
# What it does NOT check:
#   1. Schema only. Cross-field rules the API enforces (a template naming a key
#      the credential lacks, a field the provider marks optional and the API
#      requires) surface at apply and need acceptance tests.
#   2. A value from a variable is unknown here, and SDKv2 skips ValidateFunc on
#      unknown values, so `project_id = "proj_abc"` is checked but
#      `project_id = var.foo` is not. Prefer literals for anything non-secret.
#   3. Examples reaching outside the hush provider are skipped entirely -- see
#      SKIPPED below. Validating them would need `terraform init`, which pulls
#      ~700MB of third-party providers, breaks behind a firewall, and lets an
#      unrelated upstream release fail CI. Their module sources, version
#      constraints and output names are checked by nothing; the run prints them
#      every time so the gap stays visible.
#
# The skip and the missing init depend on each other. `terraform validate`
# normally wants an initialised directory; it works here without one only
# because the assembled module has no module blocks, no backend, and a single
# provider that dev_overrides resolves. Widen the skip heuristic and init stays
# unnecessary; narrow it and validate starts failing on missing providers.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

if [[ ! -x bin/terraform-provider-hush ]]; then
  echo "bin/terraform-provider-hush not found -- run 'make build' first." >&2
  exit 1
fi

WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

SKIPPED=()

# Anything needing a provider we can't resolve offline: an external module, or a
# resource/data block belonging to another provider.
needs_other_providers() {
  awk '
    /^module /           { hit = 1; exit }
    /^(resource|data) "/ { if ($2 !~ /^"hush_/) { hit = 1; exit } }
    END                  { exit(hit ? 0 : 1) }
  ' "$1"
}

add_example() {
  local file="$1" prefix="$2"

  if needs_other_providers "$file"; then
    SKIPPED+=("$file")
    return
  fi

  # Prefix output names so identically named outputs don't collide.
  sed -E "s/^output \"([^\"]+)\"/output \"${prefix}_\\1\"/" "$file" > "$WORK/${prefix}.tf"
}

for kind in resources data-sources; do
  [[ -d "examples/$kind" ]] || continue
  for dir in "examples/$kind"/*/; do
    for file in "$dir"*.tf; do
      [[ -f "$file" ]] || continue
      add_example "$file" "${kind:0:1}_$(basename "$dir")_$(basename "$file" .tf)"
    done
  done
done

# Every example directory must be one we know how to handle; a new kind (say
# ephemeral-resources or functions) must be added to the loop above, not
# silently skipped.
for dir in examples/*/; do
  case "$(basename "$dir")" in
    resources|data-sources|provider) ;;
    *) echo "examples/$(basename "$dir") is not covered by this script; add it to the loop." >&2; exit 1 ;;
  esac
done

# docs/index.md embeds this one the same way, so it ships too. It is also the
# only source of required_providers, so it must not be skipped: without it every
# hush_* resource would resolve to the implicit hashicorp/hush and miss the
# override.
add_example examples/provider/provider.tf "p_provider"
if [[ ! -f "$WORK/p_provider.tf" ]]; then
  echo "examples/provider/provider.tf was skipped; it must declare only the hush provider." >&2
  exit 1
fi

# Stub every variable the examples reference. `|| true` because grep exits 1 on
# no match, which pipefail would turn into a silent abort.
{
  grep -rhoE 'var\.[a-zA-Z_][a-zA-Z0-9_]*' "$WORK" | sort -u | sed 's/var\.//' || true
} | while read -r var; do
  [[ -n "$var" ]] || continue
  printf 'variable "%s" {\n  type    = string\n  default = "placeholder"\n}\n\n' "$var"
done > "$WORK/_variables.tf"

# dev_overrides makes `terraform init` unnecessary for the provider itself. No
# other installation method is listed because nothing else may be installed.
cat > "$WORK/.terraformrc" <<EOF
provider_installation {
  dev_overrides {
    "hushsecurity/hush" = "$REPO_ROOT/bin"
  }
}
EOF

export TF_CLI_CONFIG_FILE="$WORK/.terraformrc"
export TF_IN_AUTOMATION=1

if (( ${#SKIPPED[@]} )); then
  echo "Not validated (reach outside the hush provider, see limit 3):"
  printf '  %s\n' "${SKIPPED[@]}"
fi

if ! terraform -chdir="$WORK" validate -no-color; then
  echo >&2
  echo "Failures are named after their source: p_provider.tf is" >&2
  echo "examples/provider/provider.tf, and <kind>_<dir>_<file>.tf is" >&2
  echo "examples/<kind>/<dir>/<file>.tf." >&2
  exit 1
fi
