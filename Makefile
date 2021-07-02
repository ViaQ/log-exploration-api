CONTAINER_ENGINE?=podman
EXECUTABLE:=log-exploration-api
PACKAGE:=github.com/ViaQ/log-exploration-api
IMAGE_PUSH_REGISTRY:=quay.io/openshift-logging/$(EXECUTABLE)
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
	go test ./pkg/... -coverprofile=covprofile

test-cover:
	go test ./pkg/... -coverprofile=coverage.out && go tool cover -html=coverage.out

clean:
	rm -rf $(BUILD_DIR)/

image: build
	$(CONTAINER_ENGINE) build . -t ${IMAGE_PUSH_REGISTRY}:${VERSION}

image-publish: image
	$(CONTAINER_ENGINE) push ${IMAGE_PUSH_REGISTRY}:${VERSION}

test-e2e:
	podman run -d --name elasticsearch -p 9200:9200 -p 9300:9300 -e "discovery.type=single-node" docker.elastic.co/elasticsearch/elasticsearch:7.13.0
	chmod +x test/e2e/populate_indices.sh
	test/e2e/populate_indices.sh
	go test -v test/e2e/*.go
	podman stop elasticsearch || true && podman rm elasticsearch || true
