package bluecare

type Service struct {
	Name       string `json:"name"`
	ConsoleURL string `json:"console"`
}

type ServiceList struct {
	Services map[string]Service `json:"services"`
}
