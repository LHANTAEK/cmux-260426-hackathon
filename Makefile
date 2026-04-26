.PHONY: build test install demo clean

GOENV := GOCACHE=$(CURDIR)/.gocache GOMODCACHE=$(CURDIR)/.gomodcache

build:
	$(GOENV) go build -o bin/agentsail ./cmd/agentsail

test:
	$(GOENV) go test ./...

install: build
	mkdir -p $(HOME)/.local/bin
	cp bin/agentsail $(HOME)/.local/bin/agentsail
	chmod +x $(HOME)/.local/bin/agentsail

demo: build
	./bin/agentsail ci --customer finbank --target mock:support_agent_v12 --report --soft-exit
	./bin/agentsail ci --customer retailco --target mock:support_agent_v12 --report
	./bin/agentsail ci --customer acme-bank --target mock:support_agent_v12 --report --soft-exit

clean:
	rm -rf bin .agentsail .gocache .gomodcache
