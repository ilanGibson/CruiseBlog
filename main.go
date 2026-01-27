package main

import (
	"fmt"
	"net/http"
	"sync"
)

type connectionCounter struct {
	connections int
	sync.Mutex
}

func hello(w http.ResponseWriter, req *http.Request) {
	fmt.Fprint(w, "hello\n")
}

func main() {
	c := &connectionCounter{connections: 0}

	http.HandleFunc("/hel", hello)

	http.ListenAndServe(":8090", nil)
}
