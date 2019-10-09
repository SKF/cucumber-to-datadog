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

const DataDogBaseUrl = "https://http-intake.logs.datadoghq.com/v1/input/"

func main() {
	var apiKey, cucumberPath, stage, branch, service, testRunTitle string

	flag.StringVar(&apiKey, "apikey", "", "string")
	flag.StringVar(&cucumberPath, "cucumberPath", "", "string")
	flag.StringVar(&stage, "stage", "", "string")
	flag.StringVar(&branch, "branch", "local", "string")
	flag.StringVar(&service, "service", "", "string")
	flag.StringVar(&testRunTitle, "testRunTitle", "", "string")

	flag.Parse()

	fmt.Println(cucumberPath)

	if stage == "" {
		fmt.Printf("stage not set")
		return
	}
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

	testResults, err := parseCucumberFiles(cucumberPath)
	if err!= nil {
		fmt.Println(error.Error)
		return
	}

	dt := time.Now()

	for _, testResult := range testResults {
		featureOutcome := "passed"
		featureErrorMessage := ""
		for _, element := range testResult.Elements {
			scenarioOutcome := "passed"
			scenarioErrorMessage := ""
			for _, step := range element.Steps {
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
					Scenario:     strings.Replace(element.Name, " ", "_", -1),
					Step:         step.Keyword + step.Name,
					Outcome:      step.Result.Status,
					ErrorMessage: strings.Split(step.Result.ErrorMessage, "\n")[0],
					Branch:       branch,
					TestRunTitle: testRunTitle,
				}
				fmt.Printf("%+v\n", ddStep)
				if ddStep.Outcome != "skipped" {
					if err,_ := sendToDatadog(ddStep, apiKey); err!= nil {
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
				Scenario:     strings.Replace(element.Name, " ", "_", -1),
				Outcome:      scenarioOutcome,
				ErrorMessage: scenarioErrorMessage,
				Branch:       branch,
				TestRunTitle: testRunTitle,
			}
			if err,_ := sendToDatadog(ddScenario, apiKey); err!= nil {
				fmt.Println(err.Error())
				return
			}
			fmt.Printf("%+v\n", ddScenario)
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
		if err,_ := sendToDatadog(ddFeature, apiKey); err!= nil {
			fmt.Println(err.Error())
			return
		}
		fmt.Printf("%+v\n", ddFeature)
	}
}

func sendToDatadog(testResult interface{}, datadogApiKey string) (err error, response interface{}) {

	buf := &bytes.Buffer{}
	if err = json.NewEncoder(buf).Encode(testResult); err != nil {
		return
	}

	url := DataDogBaseUrl + datadogApiKey
	req, err := http.NewRequest("POST", url, buf)
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("error response code: %v", resp.StatusCode), nil
	}

	if resp.ContentLength != 0 {
		if err = json.NewDecoder(resp.Body).Decode(response); err != nil {
			return
		}
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
			json.Unmarshal(byteValue, &featureTestResults)

			for _, testResult := range featureTestResults {
				testResults = append(testResults, testResult)
			}
		}
	}
	return
}
