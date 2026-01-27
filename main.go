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

type user struct {
}

func (c *connectionCounter) addConnection() {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	c.connections++
}

func (c *connectionCounter) removeConnection() {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	c.connections--
}

func hello(w http.ResponseWriter, req *http.Request) {
	cookie := new(http.Cookie)
	cookie.Name = "username"
	cookie.Value = "dilliontomphson"
	fmt.Printf("%+v\n", cookie)
	http.SetCookie(w, cookie)
	fmt.Fprint(w, "hello\n")
}

var c connectionCounter

func main() {
	// msgs := make(map[string]string)
	http.HandleFunc("/hel", hello)

	fmt.Println("running server...")
	http.ListenAndServe(":8090", nil)
}
