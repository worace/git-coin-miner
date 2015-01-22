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

//const baseUrl string = "http://git-coin.herokuapp.com"
const baseUrl string = "http://localhost:9292"

func gcUrl(path string) string {
	return fmt.Sprintf("%s%s", baseUrl, path)
}

func generateMessage() string {
	return fmt.Sprintf("%d%v", rand.Int63(), time.Now())
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

func digest(input string) []byte {
	h := sha1.New()
	h.Write([]byte(input))
	return h.Sum(nil)
}

func mine(finished chan bool, seed int) {
	iterations := 1
	currentTarget := fetchTarget()
	targetBytes, _ := hex.DecodeString(string(currentTarget))
	reloadTarget := func() {
		currentTarget = fetchTarget()
		targetBytes, _ = hex.DecodeString(string(currentTarget))
		fmt.Println("target now: ", string(currentTarget))
		iterations = 1
	}
	message := string(seed)
	for {
		iterations++
		hashAttempt := digest(message)
		if iterations > 4000000 {
			fmt.Println("completed 4mil attempts; re-checking target; current message: ", message)
			fmt.Printf("digest: ", hex.EncodeToString(hashAttempt))
			reloadTarget()
			iterations = 1
		}
		if bytes.Compare(hashAttempt, targetBytes) < 0 {
			fmt.Println("congrats got a hash!")
			submitMessage(message)
			reloadTarget()
			//finished <- true
		}
		message = hex.EncodeToString(hashAttempt)
	}
}

func checkTargetPeriodically(target Target) {
	ticker := time.NewTicker(time.Millisecond * 3000)
	go func() {
		for range ticker.C {
			fmt.Println("fetching target")
			target.currentTarget = fetchTarget()
			targetBytes, _ := hex.DecodeString(string(target.currentTarget))
			target.targetBytes = targetBytes
			fmt.Println("target now ", target.currentTarget)
			fmt.Println("target bytes now ", target.targetBytes)
		}
	}()
}

type Target struct {
	currentTarget []byte
	targetBytes   []byte
}

func main() {
	fmt.Println("NUM CPUS", runtime.NumCPU())
	finished := make(chan bool)

	for i := 0; i < runtime.NumCPU(); i++ {
		go mine(finished, i)
	}
	foundHash := <-finished
	fmt.Println("Found hash", foundHash)
}
