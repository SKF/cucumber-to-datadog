# cucumber-to-datadog
Tool for sending cucumber test-results to datadog

## Install

    go get github.com/SKF/cucumber-to-datadog
    
### Prerequisites
The cucumber-to-datadog package requires that the tests generates a cucumber json results file
The cucumber json filename must end with .cucumber.json

#### Prerequisites for cypress
The cucumber-to-datadog package requires the [cypress-cucumber-preprocessor](https://github.com/TheBrainFamily/cypress-cucumber-preprocessor/blob/master/README.md)

The cypress-cucumber-preprocessor must be configured to generate cucumber json.
In the package.json file add the follwing code:
```
"cypress-cucumber-preprocessor": {
    "cucumberJson": {
      "generate": true,
    }
  }
```

### Prerequisites for godog
add -f=cucumber to the datadog execution and output the results to a json file that ends with .cucumber.json
```
godog -f=cucumber mytests.feature > mytests.cucumber.json
```


## Usage

    cucumber-to-datadog --cucumberPath=<path> --stage=<stage> --service=<service> --testRunTitle=<testRunTitle> --apiKey=<apiKey>

### Options
Option | Required | Description
--------------- | -------- | -------------
--cucumberPath | Yes | The folder where the cucumber .json results are located. The cucmber result files must end with .cucumber.json
--stage| No | The environment to which the tests were running against, default value = local
--branch | No | The git branch from which the tests were executed, default value = local
--service | Yes | The name of the service
--testRunTitle | Yes | The title of the test-run
--apiKey | Yes | Your datadog api key
--region | No | The region for datadog, default value = us, available values: eu, us