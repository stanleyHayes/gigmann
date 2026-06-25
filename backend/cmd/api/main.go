// Command api is the entrypoint for the Gigmann Executive Cockpit API.
package main

import (
	"log"

	"github.com/xcreativs/gigmann/internal/bootstrap"
)

func main() {
	if err := bootstrap.Run(); err != nil {
		log.Fatal(err)
	}
}
