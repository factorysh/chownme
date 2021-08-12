package main

import (
	"fmt"
	"log"
	"os"

	"github.com/athoune/credrpc/client"
)

func main() {
	server := os.Getenv("SERVER")
	if server == "" {
		server = "/var/run/chownme/socket"
	}

	if len(os.Args) == 1 {
		fmt.Printf("%s path\n", os.Args[0])
		os.Exit(1)
	}
	cli := client.New(server)

	_, err := cli.Call([]byte(os.Args[1]))
	if err != nil {
		log.Fatal("Call error:", err)
	}
}
