build: 
	go build -ldflags "-s" -o mayday main.go
debug: 
	go build -o mayday main.go
