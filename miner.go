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
	"runtime"
	"time"
)

const baseUrl string = "http://git-coin.herokuapp.com"

func gcUrl(path string) string {
	return fmt.Sprintf("%s%s", baseUrl, path)
}

func generateMessage() string {
	return fmt.Sprintf("%d%v", rand.Int63(), time.Now())
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

func mine(finished chan bool) {
	iterations := 1
	currentTarget := fetchTarget()
	targetBytes, _ := hex.DecodeString(string(currentTarget))
	reloadTarget := func() {
		currentTarget = fetchTarget()
		targetBytes, _ = hex.DecodeString(string(currentTarget))
		fmt.Println("target now: ", string(currentTarget))
		iterations = 1
	}
	for {
		iterations++
		message := generateMessage()
		hashAttempt := digest(message)
		if iterations > 4000000 {
			fmt.Println("completed 4mil attempts; re-checking target")
			reloadTarget()
		}
		if bytes.Compare(hashAttempt, targetBytes) < 0 {
			fmt.Println("congrats got a hash!")
			submitMessage(message)
			reloadTarget()
			//finished <- true
		}
	}
}

func main() {
	fmt.Println("NUM CPUS", runtime.NumCPU())
	finished := make(chan bool)

	fmt.Println()

	for i := 0; i < runtime.NumCPU(); i++ {
		go mine(finished)
	}
	foundHash := <-finished
	fmt.Println("Found hash", foundHash)
}
