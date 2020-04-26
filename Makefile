.PHONY: all build-run build-backend build-backend run

NAME				:= devops-console
TAG					:= $(shell git rev-parse --short HEAD)

FIRST_GOPATH		:= $(firstword $(subst :, ,$(shell go env GOPATH)))
GOLANGCI_LINT_BIN	:= $(FIRST_GOPATH)/bin/golangci-lint

all: vendor build-frontend build-backend
build: build-frontend build-backend

build-run: build-frontend build-backend run

image: build
	docker build -t $(NAME):$(TAG) .

build-backend:
	#go-bindata ./templates/...
	CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o $(NAME) .

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

