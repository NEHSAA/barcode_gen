TMPGOPATH := $(shell mktemp -d)
CMD_SERVE := env GOPATH=$(TMPGOPATH) dev_appserver.py app.yaml
CLEANUP := rm -rf $(TMPGOPATH)

local_serve:
	ln -s $(PWD)/vendor $(TMPGOPATH)/src
	# cp -r $(PWD)/vendor $(TMPGOPATH)/src
	bash -xc "trap '$(CLEANUP)' EXIT; $(CMD_SERVE)"
