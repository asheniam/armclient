package main

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"
)

type ArmResourceListResponse struct {
	Values   []ArmResource `json:"value"`
	NextLink string        `json:"nextLink"`
}

type ArmResource struct {
	Id       string `json:"id"`
	Location string `json:"location"`
	Name     string `json:"name"`
	Type     string `json:"type"`
}

func convertToArmResourceListResponse(body []byte) ArmResourceListResponse {
	var armResponse ArmResourceListResponse
	err := json.Unmarshal(body, &armResponse)
	if err != nil {
		log.Fatalf("Error unmarshalling ARM reource response body: %v", err)
	}

	return armResponse
}
