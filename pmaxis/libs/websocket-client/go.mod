module github.com/pmaxis/pmaxis/libs/websocket-client

go 1.24.0

require (
	github.com/gorilla/websocket v1.5.1
	github.com/pmaxis/pmaxis/libs/logger v0.0.0
)

require golang.org/x/net v0.47.0 // indirect

replace github.com/pmaxis/pmaxis/libs/logger => ../logger
