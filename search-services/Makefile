test:
	go test -coverprofile=cover.out ./...
	go tool cover -html=cover.out -o cover.html

lint: protolint golint

protobuf:
	protoc --go_out=. --go_opt=paths=source_relative \
               --go-grpc_out=. --go-grpc_opt=paths=source_relative \
               proto/words/words.proto
	protoc --go_out=. --go_opt=paths=source_relative \
               --go-grpc_out=. --go-grpc_opt=paths=source_relative \
               proto/update/update.proto
	protoc --go_out=. --go_opt=paths=source_relative \
               --go-grpc_out=. --go-grpc_opt=paths=source_relative \
               proto/search/search.proto

protolint:
	protolint .

golint:
	golangci-lint run -E gocritic -v ./...
	
test-cover:
	go test -race -coverprofile cover.out \
		$$(go list ./... | grep -E -v 'yadro.com/course/(proto|api$$|update$$|words$$|search$$)')
	go tool cover -html cover.out -o cover.html
	go tool cover -func cover.out | grep total | awk '{print "Total Coverage: " $$3}'