package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

func main() {
	messageToCreate := Message{Content: "persist this"}

	namespace := "dummy-message-service"
	serviceName := "dummy-message-service-svc"
	apiPort := 8000
	apiUrl := fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", serviceName, namespace, apiPort)
	resp, err := http.Post(fmt.Sprintf("%s/message?content=%s", apiUrl, url.QueryEscape(messageToCreate.Content)),
		"application/json", nil)
	if err != nil {
		log.Printf("Failed to send POST /message: %v\n", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to send POST /message: %v\n", err)
	}
	log.Printf("Response: %s\n", body)

	var createdMessage Message
	err = json.Unmarshal([]byte(body), &createdMessage)
	if err != nil {
		log.Printf("Error parsing JSON: %v\n", err)
	}
	log.Printf("Created message: %+v\n", createdMessage)
}
