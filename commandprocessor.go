package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	. "github.com/ahmetb/go-linq"
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
	armResources := processor.azureClient.getAzureResources(maxContinuation)

	// Format the console output to group by {Location}, {ResourceType}, {Id}
	armResourceMap := make(map[string]map[string]map[string]ArmResource)
	for _, armResource := range armResources {
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

func (processor *CommandProcessor) processGrafanaCommand(titlePrefix string, dataSourceName string, maxContinuation int, maxDashboardResources int, resourceType string) {
	// Invoke Azure Resource Manager resource cache API to find all Azure resources on the subscription
	armResources := processor.azureClient.getAzureResources(maxContinuation)

	// Filter by resource type, TODO move into GET API
	var filteredArmResources []ArmResource
	From(armResources).WhereT(func(r ArmResource) bool {
		return strings.EqualFold(r.Type, resourceType)
	}).ToSlice(&filteredArmResources)

	armResources = filteredArmResources

	// Group by {Location}
	armResourceMap := make(map[string]map[string]ArmResource)
	for _, armResource := range armResources {
		if len(armResource.Location) == 0 {
			continue
		}

		armResourceByIdMap, ok := armResourceMap[armResource.Location]
		if !ok {
			armResourceByIdMap = make(map[string]ArmResource)
			armResourceMap[armResource.Location] = armResourceByIdMap
		}

		armResourceByIdMap[armResource.Id] = armResource
	}

	// Read Grafana JSON template for resource type
	dashboardTemplates := getGitHubGrafanaTemplates(resourceType)
	if len(dashboardTemplates) == 0 {
		fmt.Println("No dashboards found for resource type on github")
	}

	for _, dashboardTemplate := range dashboardTemplates {
		distinctRegions := getDistinctRegions(armResources)
		distinctRegions = append(distinctRegions, "allregions")

		// Generate Grafana dashboard JSONs - one dashboard for each region and one dashboard for all regions
		for _, region := range distinctRegions {
			var dashboardArmResources []ArmResource
			if strings.EqualFold(region, "allregions") {
				dashboardArmResources = armResources
			} else {
				From(armResources).WhereT(func(r ArmResource) bool {
					return strings.EqualFold(r.Location, region)
				}).ToSlice(&dashboardArmResources)
			}

			dashboard := NewGrafanaDashboard(dashboardTemplate.Contents)
			title := fmt.Sprintf("%s - %s - %s - %s", titlePrefix, resourceType, dashboardTemplate.Name, region)
			dashboard.update(title, dataSourceName, maxDashboardResources, dashboardArmResources)

			generatedDashboard, err := json.MarshalIndent(dashboard.ParsedJson, "", " ")
			if err != nil {
				log.Fatalf("Error generating dashboard: %v", err)
			}

			outputFile := strings.ToLower("dashboard_" + titlePrefix + "_" + resourceType + "_" + region)
			outputFile = strings.Replace(outputFile, " ", "_", -1)
			outputFile = strings.Replace(outputFile, "/", "_", -1)
			outputFile = strings.Replace(outputFile, ".", "_", -1)
			outputFile += ".json"
			ioutil.WriteFile(outputFile, generatedDashboard, 0644)
		}
	}
}
