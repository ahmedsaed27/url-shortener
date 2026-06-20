package main

import (
	"github.com/as9840935/url-shortener/internal/app"
	"log"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
