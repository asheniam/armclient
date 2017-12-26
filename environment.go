package main

import (
	"fmt"
	"log"
	"strings"
)

type Environment struct {
	aadLoginUrl string
	armUrl      string
}

const (
	PublicEnvironmentName = "Public"
)

var (
	// Public Azure
	publicAzureEnvironment = Environment{
		aadLoginUrl: "https://login.microsoftonline.com",
		armUrl:      "https://management.azure.com",
	}
)

func getCurrentEnvironment(environmentName string) *Environment {
	if strings.EqualFold(environmentName, PublicEnvironmentName) {
		return &publicAzureEnvironment
	}

	log.Fatalf("Unknown environment: %s", environmentName)
	return nil
}

func (environment Environment) getAadLoginUrl(TenantID string) string {
	return fmt.Sprintf("%s/%s/oauth2/token", environment.aadLoginUrl, TenantID)
}
