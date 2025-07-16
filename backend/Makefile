
run:
	@CGO_ENABLED=1 go run . --config config.json

listen: 
	@CGO_ENABLED=1 go test -v ./tests/... -count=1 -run ^TestListenSSE$

test: 
	@CGO_ENABLED=1 go test -v ./tests/... -count=1 -run ^${t}$