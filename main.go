package main

import (
	"fmt"
	"net/http"
	// "reflect"

	"CruiseBlog/server"
)

// username is given to user via cookie
var Users []string

func main() {
	blogSrvr := server.NewServer()
	blogSrvr.LoadPosts()

	http.HandleFunc("/api/posts", server.RequireAuth(blogSrvr))
	http.Handle("/home/", http.StripPrefix("/home", http.FileServer(http.Dir("./static"))))
	http.HandleFunc("/", blogSrvr.JoinServer)
	fmt.Println("running server...")
	http.ListenAndServe(":8090", nil)
}
