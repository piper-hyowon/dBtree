package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	fmt.Println("Starting server...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
