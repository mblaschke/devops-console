NAME				:= devops-console
GIT_TAG				:= $(shell git describe --dirty --tags --always)
GIT_COMMIT			:= $(shell git rev-parse --short HEAD)
LDFLAGS             := -X "main.gitTag=$(GIT_TAG)" -X "main.gitCommit=$(GIT_COMMIT)" -extldflags "-static"

FIRST_GOPATH		:= $(firstword $(subst :, ,$(shell go env GOPATH)))
GOLANGCI_LINT_BIN	:= $(FIRST_GOPATH)/bin/golangci-lint
GOSEC_BIN			:= $(FIRST_GOPATH)/bin/gosec

.PHONY: all
all: vendor build-frontend build-backend

.PHONY: build
build: build-frontend build-backend

.PHONY: build-run
build-run: build-frontend build-backend run

.PHONY: recreate-go-mod
recreate-go-mod:
	rm -f go.mod go.sum
	go mod init devops-console
	go get k8s.io/client-go@v0.22.0
	go get -u github.com/Azure/azure-sdk-for-go/...
	go get -u github.com/microcosm-cc/bluemonday
	go get
	go mod vendor

.PHONY: image
image: build
	docker build -t $(NAME):$(TAG) .

.PHONY: build-backend
build-backend:
	CGO_ENABLED=0 go build -a -ldflags '$(LDFLAGS)' -o $(NAME) .

.PHONY: run
run:
	./devops-console

.PHONY: vendor
vendor:
	go mod tidy
	go mod vendor
	go mod verify

.PHONY: build-frontend
build-frontend:
	npm run --prefix=react build
	cp react/build/index.html templates/includes/react.jet
	cat templates/includes/react.jet | sed 's/ defer="defer"//g' | sed 's/<script/<script nonce="{{ CSP_NONCE }}"/g' | tee templates/includes/react.jet > /dev/null
	test -s templates/includes/react.jet
	rm -rf static/js static/dist/boostrap static/dist/popper.js static/dist/sb-admin static/dist/webfonts
	mkdir -p static/js
	cp react/build/static/js/* static/js
	cp -a react/node_modules/bootstrap/dist/ static/dist/bootstrap
	cp -a react/node_modules/@popperjs/core/dist/umd/ static/dist/popper.js
	mkdir -p static/dist/sb-admin static/dist/webfonts static/dist/jquery/
	cp -a react/node_modules/jquery/dist/jquery.min.js static/dist/jquery/jquery.min.js
	cp -a react/node_modules/startbootstrap-sb-admin/dist/css/styles.css static/dist/sb-admin/sb-admin.css
	cp -a react/node_modules/startbootstrap-sb-admin/dist/js/scripts.js static/dist/sb-admin/sb-admin.js
	cp -a react/node_modules/@fortawesome/fontawesome-free/css/all.min.css static/dist/sb-admin/fontawesome.css
	cp -a react/node_modules/@fortawesome/fontawesome-free/webfonts/* static/dist/webfonts/
	touch static/js/.gitkeep

.PHONY: test
test:
	go test ./...

.PHONY: lint
lint: $(GOLANGCI_LINT_BIN)
	$(GOLANGCI_LINT_BIN) run -E exportloopref,gofmt --timeout=25m

.PHONY: gosec
gosec: $(GOSEC_BIN)
	$(GOSEC_BIN) ./...

.PHONY: dependencies
dependencies: $(GOLANGCI_LINT_BIN) $(GOSEC_BIN)

$(GOLANGCI_LINT_BIN):
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

$(GOSEC_BIN):
	curl -sfL https://raw.githubusercontent.com/securego/gosec/master/install.sh | sh -s -- -b $(FIRST_GOPATH)/bin v2.7.0
