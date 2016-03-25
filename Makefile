all: clean ca
	docker run \
	-v $(PWD)/:/go/src/bitlox/ \
	-e APP=$(APP) \
	-it --rm $(BUILDER_IMAGE)

builder-container: clean
	docker build -f build/Dockerfile.builder -t bitlox-builder build

clean:
	rm -rf build/bitlox build/ca-certificates.crt

ca:
	[ -f build/ca-certificates.crt ] || cp /etc/ssl/certs/ca-certificates.crt build

binary:
	CGO_CPPFLAGS="-m32 -I/usr/include" \
	CGO_LDFLAGS="-L/usr/lib -L/usr/lib/x86_64-linux-gnu -lzmq -lpthread -lusb-1.0 -lrt -lstdc++ -lm -lc -lgcc" \
	go get --ldflags '-extldflags "-static"' -a bitlox
	cp $$GOPATH/bin/$(APP) build/bitlox
