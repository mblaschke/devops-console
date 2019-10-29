package main

import (
	"bytes"
	"devops-console/models"
	"devops-console/services"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Masterminds/sprig"
	glogger "github.com/google/logger"
	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/logger"
	"github.com/kataras/iris/middleware/recover"
	"github.com/kataras/iris/sessions"
	"github.com/kataras/iris/view"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sync/atomic"
	"text/template"
)

type Server struct {
	app  *iris.Application
	tmpl *view.JetEngine

	session *sessions.Sessions

	config models.AppConfig

	logger *glogger.Logger
}

func NewServer(pathList []string) *Server {
	server := Server{}
	server.app = iris.New()
	server.logger = glogger.Init("Verbose", true, false, ioutil.Discard)
	server.app.Use(recover.New())
	server.app.Use(logger.New(logger.Config{
		// Status displays status code
		Status: true,
		// IP displays request's remote address
		IP: true,
		// Method displays the http method
		Method: true,
		// Path displays the request path
		Path: true,
		// Query appends the url query to the Path.
		Query: opts.Debug,
		// if !empty then its contents derives from `ctx.Values().Get("logger_message")
		// will be added to the logs.
		MessageContextKeys: []string{"userIdentification"},
	}))

	server.Init()
	for _, config := range pathList {
		server.setupConfig(config)
	}
	server.setupKubernetes()
	server.initApp()

	return &server
}

func (c *Server) Init() {
	c.config = models.AppConfig{}
}

func (c *Server) setupConfig(path string) {
	var config []byte

	c.logger.Infof("reading configuration from file %v", path)
	if data, err := ioutil.ReadFile(path); err == nil {
		config = data
	} else {
		panic(err)
	}

	c.logger.Info(" -  preprocessing with template engine")
	var tmplBytes bytes.Buffer
	parsedConfig, err := template.New("yaml").Funcs(sprig.TxtFuncMap()).Parse(string(config))
	if err != nil {
		panic(err)
	}

	if err := parsedConfig.Execute(&tmplBytes, nil); err != nil {
		panic(err)
	}

	if opts.DumpConfig {
		fmt.Println(tmplBytes.String())
		os.Exit(1)
	}

	c.logger.Info(" -  parsing with yaml")

	if err := yaml.Unmarshal(tmplBytes.Bytes(), &c.config); err != nil {
		panic(err)
	}

}

func (c *Server) setupKubernetes() {
	KubeNamespaceConfig := map[string]*models.KubernetesObjectList{}

	c.logger.Infof("setup kubernetes object configuration (for namespaces)")

	objectsPath := c.config.App.Kubernetes.ObjectsPath
	if !filepath.IsAbs(objectsPath) {
		if currentWorkDir, err := os.Getwd(); err == nil {
			objectsPath = filepath.Join(currentWorkDir, objectsPath)
		} else {
			panic(err)
		}
	}

	if objectsPath != "" {
		// default namespace settings
		k8sDefaultPath := filepath.Join(objectsPath, "_default")
		if IsDirectory(k8sDefaultPath) {
			c.logger.Infof(" - using default path %v", k8sDefaultPath)
		} else {
			k8sDefaultPath = ""
		}

		KubeNamespaceConfig["_default"] = buildKubeConfigList("", k8sDefaultPath)

		// parse config for each subpath as environment
		err := filepath.Walk(objectsPath, func(path string, info os.FileInfo, err error) error {
			// jump into base dir
			if path == objectsPath {
				return nil
			}

			// parse configs in dir but don't jump recursive into it
			if info.IsDir() && path != k8sDefaultPath {
				c.logger.Infof(" - processing %v", path)

				KubeNamespaceConfig[info.Name()] = buildKubeConfigList(k8sDefaultPath, path)
				return filepath.SkipDir
			}
			return nil
		})

		if err != nil {
			panic(err)
		}
	}

	c.config.App.Kubernetes.ObjectsList = KubeNamespaceConfig
}

func (c *Server) initApp() {
	c.logger.Infof("setup app server")
	c.initSession()
	c.initTemplateEngine()
	c.initRoutes()

	c.app.UseGlobal(c.before)
}

func (c *Server) Run(addr string) {
	c.logger.Infof("run app server")

	c.app.Run(iris.Addr(addr))
}

func (c *Server) newServiceOauth(ctx iris.Context) services.OAuth {
	oauth := services.OAuth{Host: ctx.Host()}
	oauth.Config = c.config.App.Oauth
	return oauth
}

func (c *Server) respondError(ctx iris.Context, err error) {
	response := struct {
		Message string
	}{
		Message: fmt.Sprintf("Error: %v", err),
	}

	c.auditLog(ctx, response.Message)

	ctx.StatusCode(iris.StatusBadRequest)
	ctx.JSON(response)
	ctx.EndRequest()
	ctx.StopExecution()
	panic(ctx)
}

func (c *Server) auditLog(ctx iris.Context, message string) {
	username := "*anonymous*"
	user, _ := c.getUser(ctx)
	if user != nil {
		username = fmt.Sprintf("%s (%s)", user.Username, user.Uuid)
	}

	c.logger.Infof("AUDIT: user[%s]: %s", username, message)
}

func (c *Server) notificationMessage(ctx iris.Context, message string) {
	if c.config.App.Notification.Slack.Webhook == "" {
		return
	}

	username := "*anonymous*"
	user, _ := c.getUser(ctx)
	if user != nil {
		username = fmt.Sprintf("%s (%s)", user.Username, user.Uuid)
	}

	payload := struct {
		Channel  string `json:"channel"`
		Username string `json:"username"`
		Text     string `json:"text"`
	}{
		Channel:  c.config.App.Notification.Slack.Channel,
		Username: "devops-console",
		Text:     fmt.Sprintf(c.config.App.Notification.Slack.Message, username, message),
	}

	payloadJson, _ := json.Marshal(payload)

	client := http.Client{}
	req, err := http.NewRequest("POST", c.config.App.Notification.Slack.Webhook, bytes.NewBuffer(payloadJson))
	defer req.Body.Close()
	if err != nil {
		c.logger.Errorf("Failed to send slack notification: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	_, err = client.Do(req)
	if err != nil {
		c.logger.Errorf("Failed to send slack notification: %v", err)
	}
}

func (c *Server) before(ctx iris.Context) {
	atomic.AddInt64(&requestCounter, 1)

	// security headers
	ctx.Header("X-Frame-Options", "SAMEORIGIN")
	ctx.Header("X-XSS-Protection", "1; mode=block")
	ctx.Header("X-Content-Type-Options", "nosniff")

	// view information
	ctx.ViewData("navigationRoute", ctx.GetCurrentRoute().Path())
	ctx.Next()
}

func (c *Server) index(ctx iris.Context) {
	user, err := c.getUser(ctx)

	if err == nil && user != nil {
		ctx.Redirect("/kubernetes/namespaces")
	} else {
		ctx.View("login.jet")
	}
}

func (c *Server) template(ctx iris.Context, title, template string) {
	c.ensureLoggedIn(ctx, func(ctx iris.Context) {
		ctx.ViewData("title", title)
		ctx.View(template)
	})
}

func (c *Server) react(ctx iris.Context, title string) {
	c.ensureLoggedIn(ctx, func(ctx iris.Context) {
		ctx.ViewData("title", title)
		ctx.View("react.jet")
	})
}

func (c *Server) ensureLoggedIn(ctx iris.Context, callback func(ctx iris.Context)) {
	c.session.Start(ctx)
	user, err := c.getUser(ctx)

	if err != nil {
		c.session.Destroy(ctx)
		ctx.ViewData("messageError", "Invalid session")
		ctx.View("login.jet")
		return
	}

	ctx.ViewData("user", user)
	ctx.Values().Set("userIdentification", fmt.Sprintf("%v[%v]", user.Username, user.Uuid))
	callback(ctx)
}

func (c *Server) getUser(ctx iris.Context) (user *models.User, err error) {
	s := c.session.Start(ctx)
	userJson := s.GetString("user")

	if opts.Debug {
		if val := os.Getenv("DEBUG_SESSION_USER"); val != "" {
			s.Set("user", "DEBUG_SESSION_USER")
			userJson = val
		}
	}

	user, err = models.UserCreateFromJson(userJson, &c.config)
	return
}

func (c *Server) getUserOrStop(ctx iris.Context) (user *models.User) {
	var err error
	user, err = c.getUser(ctx)

	if err != nil || user == nil {
		c.session.Destroy(ctx)
		c.respondError(ctx, errors.New("Invalid session"))
	}

	return
}

