.PHONY: all build-run build-backend build-backend run

all: build-frontend build-backend
build: build-frontend build-backend

build-run: build-frontend build-backend run

build-backend:
	#go-bindata ./templates/...
	go build

run:
	./devops-console

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
