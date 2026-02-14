package utils

import (
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"fmt"
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

func WritePost(contents types.Post) error {
	c, err := json.Marshal(contents)
	if err != nil {
		fmt.Println("cant marshal contents", err)
	}
	c = append(c, '\n')

	usernameFileMutex.Lock()
	defer usernameFileMutex.Unlock()
	f, err := os.OpenFile("./blog.jsonl", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(c)
	return err
}

func GetPostsFromDisk() ([]types.Post, error) {
	usernameFileMutex.Lock()
	defer usernameFileMutex.Unlock()

	file, err := os.Open("./blog.jsonl")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var returnPosts []types.Post
	for scanner.Scan() {
		var post types.Post
		if err := json.Unmarshal(scanner.Bytes(), &post); err != nil {
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
