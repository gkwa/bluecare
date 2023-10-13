package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
)

func main() {
	url := "https://raw.githubusercontent.com/aws/aws-sdk-net/master/sdk/src/Core/endpoints.json"
	jsonIncoming := "endpoints.json"
	jsonEdited := "endpoints_edited.json"

	existingData := make(map[string]map[string]string)
	if _, err := os.Stat(jsonEdited); err == nil {
		existingData = readExistingData(jsonEdited)
	}

	if _, err := os.Stat(jsonIncoming); os.IsNotExist(err) {
		response, err := http.Get(url)
		if err != nil {
			slog.Error("Failed to download the JSON file.")
			os.Exit(1)
		}
		defer response.Body.Close()

		file, err := os.Create(jsonIncoming)
		if err != nil {
			slog.Error("Failed to create JSON file.")
			os.Exit(1)
		}
		defer file.Close()

		_, err = io.Copy(file, response.Body)
		if err != nil {
			slog.Error("Failed to save JSON file.")
			os.Exit(1)
		}
	}

	jsonFile, err := os.Open(jsonIncoming)
	if err != nil {
		slog.Error("Failed to read JSON file.")
		os.Exit(1)
	}
	defer jsonFile.Close()

	decoder := json.NewDecoder(jsonFile)
	var data map[string]interface{}
	err = decoder.Decode(&data)
	if err != nil {
		slog.Error("Failed to parse JSON data.")
		os.Exit(1)
	}

	partitions, ok := data["partitions"].([]interface{})
	if !ok || len(partitions) == 0 {
		slog.Error("Invalid JSON data structure.")
		os.Exit(1)
	}

	partition, ok := partitions[0].(map[string]interface{})
	if !ok {
		slog.Error("Invalid JSON data structure.")
		os.Exit(1)
	}

	services, ok := partition["services"].(map[string]interface{})
	if !ok {
		slog.Error("Invalid JSON data structure.")
		os.Exit(1)
	}

	newServiceNames := make(map[string]map[string]string)
	for serviceName := range services {
		if existingURL, exists := existingData[serviceName]; exists {
			newServiceNames[serviceName] = existingURL
		} else {
			newServiceNames[serviceName] = map[string]string{
				"console": fmt.Sprintf("https://us-west-1.console.aws.amazon.com/%s/home?region=us-west-1#", serviceName),
			}
		}
	}

	jsonData := map[string]interface{}{
		"service_names": newServiceNames,
	}

	if err := writeJSONFile(jsonEdited, jsonData); err != nil {
		slog.Error("Failed to write JSON file.")
		os.Exit(1)
	}

	fmt.Printf("Service names and URLs updated:\n")
}

func writeJSONFile(filePath string, data interface{}) error {
	jsonBytes, err := json.MarshalIndent(data, "", "  ") // Indent with two spaces for pretty printing
	if err != nil {
		return err
	}

	jsonFile, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer jsonFile.Close()

	_, err = jsonFile.Write(jsonBytes)
	return err
}

func readExistingData(filePath string) map[string]map[string]string {
	jsonFile, err := os.Open(filePath)
	if err != nil {
		slog.Error("Failed to read existing JSON file.")
		return make(map[string]map[string]string)
	}
	defer jsonFile.Close()

	decoder := json.NewDecoder(jsonFile)
	var data map[string]map[string]map[string]string
	err = decoder.Decode(&data)
	if err != nil {
		slog.Error("Failed to parse existing JSON data.")
		return make(map[string]map[string]string)
	}

	return data["service_names"]
}
