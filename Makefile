PROJECT_NAME		:= $(shell basename $(CURDIR))
GIT_TAG				:= $(shell git describe --dirty --tags --always)
GIT_COMMIT			:= $(shell git rev-parse --short HEAD)
LDFLAGS				:= -X "main.gitTag=$(GIT_TAG)" -X "main.gitCommit=$(GIT_COMMIT)" -extldflags "-static" -s -w

FIRST_GOPATH			:= $(firstword $(subst :, ,$(shell go env GOPATH)))
GOLANGCI_LINT_BIN		:= $(FIRST_GOPATH)/bin/golangci-lint
GOSEC_BIN				:= $(FIRST_GOPATH)/bin/gosec

.PHONY: all
all: build-frontend build-backend

.PHONY: clean
clean:
	git clean -Xfd .

.PHONY: build-frontend
build-frontend:
	npm run --prefix=react build
	cat react/build/index.html | sed 's/ defer="defer"//g' | sed 's/<script/<script nonce="{{ CSP_NONCE }}"/g' > templates/includes/react.jet
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

.PHONY: build-backend
build-backend:
	GOOS=${GOOS} GOARCH=${GOARCH} CGO_ENABLED=0 go build -ldflags '$(LDFLAGS)' -o $(PROJECT_NAME) .

.PHONY: vendor
vendor:
	go mod tidy
	go mod vendor
	go mod verify

.PHONY: image
image: build
	docker build -t $(PROJECT_NAME):$(GIT_TAG) .

build-push-development:
	docker buildx create --use
	docker buildx build -t webdevops/$(PROJECT_NAME):development --platform linux/amd64,linux/arm,linux/arm64 --push .

.PHONY: test
test:
	go test ./...

.PHONY: dependencies
dependencies:
	go mod vendor

.PHONY: check-release
check-release: vendor lint gosec test

.PHONY: lint
lint: $(GOLANGCI_LINT_BIN)
	$(GOLANGCI_LINT_BIN) run -E exportloopref,gofmt --timeout=30m

.PHONY: gosec
gosec: $(GOSEC_BIN)
	$(GOSEC_BIN) ./...

$(GOLANGCI_LINT_BIN):
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(FIRST_GOPATH)/bin

$(GOSEC_BIN):
	curl -sfL https://raw.githubusercontent.com/securego/gosec/master/install.sh | sh -s -- -b $(FIRST_GOPATH)/bin

.PHONY: recreate-go-mod
recreate-go-mod:
	rm -f go.mod go.sum
	go mod init devops-console
	go get k8s.io/client-go@v0.23.0
	go get -u github.com/Azure/azure-sdk-for-go/...
	go get -u github.com/microcosm-cc/bluemonday
	go get
	go mod vendor
