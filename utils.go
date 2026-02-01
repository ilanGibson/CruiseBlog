package main

import (
	"fmt"
	"math/rand"
	"slices"
	"strings"
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
