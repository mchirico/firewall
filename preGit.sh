#!/bin/bash

#  git push --delete origin tagName
#  git tag -d tagName
#
# git tag -a v1.0 -m 'simple main test with flags'
# git checkout tags/v1.0

export GOMAXPROCS=8

gofmt -s -w utils/utils.go
gofmt -s -w utils/utils_test.go
gofmt -s -w watch/watch.go
gofmt -s -w watch/watch_test.go

gofmt -s -w firewall/*.go


go test  -v -coverprofile=c0.out -covermode=atomic github.com/mchirico/firewall/utils
go test -race -v  -coverprofile=c0.out -covermode=atomic github.com/mchirico/firewall/watch

go test -race -v  -coverprofile=c0.out -covermode=atomic github.com/mchirico/firewall/firewall

#go test -race -v -coverprofile=c1.out -covermode=atomic github.com/mchirico/gcpso/utils
sleep 3
go vet -v github.com/mchirico/firewall/utils # github.com/mchirico/gcpso/utils


# Mock just area
# mockgen -destination=./mocks/mock_stuff.go -package=mocks github.com/mchirico/gcpso/configs/mocks Area

