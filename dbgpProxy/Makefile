WEBSITE_REPO=~/dev/php/xdebug-xdebug.org
BINARY_LOCATION=$(WEBSITE_REPO)/html/files/binaries
BINARIES=dbgpProxy dbgpProxy-macos dbgpProxy.exe

.NOTPARALLEL:

.PHONY: force

all: $(BINARIES)

dbgpProxy: dbgpProxy-macos force
	GOOS=linux GOARCH=amd64 go build -tags NO_CLOUD

dbgpProxy-macos: force
	GOOS=darwin GOARCH=amd64 go build -tags NO_CLOUD
	mv dbgpProxy dbgpProxy-macos

dbgpProxy.exe: force
	GOOS=windows GOARCH=amd64 go build -tags NO_CLOUD

dbgpProxy-dev: force
	GOOS=linux GOARCH=amd64 go build

deploy: all force
	$(eval VERSION := $(shell cat main.go | grep "var clientVersion" | sed 's/.*= "//' | sed 's/"//'))
	-git tag -s $(VERSION) -m "Go with $(VERSION)"
	cd $(WEBSITE_REPO) && git checkout master && git pull
	for i in $(BINARIES); do cp $$i $(BINARY_LOCATION); done
	cd $(BINARY_LOCATION) && git add $(BINARIES) && git commit -m "Added dbgpProxy binaries version $(VERSION)"
	cd $(WEBSITE_REPO) && git push
