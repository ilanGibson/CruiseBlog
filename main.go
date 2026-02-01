package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	// "reflect"
)

type Key string

var userID Key

// slice[username]
// username is given to user via cookie
var Users []string

type Post struct {
	// todo date
	Username string `json:"username"`
	Content  string `json:"content"`
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

	tempContent := string(body)
	fmt.Println(tempContent)

	// get userID cookie value
	username := (req.Context().Value(userID)).(*http.Cookie).Value

	if CleanPost(string(tempContent)) {
		newPost := Post{Username: fmt.Sprint(username), Content: fmt.Sprint(tempContent)}
		// add post to []Post
		s.blog = append(s.blog, newPost)
		f, _ := json.Marshal(newPost)
		println(f)
		w.Write(f)
	} else {
		f, _ := json.Marshal("against cruise blog policy")
		w.Write(f)
	}
}

func requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		username, err := req.Cookie("username")
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(req.Context(), userID, username)
		next(w, req.WithContext(ctx))
	}
}

func main() {
	blogSrvr := &Server{blog: make([]Post, 0)}
	http.HandleFunc("/", blogSrvr.joinServer)

	http.Handle("/home/", http.StripPrefix("/home", http.FileServer(http.Dir("./static"))))
	http.HandleFunc("/api/posts", requireAuth(blogSrvr.addPost))
	fmt.Println("running server...")
	http.ListenAndServe(":8090", nil)
}
