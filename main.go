package main

import (
	"fmt"

	"knative.dev/client/pkg/kn/commands"
)

func main() {
	// create a client
	p := commands.KnParams{}
	p.Initialize()
	client, _ := p.NewClient("default")
	service, _ := client.GetService("helloworld-go")

	fmt.Println(service.GetName())
}
