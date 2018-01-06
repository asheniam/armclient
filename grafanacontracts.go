package main

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"
)

type GrafanaDashboard struct {
	ParsedJson map[string]interface{}
}

func NewGrafanaDashboard(templateContents string) *GrafanaDashboard {
	var dashboard GrafanaDashboard
	err := json.Unmarshal([]byte(templateContents), &dashboard.ParsedJson)
	if err != nil {
		log.Fatalf("Error parsing template: %v", err)
	}

	return &dashboard
}

// Update the contents of the Grafana dashboard template with Azure resource IDs
func (dashboard *GrafanaDashboard) update(title string, dataSourceName string, maxDashboardResources int, armResources []ArmResource) {
	var rowsJson []interface{}
	var panelsJson []interface{}
	var targetsJson []interface{}

	dashboard.ParsedJson["title"] = title
	rowsJson = dashboard.ParsedJson["rows"].([]interface{})
	for _, rowJsonObject := range rowsJson {
		rowJson := rowJsonObject.(map[string]interface{})
		panelsJson = rowJson["panels"].([]interface{})
		for _, panelJsonObject := range panelsJson {
			panelJson := panelJsonObject.(map[string]interface{})
			panelJson["datasource"] = dataSourceName

			targetsJson = panelJson["targets"].([]interface{})
			if len(targetsJson) > 0 {
				// Only the first target matters.
				targetJson := targetsJson[0].(map[string]interface{})
				azureMonitorTargetJson := targetJson["azureMonitor"].(map[string]interface{})

				newTargetsJson := make([]map[string]interface{}, 0)
				upperBound := len(armResources)
				if maxDashboardResources < upperBound {
					upperBound = maxDashboardResources
				}

				// For each ARM resource, we will generate new target
				for _, armResource := range armResources[:upperBound] {
					newAzureMonitorTargetJson := copyMap(azureMonitorTargetJson)
					newAzureMonitorTargetJson["resourceGroup"], _ = armResource.getResourceGroupName()
					newAzureMonitorTargetJson["resourceName"] = armResource.getResourceName()

					newTargetJson := copyMap(targetJson)
					newTargetJson["azureMonitor"] = newAzureMonitorTargetJson
					newTargetsJson = append(newTargetsJson, newTargetJson)
				}

				panelJson["targets"] = newTargetsJson
			}
		}
	}
}

func copyMap(originalMap map[string]interface{}) map[string]interface{} {
	newMap := make(map[string]interface{})
	for k, v := range originalMap {
		newMap[k] = v
	}

	return newMap
}
