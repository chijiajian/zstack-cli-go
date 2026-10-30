// Copyright 2025 zstack.io
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package get

import (
	"fmt"
	"github.com/zstackio/zstack-sdk-go-v2/pkg/view"
	"time"

	"github.com/chijiajian/zstack-cli-go/pkg/common"
	"github.com/spf13/cobra"
)

type FormattedAccount struct {
	Name        string `json:"name" yaml:"name" header:"NAME"`
	UUID        string `json:"uuid" yaml:"uuid" header:"UUID"`
	Description string `json:"description" yaml:"description" header:"DESCRIPTION"`
	Type        string `json:"type" yaml:"type" header:"TYPE"`
	CreateDate  string `json:"createDate" yaml:"createDate" header:"CREATE DATE"`
}

type FormattedUser struct {
	Name        string `json:"name" yaml:"name" header:"NAME"`
	UUID        string `json:"uuid" yaml:"uuid" header:"UUID"`
	AccountUUID string `json:"accountUuid" yaml:"accountUuid" header:"ACCOUNT UUID"`
	Description string `json:"description" yaml:"description" header:"DESCRIPTION"`
	CreateDate  string `json:"createDate" yaml:"createDate" header:"CREATE DATE"`
}

type FormattedIAM2Project struct {
	Name              string `json:"name" yaml:"name" header:"NAME"`
	UUID              string `json:"uuid" yaml:"uuid" header:"UUID"`
	Description       string `json:"description" yaml:"description" header:"DESCRIPTION"`
	State             string `json:"state" yaml:"state" header:"STATE"`
	LinkedAccountUUID string `json:"linkedAccountUuid" yaml:"linkedAccountUuid" header:"LINKED ACCOUNT"`
	CreateDate        string `json:"createDate" yaml:"createDate" header:"CREATE DATE"`
}

type FormattedIAM2Organization struct {
	Name                 string `json:"name" yaml:"name" header:"NAME"`
	UUID                 string `json:"uuid" yaml:"uuid" header:"UUID"`
	Description          string `json:"description" yaml:"description" header:"DESCRIPTION"`
	State                string `json:"state" yaml:"state" header:"STATE"`
	Type                 string `json:"type" yaml:"type" header:"TYPE"`
	ParentUUID           string `json:"parentUuid" yaml:"parentUuid" header:"PARENT UUID"`
	RootOrganizationUUID string `json:"rootOrganizationUuid" yaml:"rootOrganizationUuid" header:"ROOT ORG"`
	CreateDate           string `json:"createDate" yaml:"createDate" header:"CREATE DATE"`
}

type FormattedRole struct {
	Name        string `json:"name" yaml:"name" header:"NAME"`
	UUID        string `json:"uuid" yaml:"uuid" header:"UUID"`
	Description string `json:"description" yaml:"description" header:"DESCRIPTION"`
	Identity    string `json:"identity" yaml:"identity" header:"IDENTITY"`
	Type        string `json:"type" yaml:"type" header:"TYPE"`
	State       string `json:"state" yaml:"state" header:"STATE"`
	CreateDate  string `json:"createDate" yaml:"createDate" header:"CREATE DATE"`
}

type FormattedPolicy struct {
	Name           string `json:"name" yaml:"name" header:"NAME"`
	UUID           string `json:"uuid" yaml:"uuid" header:"UUID"`
	AccountUUID    string `json:"accountUuid" yaml:"accountUuid" header:"ACCOUNT UUID"`
	StatementCount int    `json:"statementCount" yaml:"statementCount" header:"STATEMENTS"`
	CreateDate     string `json:"createDate" yaml:"createDate" header:"CREATE DATE"`
}

var accountsCmd = &cobra.Command{
	Use:   "accounts [name]",
	Short: "List accounts",
	Long:  `List ZStack accounts.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cobraCmd *cobra.Command, args []string) error {

		queryParam, err := common.BuildQueryParams(cobraCmd, args, "name")
		if err != nil {
			return fmt.Errorf("Error building query parameters: %w", err)
		}

		result, err := common.Fetch[view.AccountInventoryView](cobraCmd, queryParam, "v1/accounts")
		accounts := result.Items
		if err != nil {
			return fmt.Errorf("Error querying accounts: %w", err)
		}

		formatted := make([]FormattedAccount, 0, len(accounts))
		for _, account := range accounts {
			formatted = append(formatted, FormattedAccount{
				Name:        account.Name,
				UUID:        account.UUID,
				Description: account.Description,
				Type:        account.Type,
				CreateDate:  formatTime(account.CreateDate),
			})
		}
		return common.PrintQuery(cobraCmd, formatted, result.Total)
	},
}

var usersCmd = &cobra.Command{
	Use:   "users [name]",
	Short: "List account users",
	Long:  `List ZStack account users.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cobraCmd *cobra.Command, args []string) error {

		queryParam, err := common.BuildQueryParams(cobraCmd, args, "name")
		if err != nil {
			return fmt.Errorf("Error building query parameters: %w", err)
		}

		result, err := common.Fetch[view.UserInventoryView](cobraCmd, queryParam, "v1/accounts/users")
		users := result.Items
		if err != nil {
			return fmt.Errorf("Error querying users: %w", err)
		}

		formatted := make([]FormattedUser, 0, len(users))
		for _, user := range users {
			formatted = append(formatted, FormattedUser{
				Name:        user.Name,
				UUID:        user.UUID,
				AccountUUID: user.AccountUuid,
				Description: user.Description,
				CreateDate:  formatTime(user.CreateDate),
			})
		}
		return common.PrintQuery(cobraCmd, formatted, result.Total)
	},
}

var iam2ProjectsCmd = &cobra.Command{
	Use:     "projects [name]",
	Aliases: []string{"iam2-projects"},
	Short:   "List IAM2 projects",
	Long:    `List ZStack IAM2 projects.`,
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cobraCmd *cobra.Command, args []string) error {

		queryParam, err := common.BuildQueryParams(cobraCmd, args, "name")
		if err != nil {
			return fmt.Errorf("Error building query parameters: %w", err)
		}

		result, err := common.Fetch[view.IAM2ProjectInventoryView](cobraCmd, queryParam, "v1/iam2/projects")
		projects := result.Items
		if err != nil {
			return fmt.Errorf("Error querying projects: %w", err)
		}

		formatted := make([]FormattedIAM2Project, 0, len(projects))
		for _, project := range projects {
			formatted = append(formatted, FormattedIAM2Project{
				Name:              project.Name,
				UUID:              project.UUID,
				Description:       project.Description,
				State:             project.State,
				LinkedAccountUUID: project.LinkedAccountUuid,
				CreateDate:        formatTime(project.CreateDate),
			})
		}
		return common.PrintQuery(cobraCmd, formatted, result.Total)
	},
}

var iam2OrganizationsCmd = &cobra.Command{
	Use:     "organizations [name]",
	Aliases: []string{"orgs", "iam2-organizations"},
	Short:   "List IAM2 organizations",
	Long:    `List ZStack IAM2 organizations.`,
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cobraCmd *cobra.Command, args []string) error {

		queryParam, err := common.BuildQueryParams(cobraCmd, args, "name")
		if err != nil {
			return fmt.Errorf("Error building query parameters: %w", err)
		}

		result, err := common.Fetch[view.IAM2OrganizationInventoryView](cobraCmd, queryParam, "v1/iam2/organizations")
		organizations := result.Items
		if err != nil {
			return fmt.Errorf("Error querying organizations: %w", err)
		}

		formatted := make([]FormattedIAM2Organization, 0, len(organizations))
		for _, organization := range organizations {
			formatted = append(formatted, FormattedIAM2Organization{
				Name:                 organization.Name,
				UUID:                 organization.UUID,
				Description:          organization.Description,
				State:                organization.State,
				Type:                 organization.Type,
				ParentUUID:           organization.ParentUuid,
				RootOrganizationUUID: organization.RootOrganizationUuid,
				CreateDate:           formatTime(organization.CreateDate),
			})
		}
		return common.PrintQuery(cobraCmd, formatted, result.Total)
	},
}

var rolesCmd = &cobra.Command{
	Use:   "roles [name]",
	Short: "List IAM roles",
	Long:  `List ZStack IAM roles.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cobraCmd *cobra.Command, args []string) error {

		queryParam, err := common.BuildQueryParams(cobraCmd, args, "name")
		if err != nil {
			return fmt.Errorf("Error building query parameters: %w", err)
		}

		result, err := common.Fetch[view.RoleInventoryView](cobraCmd, queryParam, "v1/identities/roles")
		roles := result.Items
		if err != nil {
			return fmt.Errorf("Error querying roles: %w", err)
		}

		formatted := make([]FormattedRole, 0, len(roles))
		for _, role := range roles {
			formatted = append(formatted, FormattedRole{
				Name:        role.Name,
				UUID:        role.UUID,
				Description: role.Description,
				Identity:    role.Identity,
				Type:        role.Type,
				State:       role.State,
				CreateDate:  formatTime(role.CreateDate),
			})
		}
		return common.PrintQuery(cobraCmd, formatted, result.Total)
	},
}

var policiesCmd = &cobra.Command{
	Use:   "policies [name]",
	Short: "List IAM policies",
	Long:  `List ZStack IAM policies.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cobraCmd *cobra.Command, args []string) error {

		queryParam, err := common.BuildQueryParams(cobraCmd, args, "name")
		if err != nil {
			return fmt.Errorf("Error building query parameters: %w", err)
		}

		result, err := common.Fetch[view.PolicyInventoryView](cobraCmd, queryParam, "v1/accounts/policies")
		policies := result.Items
		if err != nil {
			return fmt.Errorf("Error querying policies: %w", err)
		}

		formatted := make([]FormattedPolicy, 0, len(policies))
		for _, policy := range policies {
			formatted = append(formatted, FormattedPolicy{
				Name:           policy.Name,
				UUID:           policy.UUID,
				AccountUUID:    policy.AccountUuid,
				StatementCount: len(policy.Statements),
				CreateDate:     formatTime(policy.CreateDate),
			})
		}
		return common.PrintQuery(cobraCmd, formatted, result.Total)
	},
}

func init() {
	GetCmd.AddCommand(accountsCmd, usersCmd, iam2ProjectsCmd, iam2OrganizationsCmd, rolesCmd, policiesCmd)

	common.AddQueryFlags(accountsCmd)
	common.AddQueryFlags(usersCmd)
	common.AddQueryFlags(iam2ProjectsCmd)
	common.AddQueryFlags(iam2OrganizationsCmd)
	common.AddQueryFlags(rolesCmd)
	common.AddQueryFlags(policiesCmd)
}

func formatTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format(time.RFC3339)
}
