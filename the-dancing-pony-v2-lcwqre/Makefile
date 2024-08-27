# Define the binary output location
BINARY = bin/thedancingpony

# Ensure bin directory exists
bin:
	@mkdir bin

# Target to build the project
build: bin
	@go build -buildvcs=false -o $(BINARY)

# Target to run the project
run: build
	@./$(BINARY)

# Target to run tests
test:
	@go test -v ./...
