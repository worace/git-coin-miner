package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
)

const baseUrl string = "http://git-coin.herokuapp.com"

func gcUrl(path string) string {
	return fmt.Sprintf("%s%s", baseUrl, path)
}

func generateMessage() string {
	return fmt.Sprintf("%d", rand.Int())
}

func digest(input string) []byte {
	h := sha1.New()
	h.Write([]byte(input))
	return h.Sum(nil)
}

func fetchTarget() []byte {
	resp, err := http.Get(gcUrl("/target"))
	if err != nil {
		panic("Couldn't fetch target")
	}

	target, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic("error reading response body")
	}
	defer resp.Body.Close()
	return target
}

func submitMessage(message string) {
	resp, err := http.PostForm(gcUrl("/hash"), url.Values{"owner": {"worace"}, "message": {message}})
	if err != nil {
		panic("failed to submit guess")
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic("error reading response body")
	}
	fmt.Println(string(body))
}

func main() {
	iterations := 1
	currentTarget := fetchTarget()
	targetBytes, _ := hex.DecodeString(string(currentTarget))
	reloadTarget := func() {
		currentTarget = fetchTarget()
		targetBytes, _ = hex.DecodeString(string(currentTarget))
		iterations = 1
	}
	for {
		iterations++
		message := generateMessage()
		hashAttempt := digest(message)
		if iterations > 1000000 {
			fmt.Println("completed 1mil attempts; re-checking target")
			reloadTarget()
		}
		if bytes.Compare(hashAttempt, targetBytes) < 0 {
			fmt.Println("congrats got a hash!")
			submitMessage(message)
			reloadTarget()
		}
	}
}
