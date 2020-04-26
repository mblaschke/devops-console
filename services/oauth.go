package services

import (
	"context"
	"devops-console/models"
	"encoding/json"
	"fmt"
	"github.com/coreos/go-oidc"
	"golang.org/x/oauth2"
	"strings"
)

var (
	OAuthProvider string
)

type OAuth struct {
	Config models.OAuthConfig

	config       *oauth2.Config
	oidcProvider *oidc.Provider

	Host string
}

func (o *OAuth) GetConfig() (config *oauth2.Config) {
	if o.config == nil {
		o.config = o.buildConfig()
	}
	config = o.config
	return
}

func (o *OAuth) GetProvider() string {
	return o.Config.Provider
}

func (o *OAuth) AuthCodeURL(state string) string {
	return o.GetConfig().AuthCodeURL(state)
}

func (o *OAuth) Exchange(code string) (*oauth2.Token, error) {
	return o.GetConfig().Exchange(context.Background(), code)
}

func (o *OAuth) FetchUserInfo(token *oauth2.Token) (user models.User, error error) {
	ctx := context.Background()

	switch strings.ToLower(o.Config.Provider) {
	case "azuread":
		tokenSource := oauth2.StaticTokenSource(token)

		// parse basic user info
		userInfo, err := o.oidcProvider.UserInfo(ctx, tokenSource)
		if err != nil {
			error = err
			return
		}

		// get prefixes from configuration
		userPrefix := o.Config.UsernamePrefix
		groupsPrefix := o.Config.GroupPrefix

		// parse custom userinfo
		aadUserInfo := struct {
			Directory    string   `json:"iss"`
			DirectoryId  string   `json:"tid"`
			UserId       string   `json:"oid"`
			Username     string   `json:"upn"`
			PrefUsername string   `json:"preferred_username"`
			Groups       []string `json:"groups"`
		}{}
		if err := userInfo.Claims(&aadUserInfo); err != nil {
			error = err
			return
		}
		// WORKAROUND: azuread groups (json array as string?!)
		var groupList []string
		for _, val := range aadUserInfo.Groups {
			var tmp []interface{}
			if err := json.Unmarshal([]byte(val), &tmp); err == nil {
				for _, groupName := range tmp {
					groupList = append(groupList, groupName.(string))
				}
			} else {
				groupList = append(groupList, val)
			}
		}

		// add prefix
		for i, val := range groupList {
			groupList[i] = groupsPrefix + val
		}
		aadUserInfo.Groups = groupList

		// extract username from email
		split := strings.SplitN(aadUserInfo.Username, "@", 2)

		// Build user object
		user.Uuid = aadUserInfo.UserId
		user.Id = userPrefix + aadUserInfo.UserId
		user.Username = split[0]
		user.Email = aadUserInfo.Username
		user.Groups = aadUserInfo.Groups

	default:
		o.error(fmt.Sprintf("oauth.provider \"%s\" is not valid", OAuthProvider))
	}

	return
}

func (o *OAuth) buildConfig() (config *oauth2.Config) {
	var clientId, clientSecret string
	var endpoint oauth2.Endpoint

	ctx := context.Background()

	scopes := []string{}

	switch strings.ToLower(o.Config.Provider) {
	case "azuread":
		aadTenant := o.Config.Azure.Tenant

		provider, err := oidc.NewProvider(ctx, fmt.Sprintf("https://sts.windows.net/%s/", aadTenant))
		if err != nil {
			o.error(fmt.Sprintf("oauth.provider AzureAD init failed: %s", err))
		}

		o.oidcProvider = provider
		endpoint = provider.Endpoint()
		scopes = []string{oidc.ScopeOpenID, "profile", "email"}
	default:
		o.error(fmt.Sprintf("oauth.provider \"%s\" is not valid", OAuthProvider))
	}

	if o.Config.Azure.EndpointAuth != "" {
		endpoint.AuthURL = o.Config.Azure.EndpointAuth
	}

	if o.Config.Azure.EndpointToken != "" {
		endpoint.TokenURL = o.Config.Azure.EndpointToken
	}

	clientId = o.Config.Azure.ClientId
	if clientId == "" {
		o.error("No oauth.client.id configured")
	}

	clientSecret = o.Config.Azure.ClientSecret
	if clientSecret == "" {
		o.error("No oauth.client.secret configured")
	}

	config = &oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		Endpoint:     endpoint,
		Scopes:       scopes,
		RedirectURL:  o.buildRedirectUrl(),
	}

	return
}

func (o *OAuth) buildRedirectUrl() (url string) {
	url = o.Config.RedirectUrl
	url = strings.ReplaceAll(url, "$host", o.Host)
	return
}

func (o *OAuth) error(message string) {
	panic(message)
}
