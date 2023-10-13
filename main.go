package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"gopkg.in/yaml.v2"
)

func main() {
	url := "https://raw.githubusercontent.com/aws/aws-sdk-net/master/sdk/src/Core/endpoints.json"
	localJSONPath := "endpoints.json"
	localYAMLPath := "endpoints.yaml"

	if _, err := os.Stat(localJSONPath); os.IsNotExist(err) {
		response, err := http.Get(url)
		if err != nil {
			fmt.Println("Failed to download the JSON file.")
			os.Exit(1)
		}
		defer response.Body.Close()

		file, err := os.Create(localJSONPath)
		if err != nil {
			fmt.Println("Failed to create JSON file.")
			os.Exit(1)
		}
		defer file.Close()

		_, err = io.Copy(file, response.Body)
		if err != nil {
			fmt.Println("Failed to save JSON file.")
			os.Exit(1)
		}
	}

	jsonFile, err := os.Open(localJSONPath)
	if err != nil {
		fmt.Println("Failed to read JSON file.")
		os.Exit(1)
	}
	defer jsonFile.Close()

	decoder := json.NewDecoder(jsonFile)
	var data map[string]interface{}
	err = decoder.Decode(&data)
	if err != nil {
		fmt.Println("Failed to parse JSON data.")
		os.Exit(1)
	}

	partitions, ok := data["partitions"].([]interface{})
	if !ok || len(partitions) == 0 {
		fmt.Println("Invalid JSON data structure.")
		os.Exit(1)
	}

	partition, ok := partitions[0].(map[string]interface{})
	if !ok {
		fmt.Println("Invalid JSON data structure.")
		os.Exit(1)
	}

	services, ok := partition["services"].(map[string]interface{})
	if !ok {
		fmt.Println("Invalid JSON data structure.")
		os.Exit(1)
	}

	var newServiceNames []string
	for serviceName := range services {
		newServiceNames = append(newServiceNames, serviceName)
	}

	yamlData := make(map[string]interface{})
	yamlData["service_names"] = newServiceNames

	if _, err := os.Stat(localYAMLPath); os.IsNotExist(err) {
		err = writeYAMLFile(localYAMLPath, yamlData)
		if err != nil {
			fmt.Println("Failed to write YAML file.")
			os.Exit(1)
		}
	}

	fmt.Printf("New service names added: %v\n", newServiceNames)
	fmt.Println("Merged Data:")
	printYAML(yamlData)
}

func writeYAMLFile(filePath string, data interface{}) error {
	yamlBytes, err := yaml.Marshal(data)
	if err != nil {
		return err
	}

	yamlFile, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer yamlFile.Close()

	_, err = yamlFile.Write(yamlBytes)
	if err != nil {
		return err
	}

	return nil
}

func printYAML(data interface{}) {
	yamlBytes, err := yaml.Marshal(data)
	if err != nil {
		fmt.Println("Failed to marshal data to YAML.")
		os.Exit(1)
	}
	fmt.Printf("%s\n", yamlBytes)
}
