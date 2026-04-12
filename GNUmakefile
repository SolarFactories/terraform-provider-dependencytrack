default: fmt lint install generate

build:
	go build -v ./...

install: build
	go install -v ./...

lint:
	golangci-lint config verify
	golangci-lint run

generate:
	cd tools; go generate ./...

fmt:
	gofmt -s -w -e .

test:
	go test -v -cover -timeout=120s -parallel=10 -skip="^TestAcc" ./...

# Skip certain OIDC tests when testing locally, since we do not provide an OIDC IdP.
# These tests are covered using GitHub Actions OIDC ID Tokens in the pipeline.
# If wanting to test these locally, then you will be required to setup your own OIDC IdP
#	following https://docs.dependencytrack.org/getting-started/openidconnect-configuration/
testacc:
	TF_ACC=1 go test -v -cover -timeout 120m -skip='^(TestAccOidcAvailableDataSource)|(TestAccOidcLoginDataSource)$$' ./...

.PHONY: fmt lint test testacc build install generate
