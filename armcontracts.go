package main

import (
	"encoding/json"
	"fmt"
	"strings"

	. "github.com/ahmetb/go-linq"
	log "github.com/sirupsen/logrus"
)

type ArmResourceListResponse struct {
	Values   []ArmResource `json:"value"`
	NextLink string        `json:"nextLink"`
}

type ArmResource struct {
	Id       string         `json:"id"`
	Location string         `json:"location"`
	Name     string         `json:"name"`
	Type     string         `json:"type"`
	Kind     string         `json:"kind"`
	Sku      ArmResourceSku `json:"sku"`
}

type ArmResourceSku struct {
	Name string `json:"name"`
	Size string `json:"size"`
	Tier string `json:"tier"`
}

func (armResource *ArmResource) getResourceGroupName() (string, error) {
	armResourceIdParts := strings.Split(armResource.Id, "/")
	for index, armResourceIdPart := range armResourceIdParts {
		// Resource group name is the next segment after 'resourcegroups'
		if strings.EqualFold(armResourceIdPart, "resourcegroups") {
			if index < len(armResourceIdParts)-1 {
				return armResourceIdParts[index+1], nil
			} else {
				return "", fmt.Errorf("Unable to find resource group")
			}
		}
	}

	return "", fmt.Errorf("Unable to find resource group")
}

func (armResource *ArmResource) getResourceName() string {
	resourceName := ""
	armResourceIdParts := strings.Split(armResource.Id, "/")

	// To generate the ARM resource name, take every other segment after {providers}
	providerIndex := -1
	for index, armResourceIdPart := range armResourceIdParts {

		if providerIndex > 0 && index > providerIndex+1 && index%2 == 0 {
			if len(resourceName) > 0 {
				resourceName += "/"
			}

			resourceName += armResourceIdPart
		}

		if strings.EqualFold(armResourceIdPart, "providers") {
			providerIndex = index
		}
	}

	return resourceName
}

func getDistinctRegions(armResources []ArmResource) []string {
	var regions []string
	From(armResources).SelectT(
		func(r ArmResource) string {
			return r.Location
		},
	).DistinctByT(
		func(l string) string {
			return l
		},
	).ToSlice(&regions)

	return regions
}

func convertToArmResourceListResponse(body []byte) ArmResourceListResponse {
	var armResponse ArmResourceListResponse
	err := json.Unmarshal(body, &armResponse)
	if err != nil {
		log.Fatalf("Error unmarshalling ARM reource response body: %v", err)
	}

	return armResponse
}
