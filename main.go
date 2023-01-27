package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func connect(s string) *http.Response {

	url := "https://id.twitch.tv/oauth2/authorize?response_type=token&client_id=zn2s0do5jo5r312vkl0yjhqth7fiy8&redirect_uri=http://localhost:3000&scope=channel%3Amanage%3Apolls+channel%3Aread%3Apolls&state=c3ab8aa609ea11e793ae92361f002671"
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal("Failed to connect to url: " + url)
	}

	return resp

}

func main() {

	resp := connect("Hello")
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Failed to parse the body")
	}

	oath := string(body)
	fmt.Print(oath)

}
