all: build

proto_v1:
	cd proto && \
		protoc \
		--go_out=zync/v1 --go_opt=paths=source_relative \
		--go-grpc_out=zync/v1 --go-grpc_opt=paths=source_relative \
		zync.proto

bin:
	mkdir bin

zync: bin proto_v1
	go build -o bin/zync cmd/zync/*.go

zyncd: bin proto_v1
	go build -o bin/zyncd cmd/zyncd/*.go

build: zync zyncd

install: build
	cp bin/zync /usr/local/bin/zync
	cp bin/zyncd /usr/local/bin/zyncd

start:
	./bin/zyncd start

stop:
ifneq (, $(shell ls | grep pid))
	kill $(shell cat ./zyncd.pid)
	rm -f zyncd.pid
endif

send-stop:
	./bin/zyncd stop

clean: stop
	rm -f zyncd.*
