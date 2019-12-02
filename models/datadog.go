package models

type DatadogStepResult struct {
	Service      string `json:"service"`
	Source       string `json:"ddsource"`
	Env          string `json:"env"`
	Type         string `json:"type"`
	Date         string `json:"date"`
	DateTime     string `json:"datetime"`
	Feature      string `json:"feature"`
	Scenario     string `json:"scenario"`
	Step         string `json:"step"`
	Outcome      string `json:"outcome"`
	ErrorMessage string `json:"error_message"`
	Branch       string `json:"branch"`
	TestRunTitle string `json:"testruntitle"`
}

type DatadogScenarioResult struct {
	Service      string `json:"service"`
	Source       string `json:"ddsource"`
	Env          string `json:"env"`
	Type         string `json:"type"`
	Date         string `json:"date"`
	DateTime     string `json:"datetime"`
	Feature      string `json:"feature"`
	Scenario     string `json:"scenario"`
	Outcome      string `json:"outcome"`
	ErrorMessage string `json:"error_message"`
	Branch       string `json:"branch"`
	TestRunTitle string `json:"testruntitle"`
	Endpoint     string `json:"endpoint"`
	Method       string `json:"method"`
}

type DatadogFeatureResult struct {
	Service      string `json:"service"`
	Source       string `json:"ddsource"`
	Env          string `json:"env"`
	Type         string `json:"type"`
	Date         string `json:"date"`
	DateTime     string `json:"datetime"`
	Feature      string `json:"feature"`
	Outcome      string `json:"outcome"`
	ErrorMessage string `json:"error_message"`
	Branch       string `json:"branch"`
	TestRunTitle string `json:"testruntitle"`
}
