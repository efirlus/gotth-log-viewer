package main

import (
	"fmt"
	"gotthlogviewer/handlers"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handlers.Make(handlers.HandleRoot))

	fmt.Println("server listened 5174...")
	http.ListenAndServe(":5174", mux)
}
