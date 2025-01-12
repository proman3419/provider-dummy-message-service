package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func main() {
	messageToDelete := Message{Id: 0}

	namespace := "dummy-message-service"
	serviceName := "dummy-message-service-svc"
	apiPort := 8000
	apiUrl := fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", serviceName, namespace, apiPort)

	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/message?id_=%d", apiUrl, messageToDelete.Id),
		bytes.NewBuffer([]byte(nil)))
	if err != nil {
		log.Printf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Printf("Failed to send DELETE /message: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to send DELETE /message: %v\n", err)
	}
	log.Printf("Response: %s\n", body)
}
