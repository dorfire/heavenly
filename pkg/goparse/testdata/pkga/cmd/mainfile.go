package main

import (
	"example.com/pkga"
	"log"
	"net/http"
)

func main() {
	log.Print(pkga.DoSomething())
	log.Print(http.DetectContentType([]byte("<html></html>")))
}
