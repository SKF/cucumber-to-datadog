package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/SKF/cucumber-to-datadog/models"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	var apiKey, cucumberPath, stage, branch, service, testRunTitle, region, url string

	flag.StringVar(&apiKey, "apikey", "", "string")
	flag.StringVar(&cucumberPath, "cucumberPath", "", "string")
	flag.StringVar(&stage, "stage", "local", "string")
	flag.StringVar(&branch, "branch", "local", "string")
	flag.StringVar(&service, "service", "", "string")
	flag.StringVar(&testRunTitle, "testRunTitle", "", "string")
	flag.StringVar(&region, "region", "eu", "string")

	flag.Parse()

	fmt.Println(cucumberPath)

	if apiKey == "" {
		fmt.Printf("datadog api-key not set")
		return
	}
	if service == "" {
		fmt.Printf("service not set")
		return
	}
	if testRunTitle == "" {
		fmt.Printf("testRunTitle not set")
		return
	}

	switch region {
	case "eu":
		url = "https://http-intake.logs.datadoghq.eu/v1/input/"
	case "us":
		url = "https://http-intake.logs.datadoghq.com/v1/input/"
	default:
		fmt.Printf("region %s hasn't been implemented", region)
		return
	}

	testResults, err := parseCucumberFiles(cucumberPath)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	dt := time.Now()
	for _, testResult := range testResults {
		fmt.Printf("Feature: %+v\n", testResult.Name)
		featureOutcome := "passed"
		featureErrorMessage := ""

		scenarioProperties := getScenarioProperties(testResult.Elements)

		for scenarioIndex, element := range testResult.Elements {
			scenarioName := element.Name //strings.Replace(element.Name, " ", "_", -1)
			if properties, hasProperties := scenarioProperties[scenarioIndex]; hasProperties {
				scenarioName += " - (" + strings.Join(properties, ",") + ")"
			}

			fmt.Printf("Scenario: %+v\n", scenarioName)
			scenarioOutcome := "passed"
			scenarioErrorMessage := ""
			scenarioEndpoint := getScenarioEndpoint(element)
			scenarioMethod := getScenarioMethod(element)

			for _, step := range element.Steps {
				fmt.Printf("Step: %+v\n", step.Name)
				if step.Result.Status == "failed" {
					scenarioOutcome = "failed"
					scenarioErrorMessage = strings.Split(step.Result.ErrorMessage, "\n")[0]
					featureOutcome = "failed"
					featureErrorMessage = strings.Split(step.Result.ErrorMessage, "\n")[0]
				}
				ddStep := models.DatadogStepResult{
					Service:      service,
					Source:       service,
					Env:          stage,
					Type:         "CucumberStepResult",
					Date:         dt.Format("2006-01-02"),
					DateTime:     dt.Format("2006-01-02 15:04:05"),
					Feature:      strings.Replace(testResult.Name, " ", "_", -1),
					Scenario:     scenarioName,
					Step:         step.Keyword + step.Name,
					Outcome:      step.Result.Status,
					ErrorMessage: strings.Split(step.Result.ErrorMessage, "\n")[0],
					Branch:       branch,
					TestRunTitle: testRunTitle,
				}
				if ddStep.Outcome != "skipped" {
					if err := sendToDatadog(ddStep, apiKey, url); err != nil {
						fmt.Println(err.Error())
						return
					}
				}
			}
			ddScenario := models.DatadogScenarioResult{
				Service:      service,
				Source:       service,
				Env:          stage,
				Type:         "CucumberScenarioResult",
				Date:         dt.Format("2006-01-02"),
				DateTime:     dt.Format("2006-01-02 15:04:05"),
				Feature:      strings.Replace(testResult.Name, " ", "_", -1),
				Scenario:     scenarioName,
				Outcome:      scenarioOutcome,
				ErrorMessage: scenarioErrorMessage,
				Branch:       branch,
				TestRunTitle: testRunTitle,
				Endpoint:     scenarioEndpoint,
				Method:       scenarioMethod,
			}

			if err := sendToDatadog(ddScenario, apiKey, url); err != nil {
				fmt.Println(err.Error())
				return
			}
		}
		ddFeature := models.DatadogFeatureResult{
			Service:      service,
			Source:       service,
			Env:          stage,
			Type:         "CucumberFeatureResult",
			Date:         dt.Format("2006-01-02"),
			DateTime:     dt.Format("2006-01-02 15:04:05"),
			Feature:      strings.Replace(testResult.Name, " ", "_", -1),
			Outcome:      featureOutcome,
			ErrorMessage: featureErrorMessage,
			Branch:       branch,
			TestRunTitle: testRunTitle,
		}

		if err := sendToDatadog(ddFeature, apiKey, url); err != nil {
			fmt.Println(err.Error())
			return
		}
	}
}

func sendToDatadog(testResult interface{}, datadogApiKey, datadogUrl string) (err error) {
	buf := &bytes.Buffer{}
	if err = json.NewEncoder(buf).Encode(testResult); err != nil {
		return
	}
	url := datadogUrl + datadogApiKey
	req, err := http.NewRequest("POST", url, buf)
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("error response code: %v", resp.StatusCode)
	}
	return
}

func parseCucumberFiles(path string) (testResults []models.CucumberTestResult, err error) {
	fmt.Println(path)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, err
	}

	dir, err := os.Open(path)
	if err != nil {
		return
	}
	defer dir.Close()

	files, err := dir.Readdir(-1)
	if err != nil {
		return
	}
	for _, file := range files {
		matched, _ := filepath.Match("*.cucumber.json", file.Name())
		if matched {
			fmt.Println(file.Name())
			jsonFile, err := os.Open(filepath.Join(path, file.Name()))
			if err != nil {
				return nil, err
			}
			defer jsonFile.Close()
			byteValue, _ := ioutil.ReadAll(jsonFile)
			var featureTestResults []models.CucumberTestResult
			err = json.Unmarshal(byteValue, &featureTestResults)
			if err != nil {
				return nil, err
			}
			for _, testResult := range featureTestResults {
				testResults = append(testResults, testResult)
			}
		}
	}
	return
}

func getScenarioProperties(elements []models.Element) (output map[int][]string) {

	var scenarios []models.ScenarioProperties
	for i, element := range elements {
		scenario := models.ScenarioProperties{
			Scenario: element.Name,
			Steps:    make(map[int]models.StepProperties),
		}

		for stepIndex, step := range element.Steps {
			stepWithoutProperties := ""
			stepProperties := make(map[int]string)
			for i, stepPart := range strings.Split(step.Name, "\"") {
				if i%2 == 0 { // skip even numbers
					stepWithoutProperties += stepPart
				} else {
					stepProperties[i] = stepPart
				}
			}
			scenario.Steps[stepIndex] = models.StepProperties{
				Step:       stepWithoutProperties,
				Properties: stepProperties,
			}
		}
		for j, existingScenario := range scenarios {
			if element.Name == existingScenario.Scenario {
				scenarios[j].DuplicateScenarioIndexes = append(scenarios[j].DuplicateScenarioIndexes, i)
				scenarios[j].HasProperties = true
				scenario.DuplicateScenarioIndexes = append(scenario.DuplicateScenarioIndexes, j)
				scenario.HasProperties = true
			}
		}
		scenarios = append(scenarios, scenario)
	}

	output = make(map[int][]string)

	for i, scenario := range scenarios {
		if scenario.HasProperties {
			var properties []string
			for line, step := range scenario.Steps {
				for _, duplicateIndex := range scenario.DuplicateScenarioIndexes {
					for propertyIndex, stepProperty := range step.Properties {
						if stepProperty != scenarios[duplicateIndex].Steps[line].Properties[propertyIndex] {
							propExists := false
							for _, existingProps := range properties {
								if existingProps == stepProperty {
									propExists = true
								}
							}
							if !propExists {
								properties = append(properties, stepProperty)
							}
						}
					}
				}
			}
			output[i] = properties
		}
	}
	return output
}

func getScenarioMethod(scenario models.Element) string {
	for _, step := range scenario.Steps {
		for i, stepPart := range strings.Split(step.Name, "\"") {
			if i%2 != 0 { // skip odd numbers
				method := strings.ToUpper(stepPart)
				if method == "GET" || method == "POST" || method == "DELETE" || method == "PUT" {
					return method
				}

			}
		}
	}
	return ""
}

func getScenarioEndpoint(scenario models.Element) string {
	for _, step := range scenario.Steps {
		for i, stepPart := range strings.Split(step.Name, "\"") {
			if i%2 != 0 { // skip odd numbers
				if strings.HasPrefix(stepPart, "/") {
					return stepPart
				}
			}
		}
	}
	return ""
}
