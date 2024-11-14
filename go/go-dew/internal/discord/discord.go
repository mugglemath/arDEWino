package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func SendDiscordMessage(message, webhookURL string) {
	data := map[string]string{"content": message}
	jsonData, _ := json.Marshal(data)

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Failed to send message to Discord: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 204 {
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Printf("Failed to send message to Discord: %d, %s\n", resp.StatusCode, string(body))
	}
}
