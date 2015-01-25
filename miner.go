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
	target, _ = hex.DecodeString(string(target))
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

func mine(target []byte, newTarget chan []byte, incomingTargets chan []byte) {
	iterations := 1
	currentTarget := target
	message := generateMessage()
	for {
		select {
		case t := <-incomingTargets:
			fmt.Println("received new target from parent", t)
			currentTarget = t
			fmt.Println("worker finished inc target")
		default:
			//fmt.Println("worker looping")
			iterations++
			hashAttempt := digest(message)
			if iterations > 1000000 {
				fmt.Println("completed 4mil attempts; asking parrent to read target; current message: ", message)
				fmt.Println("queue size", len(newTarget))
				newTarget <- make([]byte, 0, 0)
				//reloadTarget()
				iterations = 1
			}
			if bytes.Compare(hashAttempt, currentTarget) < 0 {
				newTarget <- currentTarget
				fmt.Println("congrats got a hash! submit target", hashAttempt, message)
				submitMessage(message)
				//reloadTarget()
			}
			message = hex.EncodeToString(hashAttempt)
			//fmt.Println("worker finished default")
		}
	}
}

func main() {
	fmt.Println("NUM CPUS", runtime.NumCPU())
	finished := make(chan bool)
	cpus := runtime.NumCPU()
	newTarget := make(chan []byte, 100)
	listeners := make([]chan []byte, 0)
	target := fetchTarget()

	for i := 0; i < cpus; i++ {
		incTargets := make(chan []byte)
		listeners = append(listeners, incTargets)
		go mine(target, newTarget, incTargets)
	}

	for {
		select {
		case t := <-newTarget:
			if len(t) > 0 {
				fmt.Println("new target received at main, should send it to workers", t)
				t = fetchTarget()
				for i, l := range listeners {
					fmt.Println("parent will pre-fetched target to listener", i)
					l <- t
				}
			} else {
				fmt.Println("received empty target; need to update ourselves")
				fmt.Println(len(listeners))
				fmt.Println(listeners)
				t = fetchTarget()
				for i, l := range listeners {
					fmt.Println("parent will send target to listener", i)
					l <- t
				}
			}
			fmt.Println("parent finished newTarget")
		default:
			//fmt.Println("main looping")
		}
	}
	foundHash := <-finished
	fmt.Println("Found hash", foundHash)
}
