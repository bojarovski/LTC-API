package FunctionsHelper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

// CallAIService sends a request to the OpenAI API and returns the response.
func CallAIService(question string, responseLength int, gptRple string) (string, error) {
	apiURL := "https://api.openai.com/v1/chat/completions"
	apiKey := "api"

	requestBody := map[string]interface{}{
		"model": "gpt-4o-mini",
		"store": true,
		"messages": []map[string]string{
			{"role": "system", "content": gptRple},
			{"role": "user", "content": question},
		},
		"max_tokens": responseLength,
	}
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		log.Printf("Error marshaling request body: %v", err)
		return "", fmt.Errorf("failed to marshal request body: %v", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("HTTP request failed: %v", err)
		return "", fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		log.Printf("OpenAI API error response: %s", string(bodyBytes))
		return "", fmt.Errorf("API call failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}
	var responseData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		log.Printf("Error decoding response: %v", err)
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	choices, ok := responseData["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		log.Printf("Unexpected response structure: %+v", responseData)
		return "", fmt.Errorf("no choices found in response")
	}

	message, ok := choices[0].(map[string]interface{})["message"].(map[string]interface{})
	if !ok {
		log.Printf("Message structure missing in response: %+v", choices[0])
		return "", fmt.Errorf("message structure not found in choices")
	}

	content, ok := message["content"].(string)
	if !ok {
		log.Printf("Content missing in message: %+v", message)
		return "", fmt.Errorf("content not found in message")
	}

	return content, nil
}
