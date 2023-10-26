package models

import (
	"encoding/json"
	"errors"
	"fmt"
)

type (
	User struct {
		Uuid     string   `json:"Uuid"`
		Id       string   `json:"id"`
		Username string   `json:"upn"`
		Email    string   `json:"mail"`
		Groups   []string `json:"groups"`
		IsAdmin  bool     `json:"isAdmin"`

		config *AppConfig
	}
)

func (u *User) ApplyAppConfig(config *AppConfig) {
	u.config = config

	u.IsAdmin = false

adminLoop:
	for _, groupId := range u.Groups {
		for _, adminGroupId := range u.config.Permissions.AdminGroups {
			if groupId == adminGroupId {
				u.IsAdmin = true
				break adminLoop
			}
		}
	}
}

func (u *User) GetTeams() (teamList []*AppConfigTeam) {
	if u.IsAdmin {
		for _, team := range u.config.Permissions.Teams {
			teamList = append(teamList, team)
		}
		return
	}

	for _, groupId := range u.Groups {
		if team, exists := u.config.Permissions.Teams[groupId]; exists {
			teamList = append(teamList, team)
		}
	}

	return
}

func (u *User) GetTeam(name string) (*AppConfigTeam, error) {
	if u.IsAdmin {
		for _, row := range u.config.Permissions.Teams {
			team := row
			if team.Name == name {
				return team, nil
			}
		}
	}

	for _, groupId := range u.Groups {
		if team, exists := u.config.Permissions.Teams[groupId]; exists {
			if team.Name == name {
				return team, nil
			}
		}
	}

	return nil, errors.New("team not found")
}

func (u *User) LoggedIn() bool {
	return u.Id != ""
}

func (u *User) IsMemberOf(name string) bool {
	if u.IsAdmin {
		return true
	}

	if name == "" {
		if team, err := u.GetTeam(name); team != nil && err == nil {
			return true
		}
	}

	return false
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
		u.ApplyAppConfig(config)
	}

	return
}

func (u *User) String() string {
	return fmt.Sprintf("User(%s)", u.Username)
}
