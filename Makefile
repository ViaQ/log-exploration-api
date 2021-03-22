EXECUTABLE:=log-exploration-api
PACKAGE:=github.com/ViaQ/log-exploration-api
IMAGE_PUSH_REGISTRY:=docker://quay.io/openshift-logging/$(EXECUTABLE)
VERSION:=${shell git describe --tags --always}
BUILDTIME := ${shell date -u '+%Y-%m-%d_%H:%M:%S'}
LDFLAGS:= -s -w -X '${PACKAGE}/pkg/version.Version=${VERSION}' \
					-X '${PACKAGE}/pkg/version.BuildTime=${BUILDTIME}'
BUILD_DIR:=./bin

.PHONY: build test clean image image-publish
build: test
	mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux go build -ldflags "${LDFLAGS}" -o $(BUILD_DIR)/$(EXECUTABLE) cmd/apiserver/main.go

test:
	go test ./...

clean:
	rm -rf $(BUILD_DIR)/

image: build
	docker build . -t ${EXECUTABLE}:${VERSION}

image-publish: image
	docker push ${EXECUTABLE}:${VERSION} ${IMAGE_PUSH_REGISTRY}:${VERSION}

test-e2e:
	docker-compose up -d
	@sleep 5
	chmod +x test/e2e/populate_indices.sh
	test/e2e/populate_indices.sh
	go test -v test/e2e/*.go
	docker-compose down -v