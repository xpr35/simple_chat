CURRENT_DIR = $(shell pwd)

all:
	go install src/server.go src/fifo.go
	go install src/client.go
	
getpath:
	export GOBIN=$(CURRENT_DIR)/bin
run-server:
	go run src/server.go src/fifo.go 6578
run-client:
	go run src/client.go localhost:6578 user
clean:
	rm -rf bin