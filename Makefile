.PHONY: all build-run build-backend build-backend run vendor

NAME				:= devops-console
GIT_TAG				:= $(shell git describe --dirty --tags --always)
GIT_COMMIT			:= $(shell git rev-parse --short HEAD)
LDFLAGS             := -X "main.gitTag=$(GIT_TAG)" -X "main.gitCommit=$(GIT_COMMIT)" -extldflags "-static"

FIRST_GOPATH		:= $(firstword $(subst :, ,$(shell go env GOPATH)))
GOLANGCI_LINT_BIN	:= $(FIRST_GOPATH)/bin/golangci-lint

all: vendor build-frontend build-backend
build: build-frontend build-backend

build-run: build-frontend build-backend run

recreate-go-mod:
	rm -f go.mod go.sum
	GO111MODULE=on go mod init
	GO111MODULE=on go get k8s.io/client-go@v0.17.0
	GO111MODULE=on go get -u github.com/Azure/azure-sdk-for-go/...
	GO111MODULE=on go get
	GO111MODULE=on go mod vendor

image: build
	docker build -t $(NAME):$(TAG) .

build-backend:
	CGO_ENABLED=0 go build -a -ldflags '$(LDFLAGS)' -o $(NAME) .

run:
	./devops-console

vendor:
	go mod tidy
	go mod vendor
	go mod verify

build-frontend:
	npm run --prefix=react build
	cp react/build/index.html templates/includes/react.jet
	rm -rf static/js static/dist/boostrap static/dist/popper.js static/dist/sb-admin static/dist/webfonts
	mkdir -p static/js
	cp react/build/static/js/* static/js
	cp -a react/node_modules/bootstrap/dist/ static/dist/bootstrap
	cp -a react/node_modules/popper.js/dist/umd/ static/dist/popper.js
	mkdir -p static/dist/sb-admin static/dist/webfonts
	cp -a react/node_modules/startbootstrap-sb-admin/css/sb-admin.css static/dist/sb-admin/sb-admin.css
	cp -a react/node_modules/startbootstrap-sb-admin/js/sb-admin.js static/dist/sb-admin/sb-admin.js
	cp -a react/node_modules/startbootstrap-sb-admin/vendor/fontawesome-free/css/all.min.css static/dist/sb-admin/fontawesome.css
	cp -a react/node_modules/startbootstrap-sb-admin/vendor/fontawesome-free/webfonts/* static/dist/webfonts/
	touch static/js/.gitkeep

.PHONY: lint
lint: $(GOLANGCI_LINT_BIN)
	# megacheck fails to respect build flags, causing compilation failure during linting.
	# instead, use the unused, gosimple, and staticcheck linters directly
	$(GOLANGCI_LINT_BIN) run -D megacheck -E unused,gosimple,staticcheck --timeout=10m

dependencies: $(GOLANGCI_LINT_BIN)

$(GOLANGCI_LINT_BIN):
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $(FIRST_GOPATH)/bin v1.23.8

