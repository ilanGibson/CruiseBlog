package utils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"slices"
	"strings"

	"CruiseBlog/types"
)

func CleanPost(content string) bool {
	badWords := []string{"fuck", "shit", "ass", "bitch", "cunt", "whore"}
	// maxPostLen

	// againstLangPolicyFlag = check if content contains badWords
	if slices.Contains(badWords, content) {
		fmt.Println("here")
		return false
	}
	// againstLengthPolicyFlag = check if content within maxPostLen
	return true
}

func SavePost(contents types.Post) error {
	c, err := json.Marshal(contents)
	if err != nil {
		fmt.Println("cant marshal contents", err)
	}
	c = append(c, '\n')

	f, err := os.OpenFile("./blog.jsonl", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(c)
	return err
}

func GetPostsFromDisk() ([]types.Post, error) {
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
