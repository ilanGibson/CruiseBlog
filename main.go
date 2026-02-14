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

	http.HandleFunc("/api/posts", blogSrvr.RequireAuth())
	http.HandleFunc("/server", blogSrvr.ServerInfo)
	http.HandleFunc("/", blogSrvr.JoinServer)
	http.Handle("/home/", http.StripPrefix("/home", http.FileServer(http.Dir("./static"))))

	fmt.Println("running server...")
	http.ListenAndServe(":8090", nil)
}
