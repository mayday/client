server: 
	go build -ldflags "-s" -o mayday-server server/main.go
client: 
	go build -ldflags "-s" -o mayday main.go
debug: 
	go build -o mayday main.go
