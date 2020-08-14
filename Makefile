.PHONY: build

cnf ?= config.env
include $(cnf)
export $(shell sed 's/=.*//' $(cnf))

ifeq ($(strip $(DOCKER_REGISTRY)),)
	REGISTRY_TARGET = $(APP_NAME)
else
	REGISTRY_TARGET = $(DOCKER_REGISTRY)/$(APP_NAME)
endif

ifdef VERSION
	RELEASE_VERSION = $(VERSION)
endif

all: start

build:
	mkdir -p $(RELEASE_SERVER)
	@GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=0 go build -mod=vendor -o $(SERVER_BIN) .
	chmod a+x $(SERVER_BIN)
	rm -rf $(RELEASE_SERVER)/configs
	mkdir -p $(RELEASE_SERVER)/configs

start: build
	$(SERVER_BIN) -config $(RELEASE_SERVER)/configs/config.toml

clean:
	rm -rf release

pack: build
	cd $(RELEASE_SERVER) && tar zcvf ../$(APP_NAME).$(RELEASE_VERSION).tar.gz .
	rm -rf $(RELEASE_SERVER)

docker-build: pack
	cp -f Dockerfile Shanghai $(RELEASE_ROOT)
	cd $(RELEASE_ROOT) && docker build --no-cache -t $(REGISTRY_TARGET):$(RELEASE_VERSION) \
	--build-arg appname=$(APP_NAME) \
	--build-arg exposeport=$(SERVER_PORT) \
	--build-arg packagefile=$(APP_NAME).$(RELEASE_VERSION).tar.gz .

docker-clean:
	docker rm $(shell docker ps -f "status=exited" -q)
	docker rmi $(shell docker images -f "dangling=true" -q)

docker-push:
	docker push $(REGISTRY_TARGET):$(RELEASE_VERSION)

docker-all: docker-build docker-push
