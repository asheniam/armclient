package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"strings"
)

type Environment struct {
	aadLoginUrl string
	apiVersion  string
	armUrl      string
	httpTransport http.Transport
}

const (
	PublicEnvironmentName = "Public"
	GermanEnvironmentName = "AzureGermanCloud"
)

var (
	// Public Azure
	publicAzureEnvironment = Environment{
		aadLoginUrl:   "https://login.microsoftonline.com",
		apiVersion:    "2017-08-01",
		armUrl:        "https://management.azure.com",
		httpTransport: http.Transport{},
	}

	germanAzureEnvironment = Environment{
		aadLoginUrl: "https://login.microsoftonline.de",
		apiVersion: "2016-06-01",
		armUrl:      "https://management.microsoftazure.de",
		httpTransport: http.Transport{
			TLSClientConfig: &tls.Config{
				MaxVersion:    tls.VersionTLS11,
				Renegotiation: tls.RenegotiateFreelyAsClient,
			},
		},
	}
)

func getCurrentEnvironment(environmentName string) *Environment {
	if strings.EqualFold(environmentName, PublicEnvironmentName) {
		return &publicAzureEnvironment
	} else if strings.EqualFold(environmentName, GermanEnvironmentName) {
		return &germanAzureEnvironment
	}

	log.Fatalf("Unknown environment: %s", environmentName)
	return nil
}

func (environment Environment) getAadLoginUrl(TenantID string) string {
	return fmt.Sprintf("%s/%s/oauth2/token", environment.aadLoginUrl, TenantID)
}
