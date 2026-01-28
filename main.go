package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"strings"
)

type Post struct {
	username string
	// todo add title
	content string
}

type Server struct {
	blog []Post
}

func (s *Server) joinServer(w http.ResponseWriter, req *http.Request) {
	_, err := req.Cookie("username")
	if err != nil {
		cookie := new(http.Cookie)
		cookie.Name = "username"
		cookie.Value = getRandValue()
		http.SetCookie(w, cookie)
	}

	http.Redirect(w, req, "/home", http.StatusSeeOther)
}

func (s *Server) addPost(w http.ResponseWriter, req *http.Request) {
	existingCookie, err := req.Cookie("username")
	if err == nil {
		s.blog = append(s.blog, Post{username: existingCookie.Value, content: "yes"})
		fmt.Printf("added: %v=yes", existingCookie.Value)
		fmt.Fprint(w, s.blog)
	} else {
		fmt.Fprint(w, "missing cookie")
	}
}

func getRandValue() string {
	var characters = []rune("ABCDEFG0123456789")
	var sb strings.Builder

	for range 8 {
		randomIndex := rand.Intn(len(characters))
		randomChar := characters[randomIndex]
		sb.WriteRune(randomChar)
	}

	return sb.String()
}

func main() {
	blogSrvr := &Server{blog: make([]Post, 0)}

	http.Handle("/home/", http.StripPrefix("/home", http.FileServer(http.Dir("./static"))))
	http.HandleFunc("/join", blogSrvr.joinServer)
	http.HandleFunc("/add", blogSrvr.addPost)
	fmt.Println("running server...")
	http.ListenAndServe(":8090", nil)
}
