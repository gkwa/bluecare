package bluecare

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"

	mymazda "github.com/taylormonacelli/forestfish/mymazda"
)

var endpointsEditedJsonPath = "/tmp/endpoints_edited.json"

func FetchEditedEndpoints() {
	url := "https://raw.githubusercontent.com/taylormonacelli/bluecare/master/endpoints_edited.json"

	slog.Debug("fetching file", "url", url)

	resp, err := http.Get(url)
	if err != nil {
		slog.Error("fetching file failed", "error", err.Error())
		return
	}

	slog.Debug("file fetched successfully", "url", url, "path", endpointsEditedJsonPath)

	// Print the request for debugging
	requestDump, err := httputil.DumpRequestOut(resp.Request, true)
	if err != nil {
		slog.Error("failed to dump request", "error", err.Error())
	} else {
		slog.Debug("request dump", "dump", string(requestDump))
	}

	defer resp.Body.Close()

	outFile, err := os.Create(endpointsEditedJsonPath)
	if err != nil {
		slog.Error("failed to create output file", "path", endpointsEditedJsonPath, "error", err.Error())
		return
	}
	slog.Debug("output file created", "path", endpointsEditedJsonPath)

	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		slog.Error("file save failed", "error", err.Error())
		return
	}
	slog.Debug("file saved", "path", endpointsEditedJsonPath)

	// Print the response for debugging
	responseDump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		slog.Error("Failed to dump response:", err)
	} else {
		slog.Debug("http fetch", "response", string(responseDump))
	}
}

func GetServices() []string {
	myMap, err := GetServiceURLMap()
	if err != nil {
		panic(err)
	}
	keys := make([]string, 0, len(myMap))
	for key := range myMap {
		keys = append(keys, key)
	}

	return keys
}

func testLoad() error {
	slog.Debug("check file exists", "path", endpointsEditedJsonPath)

	file, err := os.Open(endpointsEditedJsonPath)
	if err != nil {
		slog.Warn("error opening JSON file", "path", endpointsEditedJsonPath, "error", err.Error())
		return err
	}
	defer file.Close()

	var serviceList ServiceList

	decoder := json.NewDecoder(file)

	if err := decoder.Decode(&serviceList); err != nil {
		slog.Warn("error decoding JSON", "error", err.Error())
		return err
	}

	return nil
}

func GetServiceURLMap() (map[string]string, error) {
	if !mymazda.FileExists(endpointsEditedJsonPath) {
		FetchEditedEndpoints()
	}

	file, err := os.Open(endpointsEditedJsonPath)
	if err != nil {
		slog.Error("Error opening the file", "error", err.Error())
		return make(map[string]string), err
	}
	defer file.Close()

	serviceURLMap := make(map[string]string)

	var data map[string]map[string]map[string]string

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		slog.Error("Error decoding JSON", "error", err.Error())
		return make(map[string]string), err
	}

	for serviceName, serviceData := range data["services"] {
		serviceURL := serviceData["console"]
		serviceURLMap[serviceName] = serviceURL
	}

	return serviceURLMap, nil
}

func fetchAndReconcile() int {
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
			return 1
		}
		defer response.Body.Close()

		file, err := os.Create(jsonIncoming)
		if err != nil {
			slog.Error("Failed to create JSON file.")
			return 1
		}
		defer file.Close()

		_, err = io.Copy(file, response.Body)
		if err != nil {
			slog.Error("Failed to save JSON file.")
			return 1
		}
	}

	jsonFile, err := os.Open(jsonIncoming)
	if err != nil {
		slog.Error("Failed to read JSON file.")
		return 1
	}
	defer jsonFile.Close()

	decoder := json.NewDecoder(jsonFile)
	var data map[string]interface{}
	err = decoder.Decode(&data)
	if err != nil {
		slog.Error("Failed to parse JSON data.")
		return 1
	}

	partitions, ok := data["partitions"].([]interface{})
	if !ok || len(partitions) == 0 {
		slog.Error("Invalid JSON data structure.")
		return 1
	}

	partition, ok := partitions[0].(map[string]interface{})
	if !ok {
		slog.Error("Invalid JSON data structure.")
		return 1
	}

	services, ok := partition["services"].(map[string]interface{})
	if !ok {
		slog.Error("Invalid JSON data structure.")
		return 1
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
		"services": newServiceNames,
	}

	if err := writeJSONFile(jsonEdited, jsonData); err != nil {
		slog.Error("Failed to write JSON file.")
		return 1
	}

	return 0
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

	return data["services"]
}

func GetServiceURLInRegion(service, region string) (string, error) {
	url, err := GetServiceURL(service)
	if err != nil {
		slog.Error("resolve url to service", "error", err.Error())
		return "", err
	}

	url = strings.Replace(url, "us-west-1", region, -1)
	return url, nil
}

func GetServiceURL(service string) (string, error) {
	serviceMap, err := GetServiceURLMap()
	if err != nil {
		return "", err
	}

	return serviceMap[service], nil
}

func Execute(service, region string) int {
	fetchAndReconcile()

	err := testLoad()
	if err != nil {
		FetchEditedEndpoints()
	}

	url, _ := GetServiceURLInRegion(service, region)
	slog.Debug("get url", "service", service, "url", url)
	return 0
}
