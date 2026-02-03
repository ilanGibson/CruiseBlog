package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
	// "reflect"
)

type Key string

var userID Key

// slice[username]
// username is given to user via cookie
var Users []string

type Request struct {
	Content string `json:"content"`
}

type Post struct {
	DateOfPost string `json:"date"`
	Username   string `json:"username"`
	Content    string `json:"content"`
}

type Server struct {
	blog []Post
}

func (s *Server) joinServer(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		http.NotFound(w, req)
		return
	}

	_, err := req.Cookie("username")
	// aka if cookie does not exist
	if err != nil {
		cookie := new(http.Cookie)
		cookie.Name = "username"
		tempRandomVal := GetRandValue()
		cookie.Value = tempRandomVal
		http.SetCookie(w, cookie)
		http.Redirect(w, req, "/home/about.html", http.StatusSeeOther)
	} else {
		http.Redirect(w, req, "/home", http.StatusSeeOther)
	}
}

func (s *Server) addPost(w http.ResponseWriter, req *http.Request) {
	// get post content
	body, _ := (io.ReadAll(req.Body))
	defer req.Body.Close()

	var content Request
	err := json.Unmarshal(body, &content)
	if err != nil {
		log.Fatal(err)
	}

	// get userID cookie value
	username := (req.Context().Value(userID)).(*http.Cookie).Value

	// if CleanPost(string(tempContent)) {
	if CleanPost("") {
		date := time.Now()
		newPost := Post{DateOfPost: fmt.Sprint(date), Username: fmt.Sprint(username), Content: fmt.Sprint(content.Content)}
		// add post to []Post
		s.blog = append(s.blog, newPost)
		f, _ := json.Marshal(newPost)
		w.Write(f)
	} else {
		f, _ := json.Marshal("against cruise blog policy")
		w.Write(f)
	}
}

func (s *Server) getPosts(w http.ResponseWriter, _ *http.Request) {
	f, err := json.Marshal(s.blog)
	if err != nil {
		log.Fatal("get post err", err)
	}
	_, err = w.Write(f)
	if err != nil {
		fmt.Println("get posts err", err)
	}
}

func requireAuth(s *Server) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		username, err := req.Cookie("username")
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		switch req.Method {
		case "POST":
			ctx := context.WithValue(req.Context(), userID, username)
			s.addPost(w, req.WithContext(ctx))
		case "GET":
			s.getPosts(w, req)
		}
	}
}

func main() {
	blogSrvr := &Server{blog: make([]Post, 0)}
	http.HandleFunc("/", blogSrvr.joinServer)

	http.Handle("/home/", http.StripPrefix("/home", http.FileServer(http.Dir("./static"))))
	http.HandleFunc("/api/posts", requireAuth(blogSrvr))
	fmt.Println("running server...")
	http.ListenAndServe(":8090", nil)
}
