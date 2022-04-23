GIT_COMMIT := $(shell git rev-list -1 HEAD)
DT := $(shell date +%Y.%m.%d.%H%M%S)
ME := $(shell whoami)
HOST := $(shell hostname)
PRODUCT := compass
PRODUCT_TAG := arm64
MAIN := ./cmd/main.go

run:
	DSN="host=localhost user=xxx password=xxx dbname=xxx port=5432 sslmode=disable TimeZone=America/Sao_Paulo" \
	REDIS=redis://localhost:6379 \
	CGO_ENABLED=0 go run -ldflags "-X github.com/digitalcircle-com-br/buildinfo.Ver=$(GIT_COMMIT) -X github.com/digitalcircle-com-br/buildinfo.BuildDate=$(DT) -X github.com/digitalcircle-com-br/buildinfo.BuildUser=$(ME) -X github.com/digitalcircle-com-br/buildinfo.BuildHost=$(HOST) -X github.com/digitalcircle-com-br/buildinfo.Product=$(PRODUCT)" $(MAIN)

docker:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o deploy/api -ldflags "-X github.com/digitalcircle-com-br/buildinfo.Ver=$(GIT_COMMIT) -X github.com/digitalcircle-com-br/buildinfo.BuildDate=$(DT) -X github.com/digitalcircle-com-br/buildinfo.BuildUser=$(ME) -X github.com/digitalcircle-com-br/buildinfo.BuildHost=$(HOST) -X github.com/digitalcircle-com-br/buildinfo.Product=$(PRODUCT)" $(MAIN)
	cd deploy && \
	docker build -t $(PRODUCT):$(PRODUCT_TAG) .

docker_run:
	docker run --rm -it -p 8080:8080 digitalcircle/$(PRODUCT):$(PRODUCT_TAG)

docker_amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o deploy/compass -ldflags "-s -w -X github.com/digitalcircle-com-br/buildinfo.Ver=$(GIT_COMMIT) -X github.com/digitalcircle-com-br/buildinfo.BuildDate=$(DT) -X github.com/digitalcircle-com-br/buildinfo.BuildUser=$(ME) -X github.com/digitalcircle-com-br/buildinfo.BuildHost=$(HOST) -X github.com/digitalcircle-com-br/buildinfo.Product=$(PRODUCT)" $(MAIN)
	cd deploy && \
	docker build -t digitalcircle/$(PRODUCT):amd64 .
	

# docker_arm64:
# 	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o deploy/compass -ldflags "-s -w -X github.com/digitalcircle-com-br/buildinfo.Ver=$(GIT_COMMIT) -X github.com/digitalcircle-com-br/buildinfo.BuildDate=$(DT) -X github.com/digitalcircle-com-br/buildinfo.BuildUser=$(ME) -X github.com/digitalcircle-com-br/buildinfo.BuildHost=$(HOST) -X github.com/digitalcircle-com-br/buildinfo.Product=$(PRODUCT)" $(MAIN)
# 	cd deploy && \
# 	docker build -t digitalcircle/$(PRODUCT):arm64 .
	
	
docker_local: docker_amd64
docker_only_push:
	docker push digitalcircle/$(PRODUCT):amd64
#docker push digitalcircle/$(PRODUCT):arm64
docker_push: docker_local docker_only_push

pubcfg:
	curl -v -X POST --data-binary @config.yaml http://localhost:10001/compass.yaml

reload:
	redis-cli -x set compass < config.yaml