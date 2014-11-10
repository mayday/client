package main

import ( 
	"mayday/core" 
        "fmt"
)

func main() {
	config, err := core.NewConfig("./config.yaml")

	if err != nil {
		fmt.Printf("Error %s\n", err)
	}

	mayday, err := core.NewClient(config)
	mayday.Run()

}
