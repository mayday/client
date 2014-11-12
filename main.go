package main

import (
	"fmt"
	"mayday/core"
)

func main() {
	config, err := core.NewConfig("./config.yaml", "./config.yaml.sig")

	if err != nil {
		fmt.Printf("Error %s\n", err)
		return
	}

	mayday, err := core.NewClient(config)
	mayday.Run()

}
