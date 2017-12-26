package main

import (
	"fmt"
	"io/ioutil"

	log "github.com/sirupsen/logrus"
)

type CommandProcessor struct {
	azureClient *AzureClient
}

func NewCommandProcessor(config *Config, environment *Environment) *CommandProcessor {
	return &CommandProcessor{
		azureClient: NewAzureClient(config, environment),
	}
}

func (processor *CommandProcessor) processGetCommand(getUrl string) {
	response := processor.azureClient.sendHttpMessage("GET", getUrl)

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatalf("Error reading body of response: %v", err)
	}

	prettyPrintJson(body)
}

func (processor *CommandProcessor) processSummarizeCommand(maxContinuation int) {
	// Invoke Azure Resource Manager resource cache API to find all Azure resources on the subscription
	armResourceSlice := make([]ArmResource, 10)
	targetUrl := fmt.Sprintf("/subscriptions/%s/resources?api-version=2017-08-01", processor.azureClient.config.Credentials.SubscriptionID)

	// Follow nextLink continuation tokens
	i := 0
	for len(targetUrl) > 0 && i <= maxContinuation {

		targetUrl = func(getUrl string) string {
			response := processor.azureClient.sendHttpMessage("GET", getUrl)
			defer response.Body.Close()
			body, err := ioutil.ReadAll(response.Body)
			if err != nil {
				log.Fatalf("Error reading body of response: %v", err)
			}

			armResourceListResponse := convertToArmResourceListResponse(body)
			for _, armResource := range armResourceListResponse.Values {
				armResourceSlice = append(armResourceSlice, armResource)
			}

			return armResourceListResponse.NextLink
		}(targetUrl)

		i++
	}

	// Format the console output to group by {Location}, {ResourceType}, {Id}
	armResourceMap := make(map[string]map[string]map[string]ArmResource)
	for _, armResource := range armResourceSlice {
		if len(armResource.Location) == 0 {
			continue
		}

		armResourceByResourceTypeMap, ok := armResourceMap[armResource.Location]
		if !ok {
			armResourceByResourceTypeMap = make(map[string]map[string]ArmResource)
			armResourceMap[armResource.Location] = armResourceByResourceTypeMap
		}

		armResourceByIdMap, ok := armResourceByResourceTypeMap[armResource.Type]
		if !ok {
			armResourceByIdMap = make(map[string]ArmResource)
			armResourceByResourceTypeMap[armResource.Type] = armResourceByIdMap
		}

		armResourceByIdMap[armResource.Id] = armResource
	}

	for armLocation, armResourceByResourceTypeMap := range armResourceMap {
		fmt.Printf("Location: %s:\n", armLocation)
		for armResourceType, armResourceByIdMap := range armResourceByResourceTypeMap {
			fmt.Printf("  ResourceType: %s:\n", armResourceType)
			for _, armResource := range armResourceByIdMap {
				fmt.Printf("    Id: %s:\n", armResource.Id)
			}

			fmt.Println()
		}
	}
}
