package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
)

type (
	User struct {
		Uuid     string   `json:"Uuid"`
		Id       string   `json:"id"`
		Username string   `json:"u"`
		Email    string   `json:"e"`
		Teams    []Team   `json:"-"`
		Groups   []string `json:"g"`
		IsAdmin  bool     `json:"a"`
	}
)

func (u *User) init(config *AppConfig) {
	u.initTeams(config)
}

func (u *User) initTeams(config *AppConfig) (teams []Team) {
	teamList := map[string]string{}

	if config == nil {
		return
	}

	// Default teams
	for _, teamName := range config.Permissions.Default.Teams {
		teamList[teamName] = teamName
	}

	// User teams
	if config.Permissions.User != nil {
		if val, exists := config.Permissions.User[u.Username]; exists {
			for _, teamName := range val.Teams {
				teamList[teamName] = teamName
			}
		}
	}

	// Group teams
	if config.Permissions.Group != nil {
		for _, group := range u.Groups {
			if val, exists := config.Permissions.Group[group]; exists {
				for _, teamName := range val.Teams {
					teamList[teamName] = teamName
				}
			}
		}
	}

	// Sort
	teamNameList := make([]string, 0, len(teamList))
	for teamName := range teamList {
		teamNameList = append(teamNameList, teamName)
	}
	sort.Strings(teamNameList)

	// Build teams (with permissions)
	for _, teamName := range teamNameList {
		if _, exists := config.Permissions.Team[teamName]; exists {
			teamConfig := config.Permissions.Team[teamName]
			teams = append(teams, Team{Name: teamName, K8sPermissions: teamConfig.K8sRoleBinding, AzureRoleAssignments: teamConfig.AzureRoleAssignments})
		}
	}

	u.Teams = teams
	return
}

func (u *User) GetTeam(name string) (team *Team, err error) {
	for _, val := range u.Teams {
		if val.Name == name {
			teamVal := val
			team = &teamVal
			break
		}
	}

	if team == nil {
		err = errors.New("team not found")
	}

	return
}

func (u *User) LoggedIn() bool {
	return u.Id != ""
}

func (u *User) IsMemberOf(teamName string) (status bool) {
	status = false

	// check for invalid teamname
	if teamName == "" {
		return
	}

	for _, team := range u.Teams {
		if teamName == team.Name {
			status = true
			break
		}
	}

	return
}

func (u *User) ToJson() (jsonString string, error error) {
	jsonBytes, err := json.Marshal(u)
	if err != nil {
		error = err
		return
	}

	jsonString = string(jsonBytes)
	return
}

func UserCreateFromJson(jsonString string, config *AppConfig) (u *User, err error) {
	if err = json.Unmarshal([]byte(jsonString), &u); err == nil {
		u.init(config)
	}

	return
}

func (u *User) String() string {
	return fmt.Sprintf("User(%s)", u.Username)
}
