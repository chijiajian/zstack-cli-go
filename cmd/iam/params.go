package iam

import (
	"bytes"
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zstackio/zstack-sdk-go-v2/pkg/param"
)

func text(cmd *cobra.Command, name string) string {
	value, _ := cmd.Flags().GetString(name)
	return value
}
func optional(cmd *cobra.Command, name string) *string {
	if !cmd.Flags().Changed(name) {
		return nil
	}
	value := text(cmd, name)
	return &value
}
func nonempty(value, label string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s must be nonempty", label)
	}
	return nil
}
func password(cmd *cobra.Command, required, preview bool) (*string, error) {
	file := text(cmd, "password-file")
	if file == "" {
		if required {
			return nil, fmt.Errorf("--password-file is required")
		}
		return nil, nil
	}
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("cannot read password file")
	}
	value := strings.TrimSuffix(strings.TrimSuffix(string(data), "\n"), "\r")
	if value == "" {
		return nil, fmt.Errorf("password file is empty")
	}
	digest := fmt.Sprintf("%x", sha512.Sum512([]byte(value)))
	if preview {
		digest = "<redacted>"
	}
	return &digest, nil
}
func statements(cmd *cobra.Command, required bool) ([]param.PolicyStatementParam, error) {
	file := text(cmd, "statements-file")
	if file == "" {
		if required {
			return nil, fmt.Errorf("--statements-file is required")
		}
		return nil, nil
	}
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("cannot read statements file")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var values []param.PolicyStatementParam
	if err := decoder.Decode(&values); err != nil {
		return nil, fmt.Errorf("statements file must be a JSON array of policy statements")
	}
	if decoder.Decode(new(any)) != io.EOF {
		return nil, fmt.Errorf("unexpected trailing statements data")
	}
	if len(values) == 0 {
		return nil, fmt.Errorf("at least one policy statement is required")
	}
	for _, value := range values {
		if value.Effect != "Allow" && value.Effect != "Deny" {
			return nil, fmt.Errorf("statement effect must be Allow or Deny")
		}
		if len(value.Actions) == 0 {
			return nil, fmt.Errorf("statement actions must be nonempty")
		}
		for _, action := range value.Actions {
			if strings.TrimSpace(action) == "" {
				return nil, fmt.Errorf("statement actions must be nonempty")
			}
		}
	}
	return values, nil
}

// Build SDK parameter types at the CLI boundary; omission preserves server defaults.
func build(cmd *cobra.Command, kind, action, name string, preview bool) (any, error) {
	description := optional(cmd, "description")
	if action == "create" {
		if err := nonempty(name, "name"); err != nil {
			return nil, err
		}
		resource := optional(cmd, "resource-uuid")
		switch kind {
		case "account":
			secret, err := password(cmd, true, preview)
			if err != nil {
				return nil, err
			}
			return param.CreateAccountParam{Params: param.CreateAccountParamDetail{Name: name, Password: *secret, Description: description, Type: optional(cmd, "type"), ResourceUuid: resource}}, nil
		case "user":
			secret, err := password(cmd, true, preview)
			if err != nil {
				return nil, err
			}
			return param.CreateUserParam{Params: param.CreateUserParamDetail{Name: name, Password: *secret, Description: description, ResourceUuid: resource}}, nil
		case "policy":
			values, err := statements(cmd, true)
			if err != nil {
				return nil, err
			}
			return param.CreatePolicyParam{Params: param.CreatePolicyParamDetail{Name: name, Description: description, Statements: values, ResourceUuid: resource}}, nil
		case "role":
			values, err := statements(cmd, false)
			if err != nil {
				return nil, err
			}
			policies, _ := cmd.Flags().GetStringSlice("policy")
			return param.CreateRoleParam{Params: param.CreateRoleParamDetail{Name: name, Description: description, Statements: values, PolicyUuids: policies, Identity: optional(cmd, "identity"), ResourceUuid: resource}}, nil
		case "project":
			return param.CreateIAM2ProjectParam{Params: param.CreateIAM2ProjectParamDetail{Name: name, Description: description, OrganizationUuid: optional(cmd, "organization"), LinkAccountUuid: optional(cmd, "link-account"), ResourceUuid: resource}}, nil
		case "organization":
			if err := nonempty(text(cmd, "type"), "--type"); err != nil {
				return nil, err
			}
			return param.CreateIAM2OrganizationParam{Params: param.CreateIAM2OrganizationParamDetail{Name: name, Description: description, Type: text(cmd, "type"), ParentUuid: optional(cmd, "parent"), ResourceUuid: resource}}, nil
		}
	}
	newName := text(cmd, "name")
	if cmd.Flags().Changed("name") {
		if err := nonempty(newName, "--name"); err != nil {
			return nil, err
		}
	}
	changed := false
	for _, field := range []string{"name", "description", "password-file", "statements-file", "policy", "parent", "type"} {
		if cmd.Flags().Changed(field) {
			changed = true
		}
	}
	if !changed {
		return nil, fmt.Errorf("specify at least one field to update")
	}
	switch kind {
	case "account":
		secret, err := password(cmd, false, preview)
		if err != nil {
			return nil, err
		}
		return param.UpdateAccountParam{Params: param.UpdateAccountParamDetail{Name: newName, Description: description, Password: secret}}, nil
	case "user":
		secret, err := password(cmd, false, preview)
		if err != nil {
			return nil, err
		}
		return param.UpdateUserParam{Params: param.UpdateUserParamDetail{Name: newName, Description: description, Password: secret}}, nil
	case "role":
		values, err := statements(cmd, false)
		if err != nil {
			return nil, err
		}
		policies, _ := cmd.Flags().GetStringSlice("policy")
		if cmd.Flags().Changed("policy") && len(policies) == 0 {
			return nil, fmt.Errorf("--policy cannot be empty")
		}
		return param.UpdateRoleParam{Params: param.UpdateRoleParamDetail{Name: newName, Description: description, Statements: values, PolicyUuids: policies}}, nil
	case "project":
		return param.UpdateIAM2ProjectParam{Params: param.UpdateIAM2ProjectParamDetail{Name: newName, Description: description}}, nil
	case "organization":
		return param.UpdateIAM2OrganizationParam{Params: param.UpdateIAM2OrganizationParamDetail{Name: newName, Description: description, Type: optional(cmd, "type"), ParentUuid: optional(cmd, "parent")}}, nil
	}
	return nil, fmt.Errorf("unsupported IAM operation")
}
