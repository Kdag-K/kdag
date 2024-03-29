BUILD_TAGS?=kdag

## vendor installs all the Go dependencies in vendor/
#vendor:
#	(rm -rf vendor ) && GO111MODULE=on go mod vendor

# install compiles and places the binary in GOPATH/bin
install:
	go install --ldflags '-extldflags "-static"' \
		--ldflags "-X github.com/Kdag-K/kdag/src/version.GitCommit=`git rev-parse HEAD`" \
		./cmd/kdag

# build compiles and places the binary in /build
build:
	CGO_ENABLED=0 go build \
		--ldflags "-X github.com/Kdag-K/kdag/src/version.GitCommit=`git rev-parse HEAD`" \
		-o build/kdag ./cmd/kdag/

# dist builds binaries for all platforms and packages them for distribution
dist:
	@BUILD_TAGS='$(BUILD_TAGS)' $(CURDIR)/scripts/dist.sh $(VERSION)

# dist builds aar for mobile android
mobile-dist:
	@BUILD_TAGS='$(BUILD_TAGS)' $(CURDIR)/scripts/dist_mobile.sh $(VERSION)

tests:  test

test:
	go test -count=1 -tags=unit ./...

flagtest:
	go test -count=1 -run TestFlagEmpty ./...

extratests:
	 go test -count=1 -run Extra ./...

alltests:
	go test -count=1 ./...
 
lint:
	golint ./...

doc:
	GO111MODULE=off godoc -play -goroot .

.PHONY: vendor install build dist test flagtest extratests alltests tests mobile-dist
