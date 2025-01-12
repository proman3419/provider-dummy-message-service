package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func main() {
	messageToObserve := Message{Content: "persist this"}

	namespace := "dummy-message-service"
	serviceName := "dummy-message-service-svc"
	apiPort := 8000
	apiUrl := fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", serviceName, namespace, apiPort)
	resp, err := http.Get(fmt.Sprintf("%s/messages", apiUrl))
	if err != nil {
		log.Printf("Failed to send GET /messages: %v\n", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to send GET /messages: %v\n", err)
	}

	var messages Messages
	err = json.Unmarshal(body, &messages)
	if err != nil {
		log.Printf("Failed to parse JSON: %v\n", err)
		return
	}

	for _, message := range messages.Messages {
		if message.Content == messageToObserve.Content {
			log.Printf("Observed a message with content: '%s'\n", messageToObserve.Content)
			return
		}
	}
	log.Printf("Didn't observe a message with content: '%s'\n", messageToObserve.Content)
}
