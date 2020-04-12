package main

import (
	"fmt"
	"github.com/zhshch2002/goribot"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "Hello goribot")
	})
	goribot.Log.Info("Benchmark Server Start")
	err := http.ListenAndServe(":1229", nil)
	if err != nil {
		log.Fatalln("ListenAndServe: ", err)
	}
}
