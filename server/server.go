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

type broadcaster struct {
	mu             sync.Mutex
	subscriberLine chan types.Event
}

func newBroadcaster() *broadcaster {
	return &broadcaster{}
}

func (b *broadcaster) register() chan types.Event {
	b.mu.Lock()
	defer b.mu.Unlock()

	ch := make(chan types.Event)
	b.subscriberLine = ch

	return ch
}

func (b *broadcaster) unregister() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.subscriberLine != nil {
		close(b.subscriberLine)
		b.subscriberLine = nil
	}
}

func (b *broadcaster) publish(evt types.Event) {
	b.mu.Lock()
	ch := b.subscriberLine
	b.mu.Unlock()

	if ch == nil {
		// (note) no subscriber
		fmt.Println("no subs to chan")
		return
	}
	select {
	case ch <- evt:
		log.Println("event sent to subscriber")
	default:
		return
	}
}

type Server struct {
	blog              []types.Post
	usernames         map[string]string
	uniqueUsers       atomic.Uint64
	blogMu            sync.Mutex
	lastServerRestart time.Time
	ipHashes          types.IpSlice
	Admin             types.Admin
	broadcaster       *broadcaster
}

func NewServer() *Server {
	return &Server{blog: make([]types.Post, 0), usernames: make(map[string]string), lastServerRestart: time.Now(), ipHashes: *utils.NewIpSlice(),
		broadcaster: newBroadcaster()}
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
		s.uniqueUsers.Add(1)
		s.broadcaster.publish(types.Event{
			TotalUsers: int(s.uniqueUsers.Load()),
			TotalPosts: len(s.blog),
		})
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
			s.uniqueUsers.Add(1)
			s.broadcaster.publish(types.Event{
				TotalUsers: int(s.uniqueUsers.Load()),
				TotalPosts: len(s.blog),
			})
		}
		s.blogMu.Unlock()
		http.Redirect(w, req, "/home", http.StatusSeeOther)
	}
}

func (s *Server) addPost(w http.ResponseWriter, req *http.Request) {
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
		utils.SetPostUsernameDateHash(&newPost)

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
		s.broadcaster.publish(types.Event{
			TotalUsers: int(s.uniqueUsers.Load()),
			TotalPosts: len(s.blog),
		})

	} else {
		// (note) unprocessable entity
		w.WriteHeader(422)
	}
}

func (s *Server) getPosts(w http.ResponseWriter, req *http.Request) {
	s.blogMu.Lock()
	if _, ok := s.usernames[(req.Context().Value(userID)).(*http.Cookie).Value]; !ok {
		s.usernames[(req.Context().Value(userID)).(*http.Cookie).Value] = utils.GetRandValue()
		s.uniqueUsers.Add(1)
		s.broadcaster.publish(types.Event{
			TotalUsers: int(s.uniqueUsers.Load()),
			TotalPosts: len(s.blog),
		})
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

func (s *Server) LoadPosts(callFromServerStartUp bool) {
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

	if callFromServerStartUp {
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
}

func (s *Server) UpdatePost(w http.ResponseWriter, req *http.Request) {
	body, err := (io.ReadAll(req.Body))
	if err != nil {
		log.Println("reading request body for update failed: %w", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer req.Body.Close()

	var postUpdate types.Post
	if err := json.Unmarshal(body, &postUpdate); err != nil {
		log.Println("unmarshal post update into post failed: %w", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	utils.SetPostUsernameDateHash(&postUpdate)

	posts, err := utils.GetPostsFromDisk()
	if err != nil {
		log.Println("fetch posts for update failed: %w", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var newPosts []types.Post
	for _, post := range posts {
		if post.PostId == postUpdate.PostId {
			newPosts = append(newPosts, postUpdate)
			continue
		}
		newPosts = append(newPosts, post)
	}

	if err = utils.RewriteBlogDisk(newPosts); err != nil {
		log.Println("rewrite update amongst all posts failed: %w", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	s.LoadPosts(false)
}

func (s *Server) DeletePost(w http.ResponseWriter, req *http.Request) {
	body, err := (io.ReadAll(req.Body))
	if err != nil {
		log.Println("reading request body for deletion failed: %w", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer req.Body.Close()

	var postDeletion types.Post
	if err := json.Unmarshal(body, &postDeletion); err != nil {
		log.Println("unmarshal post deletion into post failed: %w", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	utils.SetPostUsernameDateHash(&postDeletion)

	posts, err := utils.GetPostsFromDisk()
	if err != nil {
		log.Println("fetch posts for deletion failed: %w", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var newPosts []types.Post
	// (note) deletion operation
	for _, post := range posts {
		if post.PostId == postDeletion.PostId {
			continue
		}
		newPosts = append(newPosts, post)
	}

	if err = utils.RewriteBlogDisk(newPosts); err != nil {
		log.Println("rewrite all posts minus deletion failed: %w", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	s.blogMu.Lock()
	s.blog = newPosts
	s.blogMu.Unlock()

	s.broadcaster.publish(types.Event{
		TotalUsers: int(s.uniqueUsers.Load()),
		TotalPosts: len(s.blog),
	})
}

func (s *Server) SetAdminCookie(w http.ResponseWriter, req *http.Request) {
	if s.Admin.HasPathBeenUsed || s.Admin.IsPathExpired {
		http.Redirect(w, req, "/", http.StatusUnauthorized)
		return
	}

	var randCookieVal [16]byte
	_, err := rand.Read(randCookieVal[:])
	if err != nil {
		log.Println("adminCookie generation failed: %w", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	s.Admin.HasPathBeenUsed = true

	adminCookie := hex.EncodeToString(randCookieVal[:])
	cookie := new(http.Cookie)
	cookie.Name = "admin_session"
	cookie.Value = adminCookie
	s.Admin.Key = adminCookie
	http.SetCookie(w, cookie)
	http.Redirect(w, req, "/", http.StatusSeeOther)
}

func (s *Server) SseHandler(w http.ResponseWriter, req *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming Unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch := s.broadcaster.register()
	defer s.broadcaster.unregister()

	s.blogMu.Lock()
	snapShot := types.SseMsg{
		TotalUsers: int(s.uniqueUsers.Load()),
		TotalPosts: len(s.blog),
	}
	s.blogMu.Unlock()

	snapShotBytes, err := json.Marshal(snapShot)
	if err != nil {
		log.Println("marshal snapShot failed: %w", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

	fmt.Fprintf(w, "data: %s\n\n", snapShotBytes)
	flusher.Flush()
	log.Println("snapShot sent")

	// create a channel for client disconnection
	clientGone := req.Context()

	for {
		select {
		case event, ok := <-ch:
			if !ok {
				return
			}

			msg := types.SseMsg{
				TotalUsers: event.TotalUsers,
				TotalPosts: event.TotalPosts,
			}
			msgBytes, err := json.Marshal(msg)
			if err != nil {
				log.Println("marshal total users/posts failed: %w", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			fmt.Fprintf(w, "data: %s\n\n", msgBytes)
			flusher.Flush()
			log.Println("sse event sent")

		case <-clientGone.Done():
			fmt.Println("client disconnected")
			return
		}
	}
}

func (s *Server) RequireAuthAdmin(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		adminUsername, err := req.Cookie("admin_session")
		if err != nil {
			log.Println("get admin_session cookie failed: %w", err)
			http.Redirect(w, req, "/", http.StatusUnauthorized)
			return
		}

		t := adminUsername.Value
		if t == s.Admin.Key {
			next.ServeHTTP(w, req)
			return
		}

		log.Println("admin auth cookie failed")
		http.Redirect(w, req, "/", http.StatusUnauthorized)
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

		fmt.Println("testing rquireauth calls")
		ctx := context.WithValue(req.Context(), userID, username)
		switch req.Method {
		case "POST":
			fmt.Println("testing POST calls")
			s.addPost(w, req.WithContext(ctx))
		case "GET":
			fmt.Println("testing GET calls")
			s.getPosts(w, req.WithContext(ctx))
		}
	}
}
