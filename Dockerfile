#############################################
# BUILD REACT APP
#############################################
FROM --platform=$BUILDPLATFORM node:alpine as frontend

RUN apk upgrade --no-cache --force
RUN apk add --update build-base make git

# get npm modules (cache)
COPY ./react/package.json /app/react/
COPY ./react/package-lock.json /app/react/
RUN set -x \
    && cd /app/react \
    && npm install

# Copy app and build
WORKDIR /app
COPY . .
RUN set -x \
    && make build-frontend \
    && rm -rf /app/react \
    && touch /app/static/dist/.gitkeep

#############################################
# BUILD GO APP
#############################################
FROM --platform=$BUILDPLATFORM golang:1.18-alpine as backend

RUN apk upgrade --no-cache --force
RUN apk add --update build-base make git

WORKDIR /go/src/devops-console
COPY ./ /go/src/devops-console
RUN go mod vendor

RUN git status
RUN make test
ARG TARGETOS TARGETARCH
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} make build-backend

#############################################
# Test
#############################################
FROM gcr.io/distroless/static as test
USER 0:0
WORKDIR /app
COPY --from=backend /go/src/devops-console/devops-console .
COPY --from=backend /go/src/devops-console/config ./config
COPY --from=frontend /app/templates ./templates
COPY --from=frontend /app/static ./static
RUN ["./devops-console", "--help"]

#############################################
# FINAL IMAGE
#############################################
FROM gcr.io/distroless/static
ENV LOG_JSON=1
WORKDIR /app
COPY --from=test /app .
USER 1000:1000
EXPOSE 9000
ENTRYPOINT ["/app/devops-console"]
