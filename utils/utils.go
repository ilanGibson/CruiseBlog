package utils

import (
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"slices"
	"strings"
	"sync"

	"CruiseBlog/types"
)

var usernameFileMutex sync.Mutex
var ipMutex sync.Mutex

func CleanPost(content string) bool {
	badWords := []string{"fuck", "shit", "ass", "bitch", "cunt", "whore"}

	// againstLangPolicyFlag = check if content contains badWords
	for _, word := range badWords {
		if strings.Contains(content, word) {
			return false
		}
	}
	return true
}

func WritePost(post []byte) error {
	post = append(post, '\n')

	usernameFileMutex.Lock()
	defer usernameFileMutex.Unlock()

	file, err := os.OpenFile("./blog.jsonl", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("open file blog.jsonl failed: %w", err)
		return err
	}
	defer file.Close()

	_, err = file.Write(post)
	if err != nil {
		log.Println("write new post to blog.jsonl for write failed: %w", err)
		return err
	}
	return nil
}

func GetPostsFromDisk() ([]types.Post, error) {
	usernameFileMutex.Lock()
	defer usernameFileMutex.Unlock()

	file, err := os.Open("./blog.jsonl")
	if err != nil {
		log.Println("open file blog.jsonl for read failed: %w", err)
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var returnPosts []types.Post
	for scanner.Scan() {
		var post types.Post
		if err := json.Unmarshal(scanner.Bytes(), &post); err != nil {
			log.Println("unmarshal user post into content failed: %w", err)
			return nil, err
		}
		returnPosts = append(returnPosts, post)
	}
	return returnPosts, nil
}

func GetRandValue() string {
	var characters = []rune("ABCDEFG0123456789")
	var sb strings.Builder

	for range 8 {
		randomIndex := rand.Intn(len(characters))
		randomChar := characters[randomIndex]
		sb.WriteRune(randomChar)
	}

	return sb.String()
}

func NewIpSlice() *types.IpSlice {
	return &types.IpSlice{IpHashes: make([]string, 0)}
}

func IpIsUnique(ip string, serverHashes *types.IpSlice) bool {
	ipInQuestion := hashIp(ip)

	if slices.Contains(serverHashes.IpHashes, ipInQuestion) {
		return false
	}

	return true
}

func hashIp(ip string) string {
	hash := sha256.New()
	hash.Write([]byte(ip))

	ipHash := hash.Sum(nil)
	return string(ipHash)
}

func WriteIpHash(ip string, serverHashes *types.IpSlice) {
	ipMutex.Lock()
	defer ipMutex.Unlock()

	serverHashes.IpHashes = append(serverHashes.IpHashes, hashIp(ip))
}
