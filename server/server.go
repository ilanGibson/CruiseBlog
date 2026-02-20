package server

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"CruiseBlog/types"
	"CruiseBlog/utils"
	"crypto/rand"
)

var userID types.Key

type Server struct {
	blog              []types.Post
	usernames         map[string]string
	uniqueUsers       atomic.Uint64
	blogMu            sync.Mutex
	lastServerRestart time.Time
	ipHashes          types.IpSlice
	Admin             types.Admin
}

func NewServer() *Server {
	hashes := utils.NewIpSlice()
	return &Server{blog: make([]types.Post, 0), usernames: make(map[string]string), lastServerRestart: time.Now(), ipHashes: *hashes}
}

func (s *Server) JoinServer(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		http.NotFound(w, req)
		return
	}

	cookie, err := req.Cookie("session")
	// aka if cookie does not exist
	if err != nil {
		// IpIsUnique returns true is ip in unique
		ip := strings.Split(req.RemoteAddr, ":")[0]
		if uniqueIP := utils.IpIsUnique(ip, &s.ipHashes); uniqueIP == true {
			s.uniqueUsers.Add(1)
			s.Admin.AdminChan <- 1
			utils.WriteIpHash(ip, &s.ipHashes)
		}

		var UUID [16]byte
		_, err := rand.Read(UUID[:])
		// default err handling for handler
		if err != nil {
			log.Println("uuid generation failed: %w", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		userUUID := hex.EncodeToString(UUID[:])
		s.blogMu.Lock()
		s.usernames[userUUID] = utils.GetRandValue()
		s.blogMu.Unlock()
		cookie := new(http.Cookie)
		cookie.Name = "session"
		cookie.Value = userUUID
		http.SetCookie(w, cookie)
		http.Redirect(w, req, "/home/about.html", http.StatusSeeOther)
	} else {
		// aka if cookie is not mapped to username in server
		// happens when user has correct cookie and server is taken down then restarted
		// user still has cookie in browser to pass initial check but their cookie
		// is not mapped to username because username map[string]string is not persistant
		// between server restarts
		s.blogMu.Lock()
		if _, ok := s.usernames[cookie.Value]; !ok {
			s.usernames[cookie.Value] = utils.GetRandValue()
		}
		s.blogMu.Unlock()
		s.LoadPosts()
		http.Redirect(w, req, "/home", http.StatusSeeOther)
	}
}

func (s *Server) AddPost(w http.ResponseWriter, req *http.Request) {
	s.blogMu.Lock()
	if _, ok := s.usernames[(req.Context().Value(userID)).(*http.Cookie).Value]; !ok {
		s.usernames[(req.Context().Value(userID)).(*http.Cookie).Value] = utils.GetRandValue()
	}
	s.blogMu.Unlock()

	// get post content
	body, _ := (io.ReadAll(req.Body))
	defer req.Body.Close()

	var content types.ClientRequest
	if err := json.Unmarshal(body, &content); err != nil {
		log.Println("unmarshal user post into content failed: %w", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	s.blogMu.Lock()
	// get userID cookie value
	username := s.usernames[(req.Context().Value(userID)).(*http.Cookie).Value]
	s.blogMu.Unlock()

	if utils.CleanPost(content.Content) {

		date := strings.Split(fmt.Sprint(time.Now()), ".")[0]
		newPost := types.Post{DateOfPost: date, Username: username, Content: content.Content}

		s.blogMu.Lock()
		// add post to []Post
		s.blog = append(s.blog, newPost)
		s.blogMu.Unlock()

		f, err := json.Marshal(newPost)
		if err != nil {
			log.Println("marshal newPost failed: %w", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if err = utils.WritePost(f); err != nil {
			log.Println("WritePost(newPost) failed: %w", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		_, err = w.Write(f)
		if err != nil {
			log.Println("write newPost to response failed: %w", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		s.Admin.AdminChan <- 1

	} else {
		w.WriteHeader(422)
	}
}

func (s *Server) GetPosts(w http.ResponseWriter, req *http.Request) {
	s.blogMu.Lock()
	if _, ok := s.usernames[(req.Context().Value(userID)).(*http.Cookie).Value]; !ok {
		s.usernames[(req.Context().Value(userID)).(*http.Cookie).Value] = utils.GetRandValue()
	}

	f, err := json.Marshal(s.blog)
	s.blogMu.Unlock()

	if err != nil {
		log.Println("marshal s.blog failed: %w", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	_, err = w.Write(f)
	if err != nil {
		log.Println("write s.blog to response failed: %w", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (s *Server) LoadPosts() {
	log.Println("reading from disk")
	var posts []types.Post
	var err error

	for i := range 3 {
		posts, err = utils.GetPostsFromDisk()
		if err != nil {
			log.Printf("load posts from disk failed %v", i)
			if i == 2 {
				log.Fatal("load posts from disk to memory failed 3 times")
			}
			time.Sleep(1 * time.Second)
			continue
		}
		break
	}

	s.blogMu.Lock()
	s.blog = posts
	s.blogMu.Unlock()

	go func() {
		ticker := time.NewTicker(3 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			log.Println("reading from disk")
			posts, err := utils.GetPostsFromDisk()
			if err != nil {
				log.Println("load posts from disk failed: %w", err)
				continue
			}

			s.blogMu.Lock()
			s.blog = posts
			s.blogMu.Unlock()
		}
	}()
}

func (s *Server) SetAdminCookie(w http.ResponseWriter, req *http.Request) {
	if s.Admin.HasKeyBeenUsed || s.Admin.IsKeyExpired {
		http.Redirect(w, req, "/", http.StatusUnauthorized)
	}

	var randCookieVal [16]byte
	_, err := rand.Read(randCookieVal[:])
	if err != nil {
		log.Println("adminCookie generation failed: %w", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	adminCookie := hex.EncodeToString(randCookieVal[:])
	cookie := new(http.Cookie)
	cookie.Name = "admin_session"
	cookie.Value = adminCookie
	s.Admin.HasKeyBeenUsed = true
	http.SetCookie(w, cookie)
	http.Redirect(w, req, "/", http.StatusSeeOther)
}

func (s *Server) ServerInfo(w http.ResponseWriter, req *http.Request) {
	ServerInfo := types.ServerInfo{UniqueUsers: s.uniqueUsers.Load(), LastServerRestart: s.lastServerRestart, ServerAge: time.Duration(time.Since(s.lastServerRestart).Seconds())}

	f, err := json.Marshal(ServerInfo)
	if err != nil {
		log.Println("marshal ServerInfo failed: %w", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	_, err = w.Write(f)
	if err != nil {
		log.Println("write newPost to response failed: %w", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (s *Server) SseHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// create a channel for client disconnection
	clientGone := req.Context().Done()

	rc := http.NewResponseController(w)

	for {
		select {
		case <-clientGone:
			fmt.Println("client disconnected")
			// close(clientChan)
			return

		case <-s.Admin.AdminChan:
			s.blogMu.Lock()
			msg := types.SseMsg{
				TotalUsers: int(s.uniqueUsers.Load()),
				TotalPosts: len(s.blog),
			}
			msgBytes, err := json.Marshal(msg)
			s.blogMu.Unlock()

			if err != nil {
				log.Println("marshal total users/posts failed: %w", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			ssePayload := fmt.Sprintf("data: %s\n\n", msgBytes)
			_, err = w.Write([]byte(ssePayload))
			if err != nil {
				log.Println("write total users/posts failed: %w", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			err = rc.Flush()
			if err != nil {
				log.Println("flush total users/posts failed: %w", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
		}
	}
}

func (s *Server) RequireAuthAdmin(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		_, err := req.Cookie("admin_session")
		if err != nil {
			log.Println("get admin_session cookie failed: %w", err)
			http.Redirect(w, req, "/", http.StatusUnauthorized)
			return
		}

		fmt.Println("err nil")
		next.ServeHTTP(w, req)
	}
}
func (s *Server) RequireAuthHome() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		username, err := req.Cookie("session")
		if err != nil {
			log.Println("get user auth cookie failed: %w", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(req.Context(), userID, username)
		switch req.Method {
		case "POST":
			s.AddPost(w, req.WithContext(ctx))
		case "GET":
			s.GetPosts(w, req.WithContext(ctx))
		}
	}
}
