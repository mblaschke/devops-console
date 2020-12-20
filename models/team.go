package models

import (
	"fmt"
)

type (
	TeamPermissionsServiceAccount struct {
		Name      string `yaml:"name"`
		Namespace string `yaml:"namespace"`
	}

	TeamK8sPermissions struct {
		Name            string                          `yaml:"name"`
		Groups          []string                        `yaml:"groups"`
		Users           []string                        `yaml:"users"`
		ServiceAccounts []TeamPermissionsServiceAccount `yaml:"serviceaccounts"`
		ClusterRole     string                          `yaml:"clusterrole"`
	}

	TeamAzureRoleAssignments struct {
		PrincipalId string `yaml:"principalid"`
		Role        string `yaml:"role"`
	}

	TeamServiceConnections struct {
		Token string `yaml:"token"`
	}

	Team struct {
		Name                 string                     `json:"-"`
		K8sPermissions       []TeamK8sPermissions       `json:"-"`
		AzureRoleAssignments []TeamAzureRoleAssignments `json:"-"`
	}
)

func (t *Team) String() string {
	return fmt.Sprintf("Team(%s)", t.Name)
}
