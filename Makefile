gofiles_novendor = $(shell find . -name vendor -prune -o -type f -name '*.go' -print)
# https://github.com/golang/lint#installationdocker
golint = $(shell go list -f {{.Target}} golang.org/x/lint/golint)
commit = $(shell git rev-parse --short=8 HEAD)

checks: fmt vet lint staticcheck errcheck test

test:
	go test ./...

vet:
	go vet ./...

lint:
	# golint install is borken on travisci
	$(golint) ./... || true

staticcheck:	
	staticcheck ./...

fmt:	
	@$(foreach gofile, $(gofiles_novendor),\
			out=$$(gofmt -s -l -d -e $(gofile) | tee /dev/stderr); if [ -n "$$out" ]; then exit 1; fi;)

errcheck:
	errcheck -verbose ./...

docker-build:
	docker build \
	    --build-arg commit="${commit}" \
	    -t pubsub-slack:${commit} .
