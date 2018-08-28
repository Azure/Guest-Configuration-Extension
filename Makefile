BUNDLEDIR=bundle
BINDIR=bin
BIN=guest-configuration-extension
BUNDLE=guest-configuration-extension.zip

bundle: clean binary
	@mkdir -p $(BUNDLEDIR)
	zip ./$(BUNDLEDIR)/$(BUNDLE) ./$(BINDIR)/$(BIN)
	zip ./$(BUNDLEDIR)/$(BUNDLE) ./agent/DesiredStateConfiguration_*.zip
	zip ./$(BUNDLEDIR)/$(BUNDLE) ./$(BINDIR)/guest-configuration-shim
	zip -j ./$(BUNDLEDIR)/$(BUNDLE) ./misc/HandlerManifest.json
	zip -j ./$(BUNDLEDIR)/$(BUNDLE) ./misc/manifest.xml

build: binary

binary: clean
	if [ -z "$$GOPATH" ]; then \
		echo "GOPATH is not set"; \
		exit 1; \
	fi
	
	go get github.com/ahmetalpbalkan/govvv
	go get github.com/tools/godep

	$$GOPATH/bin/godep restore

	go list ./... | grep -v '/vendor/' | xargs go test -cover

	GOOS=linux GOARCH=amd64 govvv build -v \
		-ldflags "-X main.Version=`grep -E -m 1 -o  '<Version>(.*)</Version>' misc/manifest.xml | awk -F">" '{print $$2}' | awk -F"<" '{print $$1}'`" \
		-o $(BINDIR)/$(BIN) ./main
	cp ./misc/guest-configuration-shim ./$(BINDIR)

test: clean
	go list ./... | grep -v '/vendor/' | xargs go test -cover

sanity: clean prereqs
	golint ./main/ ./pkg/
	gofmt -w -s  ./main/ ./pkg/
	go vet -v ./main/
	go list ./... | grep -v '/vendor/' | xargs go test -cover

prereqs: clean
	go get golang.org/x/lint/golint

clean:
	rm -rf "$(BINDIR)" "$(BUNDLEDIR)"

.PHONY: clean binary test
