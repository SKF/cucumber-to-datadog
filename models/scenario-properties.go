package models

type ScenarioProperties struct {
	Scenario                 string
	Steps                    map[int]StepProperties
	HasProperties            bool
	DuplicateScenarioIndexes []int
}

type StepProperties struct {
	Step       string
	Properties map[int]string
}
