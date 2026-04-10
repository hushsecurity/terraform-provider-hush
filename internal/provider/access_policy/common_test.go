package access_policy

import (
	"testing"

	"github.com/hushsecurity/terraform-provider-hush/internal/client"
)

func TestExpandVolumeDeliveryConfig(t *testing.T) {
	input := []any{
		map[string]any{
			"mount_point": "/etc/secrets",
			"item": []any{
				map[string]any{
					"path": "db_password",
					"key":  "password",
					"type": "key",
				},
				map[string]any{
					"path": "api_key",
					"key":  "secret",
					"type": "key",
				},
			},
		},
	}

	result := expandVolumeDeliveryConfig(input)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Type != client.DeliveryTypeVolume {
		t.Errorf("expected type %q, got %q", client.DeliveryTypeVolume, result.Type)
	}
	if result.MountPoint != "/etc/secrets" {
		t.Errorf("expected mount_point %q, got %q", "/etc/secrets", result.MountPoint)
	}
	if len(result.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(result.Items))
	}
	if result.Items[0].Path != "db_password" {
		t.Errorf("expected item[0].path %q, got %q", "db_password", result.Items[0].Path)
	}
	if result.Items[0].Key != "password" {
		t.Errorf("expected item[0].key %q, got %q", "password", result.Items[0].Key)
	}
	if result.Items[0].Type != client.DeliveryMappingTypeKey {
		t.Errorf("expected item[0].type %q, got %q", client.DeliveryMappingTypeKey, result.Items[0].Type)
	}
	if result.Items[1].Path != "api_key" {
		t.Errorf("expected item[1].path %q, got %q", "api_key", result.Items[1].Path)
	}
	if result.Items[1].Key != "secret" {
		t.Errorf("expected item[1].key %q, got %q", "secret", result.Items[1].Key)
	}
}

func TestExpandVolumeDeliveryConfig_template(t *testing.T) {
	input := []any{
		map[string]any{
			"mount_point": "/var/secrets",
			"item": []any{
				map[string]any{
					"path": "db_config.json",
					"key":  "postgresql://${username}:${password}@host:5432/db",
					"type": "template",
				},
			},
		},
	}

	result := expandVolumeDeliveryConfig(input)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Items[0].Type != client.DeliveryMappingTypeTemplate {
		t.Errorf("expected item[0].type %q, got %q", client.DeliveryMappingTypeTemplate, result.Items[0].Type)
	}
	if result.Items[0].Key != "postgresql://${username}:${password}@host:5432/db" {
		t.Errorf("expected template key preserved, got %q", result.Items[0].Key)
	}
}

func TestExpandVolumeDeliveryConfig_nil(t *testing.T) {
	result := expandVolumeDeliveryConfig([]any{})
	if result != nil {
		t.Errorf("expected nil result for empty input, got %+v", result)
	}

	result = expandVolumeDeliveryConfig([]any{nil})
	if result != nil {
		t.Errorf("expected nil result for nil element, got %+v", result)
	}
}

func TestFlattenVolumeDeliveryConfig(t *testing.T) {
	input := map[string]any{
		"type":        "volume",
		"mount_point": "/etc/secrets",
		"items": []any{
			map[string]any{
				"path": "db_password",
				"key":  "password",
				"type": "key",
			},
			map[string]any{
				"path": "api_key",
				"key":  "secret",
				"type": "key",
			},
		},
	}

	result := flattenVolumeDeliveryConfig(input)

	if len(result) != 1 {
		t.Fatalf("expected 1 config block, got %d", len(result))
	}

	configMap, ok := result[0].(map[string]any)
	if !ok {
		t.Fatal("expected result[0] to be map[string]any")
	}
	if configMap["mount_point"] != "/etc/secrets" {
		t.Errorf("expected mount_point %q, got %q", "/etc/secrets", configMap["mount_point"])
	}

	items, ok := configMap["item"].([]any)
	if !ok {
		t.Fatal("expected item to be []any")
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	item0, ok := items[0].(map[string]any)
	if !ok {
		t.Fatal("expected items[0] to be map[string]any")
	}
	if item0["path"] != "db_password" {
		t.Errorf("expected item[0].path %q, got %q", "db_password", item0["path"])
	}
	if item0["key"] != "password" {
		t.Errorf("expected item[0].key %q, got %q", "password", item0["key"])
	}
	if item0["type"] != "key" {
		t.Errorf("expected item[0].type %q, got %q", "key", item0["type"])
	}

	item1, ok := items[1].(map[string]any)
	if !ok {
		t.Fatal("expected items[1] to be map[string]any")
	}
	if item1["path"] != "api_key" {
		t.Errorf("expected item[1].path %q, got %q", "api_key", item1["path"])
	}
}

func TestExpandEnvDeliveryConfig(t *testing.T) {
	input := []any{
		map[string]any{
			"name": "PORT",
			"key":  "port",
			"type": "key",
		},
		map[string]any{
			"name": "DATABASE_URL",
			"key":  "postgresql://${username}:${password}@host:5432/db",
			"type": "template",
		},
	}

	result := expandEnvDeliveryConfig(input)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Type != client.DeliveryTypeEnv {
		t.Errorf("expected type %q, got %q", client.DeliveryTypeEnv, result.Type)
	}
	if len(result.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(result.Items))
	}
	if result.Items[0].Name != "PORT" {
		t.Errorf("expected item[0].name %q, got %q", "PORT", result.Items[0].Name)
	}
	if result.Items[0].Key != "port" {
		t.Errorf("expected item[0].key %q, got %q", "port", result.Items[0].Key)
	}
	if result.Items[0].Type != client.DeliveryMappingTypeKey {
		t.Errorf("expected item[0].type %q, got %q", client.DeliveryMappingTypeKey, result.Items[0].Type)
	}
	if result.Items[1].Name != "DATABASE_URL" {
		t.Errorf("expected item[1].name %q, got %q", "DATABASE_URL", result.Items[1].Name)
	}
	if result.Items[1].Type != client.DeliveryMappingTypeTemplate {
		t.Errorf("expected item[1].type %q, got %q", client.DeliveryMappingTypeTemplate, result.Items[1].Type)
	}
}

func TestExpandEnvDeliveryConfig_nil(t *testing.T) {
	result := expandEnvDeliveryConfig([]any{})
	if result != nil {
		t.Errorf("expected nil result for empty input, got %+v", result)
	}

	result = expandEnvDeliveryConfig([]any{nil})
	if result != nil {
		t.Errorf("expected nil result for nil element, got %+v", result)
	}
}

func TestExpandAwsWifDeliveryConfig(t *testing.T) {
	input := []any{
		map[string]any{
			"role_arn":     "arn:aws:iam::123456789012:role/test-role",
			"subject_kind": "hush_subject",
			"subject":      "my-subject",
		},
	}

	result := expandAwsWifDeliveryConfig(input)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Type != client.DeliveryTypeAwsWif {
		t.Errorf("expected type %q, got %q", client.DeliveryTypeAwsWif, result.Type)
	}
	if result.RoleArn != "arn:aws:iam::123456789012:role/test-role" {
		t.Errorf("expected role_arn %q, got %q", "arn:aws:iam::123456789012:role/test-role", result.RoleArn)
	}
	if result.SubjectKind != client.WifSubjectKindHushSubject {
		t.Errorf("expected subject_kind %q, got %q", client.WifSubjectKindHushSubject, result.SubjectKind)
	}
	if result.Subject != "my-subject" {
		t.Errorf("expected subject %q, got %q", "my-subject", result.Subject)
	}
}

func TestExpandAwsWifDeliveryConfig_serviceAccount(t *testing.T) {
	input := []any{
		map[string]any{
			"role_arn":     "arn:aws:iam::123456789012:role/sa-role",
			"subject_kind": "service_account",
			"subject":      "",
		},
	}

	result := expandAwsWifDeliveryConfig(input)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.SubjectKind != client.WifSubjectKindServiceAccount {
		t.Errorf("expected subject_kind %q, got %q", client.WifSubjectKindServiceAccount, result.SubjectKind)
	}
	if result.Subject != "" {
		t.Errorf("expected empty subject, got %q", result.Subject)
	}
}

func TestExpandAwsWifDeliveryConfig_nil(t *testing.T) {
	result := expandAwsWifDeliveryConfig([]any{})
	if result != nil {
		t.Errorf("expected nil result for empty input, got %+v", result)
	}

	result = expandAwsWifDeliveryConfig([]any{nil})
	if result != nil {
		t.Errorf("expected nil result for nil element, got %+v", result)
	}
}

func TestFlattenAwsWifDeliveryConfig(t *testing.T) {
	input := map[string]any{
		"type":         "aws_wif",
		"role_arn":     "arn:aws:iam::123456789012:role/test-role",
		"subject_kind": "hush_subject",
		"subject":      "my-subject",
	}

	result := flattenAwsWifDeliveryConfig(input)

	if len(result) != 1 {
		t.Fatalf("expected 1 config block, got %d", len(result))
	}

	configMap, ok := result[0].(map[string]any)
	if !ok {
		t.Fatal("expected result[0] to be map[string]any")
	}
	if configMap["role_arn"] != "arn:aws:iam::123456789012:role/test-role" {
		t.Errorf("expected role_arn %q, got %q", "arn:aws:iam::123456789012:role/test-role", configMap["role_arn"])
	}
	if configMap["subject_kind"] != "hush_subject" {
		t.Errorf("expected subject_kind %q, got %q", "hush_subject", configMap["subject_kind"])
	}
	if configMap["subject"] != "my-subject" {
		t.Errorf("expected subject %q, got %q", "my-subject", configMap["subject"])
	}
}

func TestFlattenAwsWifDeliveryConfig_serviceAccount(t *testing.T) {
	input := map[string]any{
		"type":         "aws_wif",
		"role_arn":     "arn:aws:iam::123456789012:role/sa-role",
		"subject_kind": "service_account",
	}

	result := flattenAwsWifDeliveryConfig(input)

	if len(result) != 1 {
		t.Fatalf("expected 1 config block, got %d", len(result))
	}

	configMap, ok := result[0].(map[string]any)
	if !ok {
		t.Fatal("expected result[0] to be map[string]any")
	}
	if configMap["role_arn"] != "arn:aws:iam::123456789012:role/sa-role" {
		t.Errorf("expected role_arn %q, got %q", "arn:aws:iam::123456789012:role/sa-role", configMap["role_arn"])
	}
	if _, hasSubject := configMap["subject"]; hasSubject {
		t.Errorf("expected no subject key for service_account, got %q", configMap["subject"])
	}
}
