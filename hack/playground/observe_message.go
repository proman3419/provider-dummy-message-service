package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type MessageObservation struct {
	Content string `json:"content"`
}

type MessageResponse struct {
	Messages []MessageObservation `json:"messages"`
}

func main() {
	messageToObserve := MessageObservation{Content: "persist this"}

	// 1. Call GET /messages
	namespace := "dummy-message-service"
	serviceName := "dummy-message-service-svc"
	apiPort := 8000
	apiUrl := fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", serviceName, namespace, apiPort)
	resp, err := http.Get(fmt.Sprintf("%s/messages", apiUrl))
	if err != nil {
		log.Printf("Failed to send GET /messages: %v", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to send GET /messages: %v", err)
	}

	// 2. Parse the JSON response
	var messageResponse MessageResponse
	err = json.Unmarshal(body, &messageResponse)
	if err != nil {
		log.Printf("Failed to parse JSON: %v", err)
		return
	}

	// 3. Iterate through the messages and print them
	for _, message := range messageResponse.Messages {
		if message.Content == messageToObserve.Content {
			log.Printf("Observed a message with content: '%s'", messageToObserve.Content)
			return
		}
	}
	log.Printf("Didn't observe a message with content: '%s'", messageToObserve.Content)
}
