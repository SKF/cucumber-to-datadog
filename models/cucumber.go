package models

type CucumberTestResult struct {
	Name     string    `json:"name"`
	Keyword  string    `json:"keyword"`
	Uri      string    `json:"uri"`
	Elements []Element `json:"elements"`
}

type Element struct {
	Name               string `json:"name"`
	Keyword            string `json:"keyword"`
	Steps              []Step `json:"steps"`
}

type Step struct {
	Name    string `json:"name"`
	Keyword string `json:"keyword"`
	Result  Result `json:"result"`
}

type Result struct {
	Status       string `json:"status"`
	Duration     int    `json:"duration"`
	ErrorMessage string `json:"error_message"`
}
