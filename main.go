package main

import (
	"context"
	"log"
	"zed-cli-win-unofficial/cmd"
)

func main() {
	if err := cmd.Execute(context.Background()); err != nil {
		log.Fatal(err)
	}
}
