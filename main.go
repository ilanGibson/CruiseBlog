package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"CruiseBlog/server"
	"CruiseBlog/utils"
)

// username is given to user via cookie
var Users []string

func main() {
	adminLink := flag.Bool("a", false, "flag to print link for admin cookie")
	port := flag.String("p", ":8090", "server port")
	flag.Parse()

	blogSrvr := server.NewServer()
	blogSrvr.LoadPosts()

	if *adminLink {
		blogSrvr.Admin.Key = utils.GetRandValue()
		blogSrvr.Admin.KeyExpireLength = 15 * time.Minute
		blogSrvr.Admin.AdminChan = make(chan int, 1)

		go func() {
			<-time.After(blogSrvr.Admin.KeyExpireLength)
			blogSrvr.Admin.IsKeyExpired = true
		}()

		path := fmt.Sprintf("/admin/%v", blogSrvr.Admin.Key)
		http.HandleFunc(path, blogSrvr.SetAdminCookie)
		http.HandleFunc("/admin/sseEvents", blogSrvr.RequireAuthAdmin(http.HandlerFunc(blogSrvr.SseHandler)))
		http.HandleFunc("/admin/", blogSrvr.RequireAuthAdmin(http.StripPrefix("/admin", http.FileServer(http.Dir("./static/admin/")))))
		fmt.Println(path)
	}

	http.HandleFunc("/api/posts", blogSrvr.RequireAuthHome())
	http.HandleFunc("/server", blogSrvr.ServerInfo)
	http.HandleFunc("/", blogSrvr.JoinServer)
	http.Handle("/home/", http.StripPrefix("/home", http.FileServer(http.Dir("./static/home/"))))

	log.Printf("running server... %v", *port)
	http.ListenAndServe(*port, nil)
}
