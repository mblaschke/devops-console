package main

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	iris "github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
	"github.com/oauth2-proxy/oauth2-proxy/v7/pkg/apis/options"
	"github.com/oauth2-proxy/oauth2-proxy/v7/pkg/encryption"
	"github.com/oauth2-proxy/oauth2-proxy/v7/pkg/requests"
	"github.com/prometheus/client_golang/prometheus"

	providers "github.com/oauth2-proxy/oauth2-proxy/v7/providers"

	"github.com/mblaschke/devops-console/models"
)

type ApplicationAuth struct {
	*Server

	provider *providers.AzureProvider
}

func NewApplicationAuth(c *Server) *ApplicationAuth {
	app := ApplicationAuth{Server: c}

	var (
		azureLoginURL, azureRedeemURL, azureProfileURL *url.URL
	)

	switch strings.ToLower(opts.Azure.Environment) {
	case "azurecloud", "azurepubliccloud":
		azureLoginURL = &url.URL{
			Scheme: "https",
			Host:   "login.microsoftonline.com",
			Path:   fmt.Sprintf("/%v/oauth2/v2.0/authorize", c.config.App.Oauth.Azure.Tenant),
		}

		azureRedeemURL = &url.URL{
			Scheme: "https",
			Host:   "login.microsoftonline.com",
			Path:   fmt.Sprintf("/%v/oauth2/v2.0/token", c.config.App.Oauth.Azure.Tenant),
		}

		azureProfileURL = &url.URL{
			Scheme: "https",
			Host:   "graph.microsoft.com",
			Path:   "/v1.0/me",
		}

	case "azurechina", "azurechinacloud":
		azureLoginURL = &url.URL{
			Scheme: "https",
			Host:   "login.partner.microsoftonline.cn",
			Path:   fmt.Sprintf("/%v/oauth2/v2.0/authorize", c.config.App.Oauth.Azure.Tenant),
		}

		azureRedeemURL = &url.URL{
			Scheme: "https",
			Host:   "login.partner.microsoftonline.cn",
			Path:   fmt.Sprintf("/%v/oauth2/v2.0/token", c.config.App.Oauth.Azure.Tenant),
		}

		azureProfileURL = &url.URL{
			Scheme: "https",
			Host:   "microsoftgraph.chinacloudapi.cn",
			Path:   "/v1.0/me",
		}
	default:
		panic(fmt.Sprintf(`Azure environment "%v" not supported`, opts.Azure.Environment))
	}

	providerData := providers.ProviderData{
		ProviderName:                  "",
		LoginURL:                      azureLoginURL,
		RedeemURL:                     azureRedeemURL,
		ProfileURL:                    azureProfileURL,
		ProtectedResource:             nil,
		ValidateURL:                   nil,
		ClientID:                      c.config.App.Oauth.Azure.ClientId,
		ClientSecret:                  c.config.App.Oauth.Azure.ClientSecret,
		ClientSecretFile:              "",
		Scope:                         "openid profile",
		CodeChallengeMethod:           "",
		SupportedCodeChallengeMethods: nil,
		AllowUnverifiedEmail:          false,
		Verifier:                      nil,
		AllowedGroups:                 nil,
	}
	azureOpts := options.AzureOptions{
		Tenant:          c.config.App.Oauth.Azure.Tenant,
		GraphGroupField: "",
	}
	app.provider = providers.NewAzureProvider(&providerData, azureOpts)
	return &app
}

func (c *ApplicationAuth) Login(ctx iris.Context) {
	s := c.recreateSession(ctx, func(ctx *context.Context, cookie *http.Cookie, op uint8) {
		if op == 1 {
			// need lax mode for oauth redirect
			cookie.SameSite = http.SameSiteLaxMode
		}
	})

	state, err := encryption.GenerateRandomASCIIString(96)
	if err != nil {
		ctx.ViewData("messageError", "OAuth failed: unable to create challenge state")
		c.templateLogin(ctx, true)
		return
	}
	s.Set("oauth:state", state)
	s.Set("oauth:redirect", ctx.FormValue("redirect"))

	loginUrl := c.config.App.Oauth.RedirectUrl
	loginUrl = strings.ReplaceAll(loginUrl, "$host", ctx.Host())

	loginURL := c.provider.GetLoginURL(
		loginUrl,
		// encryption.HashNonce(state),
		string(state),
		"",
		url.Values{},
	)

	PrometheusActions.With(prometheus.Labels{"scope": "oauth", "type": "start"}).Inc()

	ctx.Redirect(loginURL)
}

func (c *Server) Logout(ctx iris.Context) {
	ctx.ViewData("messageSuccess", "Logged out")
	c.templateLogin(ctx, true)
}

func (c *Server) LogoutForced(ctx iris.Context) {
	ctx.ViewData("ERROR_MESSAGE", "Session was terminated, please login again.")
	c.templateLogin(ctx, true)
}

func (c *ApplicationAuth) LoginViaOauth(ctx iris.Context) {
	s := c.getSession(ctx)

	loginUrl := c.config.App.Oauth.RedirectUrl
	loginUrl = strings.ReplaceAll(loginUrl, "$host", ctx.Host())

	redirectUrl := ""
	if val, ok := s.Get("oauth:redirect").(string); ok {
		redirectUrl = val
	}

	if s.Get("oauth:state") == "" || s.Get("oauth:state") == nil {
		ctx.ViewData("messageError", "OAuth pre check failed: invalid session")
		c.templateLogin(ctx, true)
		return
	}

	if err := ctx.URLParam("error"); err != "" {
		message := err
		if errorDesc := ctx.URLParam("error_description"); errorDesc != "" {
			message = fmt.Sprintf("%s:\n%s", err, errorDesc)
		}
		c.logger.Error(err)
		ctx.ViewData("messageError", message)
		c.templateLogin(ctx, true)
		return
	}

	code := ctx.URLParam("code")
	if code == "" {
		c.destroySession(ctx)
		ctx.ViewData("messageError", "OAuth pre check failed: code empty")
		c.templateLogin(ctx, true)
		return
	}

	state := ctx.URLParam("state")
	if state == "" {
		ctx.ViewData("messageError", "OAuth pre check failed: state empty")
		c.templateLogin(ctx, true)
		return
	}

	if state != s.Get("oauth:state") {
		ctx.ViewData("messageError", "OAuth pre check failed: state mismatch")
		c.templateLogin(ctx, true)
		return
	}

	oauthSess, err := c.provider.Redeem(ctx, loginUrl, code, "")
	if err != nil {
		c.logger.Error(err)
		ctx.ViewData("messageError", fmt.Sprintf("OAuth redeem failed: %v", err))
		c.templateLogin(ctx, true)
		return
	}

	err = c.provider.EnrichSession(ctx, oauthSess)
	if err != nil {
		c.logger.Error(err)
		ctx.ViewData("messageError", fmt.Sprintf("OAuth check failed: %v", err))
		c.templateLogin(ctx, true)
		return
	}

	// fetch profile from azuread
	profileResp, err := requests.New(c.provider.ProfileURL.String()).
		WithContext(ctx).
		WithHeaders(http.Header{
			"Authorization": []string{"Bearer " + oauthSess.AccessToken},
		}).
		Do().
		UnmarshalSimpleJSON()
	if err != nil {
		c.logger.Error(err)
		ctx.ViewData("messageError", fmt.Sprintf("OAuth profile fetch failed: %v", err))
		c.templateLogin(ctx, true)
		return
	}

	user := models.User{
		Uuid:     "",
		Id:       "",
		Username: "",
		Email:    oauthSess.Email,
		Teams:    nil,
		Groups:   oauthSess.Groups,
		IsAdmin:  false,
	}

	if val, err := profileResp.Get("id").String(); err == nil && val != "" {
		user.Uuid = val
		user.Id = val
	} else {
		ctx.ViewData("messageError", "OAuth login failed: provided uuid is empty")
		c.templateLogin(ctx, true)
		return
	}

	if val, err := profileResp.Get("userPrincipalName").String(); err == nil && val != "" {
		user.Username = val
	} else {
		ctx.ViewData("messageError", "OAuth login failed: provided userPrincipalName is empty")
		c.templateLogin(ctx, true)
		return
	}

	if c.config.App.Oauth.Filter.UsernameWhitelist != "" {
		filterRegexp := regexp.MustCompile(c.config.App.Oauth.Filter.UsernameWhitelist)
		if !filterRegexp.MatchString(user.Username) {
			ctx.ViewData("messageError", fmt.Sprintf(`user "%s" is not allowed to use this application (username not allowed)`, user.Username))
			c.templateLogin(ctx, true)
			return
		}
	}

	if c.config.App.Oauth.Filter.UsernameBlacklist != "" {
		filterRegexp := regexp.MustCompile(c.config.App.Oauth.Filter.UsernameBlacklist)
		if filterRegexp.MatchString(c.config.App.Oauth.Filter.UsernameBlacklist) {
			ctx.ViewData("messageError", fmt.Sprintf(`user "%s" is not allowed to use this application (username blacklisted)`, user.Username))
			c.templateLogin(ctx, true)
			return
		}
	}

	// apply team mapping
	user.ApplyAppConfig(&c.config)

	// check groups
	if len(user.Teams) == 0 {
		ctx.ViewData("messageError", fmt.Sprintf(`user "%s" is not allowed to use this application (no team assignments/mappings found)`, user.Username))
		c.templateLogin(ctx, true)
		return
	}

	if userSession, err := user.ToJson(); err == nil {
		// regenerate session
		s := c.recreateSession(ctx)

		// inject user information into session
		s.Set("user", userSession)
		c.csrfProtectionTokenRegenerate(ctx)
	} else {
		ctx.ViewData("messageError", "unable to set session")
		c.templateLogin(ctx, true)
		return
	}

	PrometheusActions.With(prometheus.Labels{"scope": "oauth", "type": "login"}).Inc()

	if redirectUrl != "" && c.checkRedirectUrl(redirectUrl) {
		c.redirectHtml(ctx, redirectUrl)
	} else {
		c.redirectHtml(ctx, "/home")
	}
}

func (c *Server) checkRedirectUrl(url string) bool {
	url = strings.TrimSpace(url)

	// must not be empty
	if url == "" {
		return false
	}

	// must start with /
	if !strings.HasPrefix(url, "/") {
		return false
	}

	// must start with /login
	if strings.HasPrefix(url, "/login") {
		return false
	}

	// must start with /logout
	if strings.HasPrefix(url, "/logout") {
		return false
	}

	// must not contain query
	if strings.Contains(url, "?") {
		return false
	}

	// must not contain http://
	if strings.Contains(url, "http://") {
		return false
	}

	// must not contain https://
	if strings.Contains(url, "https://") {
		return false
	}

	return true
}
