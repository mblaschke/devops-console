#############################################
# GET/CACHE NPM DEPS
#############################################
FROM node:alpine as npm-dependencies
WORKDIR /app
# get npm modules (cache)
COPY ./react/package.json /app/react/
COPY ./react/package-lock.json /app/react/
RUN set -x \
    && cd /app/react \
    && npm install

#############################################
# BUILD REACT APP
#############################################
FROM node:alpine as frontend
# Copy app and build
WORKDIR /app
RUN apk --no-cache add make
COPY ./ /app
COPY --from=npm-dependencies /app/react/node_modules/ /app/react/node_modules/
RUN set -x \
    && make build-frontend \
    && rm -rf /app/react

#############################################
# BUILD GO APP
#############################################
FROM golang:1.13 as backend
WORKDIR /go/src/devops-console
COPY ./ /go/src/devops-console
RUN go mod vendor
COPY --from=frontend /app/templates /go/src/devops-console/templates
COPY --from=frontend /app/static /go/src/devops-console/static
RUN set -x \
    && CGO_ENABLED=0 GOOS=linux go build -o devops-console
RUN ./devops-console --help

#############################################
# FINAL IMAGE
#############################################
FROM gcr.io/distroless/static

WORKDIR /app
COPY --from=backend /go/src/devops-console/ /app/
USER 65534
EXPOSE 9000
CMD ["/app/devops-console"]
