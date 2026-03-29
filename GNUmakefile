default: build

build:
	go build -o terraform-provider-clickstack

install: build
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/pleny-labs/clickstack/0.1.0/$$(go env GOOS)_$$(go env GOARCH)
	cp terraform-provider-clickstack ~/.terraform.d/plugins/registry.terraform.io/pleny-labs/clickstack/0.1.0/$$(go env GOOS)_$$(go env GOARCH)/

test:
	go test ./... -v

testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

generate:
	go generate ./...

docs:
	go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest generate

fmt:
	go fmt ./...

vet:
	go vet ./...

lint: fmt vet

.PHONY: build install test testacc generate docs fmt vet lint
