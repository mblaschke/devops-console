package main

import (
	"bytes"
	"devops-console/models"
	"fmt"
	"github.com/Masterminds/sprig"
	glogger "github.com/google/logger"
	iris "github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/recover"
	"github.com/kataras/iris/v12/sessions"
	"github.com/kataras/iris/v12/sessions/sessiondb/redis"
	"github.com/kataras/iris/v12/view"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"
)

type Server struct {
	app  *iris.Application
	tmpl *view.JetEngine

	session *sessions.Sessions

	redisConnection *redis.Database

	config models.AppConfig

	logger *glogger.Logger
}

func NewServer(pathList []string) *Server {
	server := Server{}

	server.app = iris.New()
	server.logger = glogger.Init("Verbose", true, false, ioutil.Discard)
	server.app.Use(recover.New())

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

	if data, err := ioutil.ReadFile(filepath.Clean(path)); err == nil {
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
}

func (c *Server) Run(addr string) {
	c.logger.Infof("run app server")
	if err := c.app.Run(iris.Addr(addr)); err != nil {
		c.logger.Fatalln(err)
	}
}

func (c *Server) responseJson(ctx iris.Context, v interface{}) {
	if _, err := ctx.JSON(v); err != nil {
		c.logger.Errorln(err)
	}
}
