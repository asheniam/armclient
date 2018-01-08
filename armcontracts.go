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

// TODO, we need to test this with SQL databases -- server/database
func (armResource *ArmResource) getResourceName() string {
	armResourceIdParts := strings.Split(armResource.Id, "/")
	return armResourceIdParts[len(armResourceIdParts)-1]
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
