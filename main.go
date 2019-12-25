package main

import (
	"fmt"
	"github.com/nandiheath/graphql-gateway/internal/config"
	"github.com/nandiheath/graphql-gateway/internal/server"
)

func main() {
	fmt.Println("started")
	config.Init()
	server.NewServer(nil).Start()
}