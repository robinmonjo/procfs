VERSION:=1.0
HARDWARE=$(shell uname -m)
GOPATH=`pwd`/cmd/gops/vendor

build:
	GOOS=linux GOPATH=$(GOPATH) bash -c 'cd ./cmd/gops && go build'

vendor:
	mkdir -p $(GOPATH)/src/github.com/robinmonjo
	ln -s `pwd` $(GOPATH)/src/github.com/robinmonjo
	GOPATH=$(GOPATH) go get github.com/codegangsta/cli

clean:
	rm -rf $(GOPATH)

test:
	GOPATH=$(GOPATH) go test -cover
