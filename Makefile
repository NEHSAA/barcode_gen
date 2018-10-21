TMPGOPATH := $(shell mktemp -d)
CMD_DEPLOY := env GOPATH=$(TMPGOPATH):${PWD} gcloud app deploy --project nehsaa-barcode-9487 appengine/app.yaml
CLEANUP := rm -rf $(TMPGOPATH)

all: build

.PHONY: vendor

vendor:
	env DEPPROJECTROOT=${PWD} dep ensure -v --vendor-only

build: vendor
	go build -o app.bin

local_serve: vendor
	env GOPATH=${PWD} dev_appserver.py app.yaml

deploy: vendor
	ln -s $(PWD)/vendor $(TMPGOPATH)/src
	bash -xc "trap '$(CLEANUP)' EXIT; $(CMD_DEPLOY)"
	# env GOPATH=${PWD} gcloud app deploy --project nehsaa-barcode-9487 appengine/app.yaml

run:
	env OAUTH_CLIENT_ID="981397958260-00h8bgke2po7f7emavesmr134asnqpk3.apps.googleusercontent.com" OAUTH_CLIENT_SECRET="-RCr-lhrCekEV-ufHcIHwsBz" ./barcode_gen

docker:
	docker build -t asia.gcr.io/nehsaa-infra/barcode_gen .

docker_deploy:
	docker push asia.gcr.io/nehsaa-infra/barcode_gen