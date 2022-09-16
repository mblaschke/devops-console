PROJECT_NAME		:= $(shell basename $(CURDIR))
GIT_TAG				:= $(shell git describe --dirty --tags --always)
GIT_COMMIT			:= $(shell git rev-parse --short HEAD)
LDFLAGS				:= -X "main.gitTag=$(GIT_TAG)" -X "main.gitCommit=$(GIT_COMMIT)" -extldflags "-static" -s -w

FIRST_GOPATH			:= $(firstword $(subst :, ,$(shell go env GOPATH)))
GOLANGCI_LINT_BIN		:= $(FIRST_GOPATH)/bin/golangci-lint

.PHONY: all
all: vendor build

.PHONY: clean
clean:
	git clean -Xfd .

#######################################
# builds
#######################################

.PHONY: vendor
vendor:
	go mod tidy
	go mod vendor
	go mod verify

.PHONY: build
build: build-frontend build-backend

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

.PHONY: image
image: image
	docker build -t $(PROJECT_NAME):$(GIT_TAG) .

.PHONY: build-push-development
build-push-development:
	docker buildx create --use
	docker buildx build -t webdevops/$(PROJECT_NAME):development --platform linux/amd64,linux/arm,linux/arm64 --push .

#######################################
# quality checks
#######################################

.PHONY: check
check: vendor lint test

.PHONY: test
test:
	time go test ./...

.PHONY: lint
lint: $(GOLANGCI_LINT_BIN)
	time $(GOLANGCI_LINT_BIN) run --verbose --print-resources-usage

$(GOLANGCI_LINT_BIN):
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(FIRST_GOPATH)/bin

#######################################
# release assets
#######################################

RELEASE_ASSETS = \
	$(foreach GOARCH,amd64 arm64,\
	$(foreach GOOS,linux darwin windows,\
		release-assets/$(GOOS).$(GOARCH))) \

word-dot = $(word $2,$(subst ., ,$1))

.PHONY: release-assets
release-assets: clean-release-assets vendor $(RELEASE_ASSETS)

.PHONY: clean-release-assets
clean-release-assets:
	rm -rf ./release-assets
	mkdir -p ./release-assets

release-assets/windows.%: $(SOURCE)
	echo 'build release-assets for windows/$(call word-dot,$*,2)'
	GOOS=windows \
 	GOARCH=$(call word-dot,$*,1) \
	CGO_ENABLED=0 \
	time go build -ldflags '$(LDFLAGS)' -o './release-assets/$(PROJECT_NAME).windows.$(call word-dot,$*,1).exe' .

release-assets/%: $(SOURCE)
	echo 'build release-assets for $(call word-dot,$*,1)/$(call word-dot,$*,2)'
	GOOS=$(call word-dot,$*,1) \
 	GOARCH=$(call word-dot,$*,2) \
	CGO_ENABLED=0 \
	time go build -ldflags '$(LDFLAGS)' -o './release-assets/$(PROJECT_NAME).$(call word-dot,$*,1).$(call word-dot,$*,2)' .
