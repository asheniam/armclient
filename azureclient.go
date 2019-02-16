package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"
)

type AzureClient struct {
	client      *http.Client
	config      *Config
	environment *Environment
	accessToken string
}

func NewAzureClient(config *Config, environment *Environment) *AzureClient {
	return &AzureClient{
		client:      &http.Client{},
		config:      config,
		environment: environment,
		accessToken: "",
	}
}

func (azureClient *AzureClient) setAccessToken() {
	aadLoginUrl := azureClient.environment.getAadLoginUrl(azureClient.config.Credentials.TenantID)

	// Important: For AAD resource, there needs to be a trailing slash after the ARM URL
	resource := azureClient.environment.armUrl
	if !strings.HasSuffix(resource, "/") {
		resource += "/"
	}

	form := url.Values{
		"grant_type":    {"client_credentials"},
		"resource":      {resource},
		"client_id":     {azureClient.config.Credentials.ClientID},
		"client_secret": {azureClient.config.Credentials.ClientSecret},
	}

	log.Debugf("Getting token from %s\n", aadLoginUrl)

	response, err := azureClient.client.PostForm(aadLoginUrl, form)

	if err != nil {
		log.Fatalf("Error authenticating against Azure API: %v", err)
	}

	if response.StatusCode != 200 {
		log.Fatalf("Error authenticating against Azure API - status code: %d", response.StatusCode)
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		log.Fatalf("Error authenticating against Azure API - error reading body of response: %v", err)
	}

	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Fatalf("or authenticating against Azure API - Error unmarshalling response body: %v", err)
	}

	azureClient.accessToken = data["access_token"].(string)
}

func (azureClient *AzureClient) ensureAccessTokenSet() {
	if len(azureClient.accessToken) == 0 {
		azureClient.setAccessToken()
	}
}

func (azureClient *AzureClient) sendHttpMessage(method string, url string) *http.Response {
	azureClient.ensureAccessTokenSet()

	if !strings.HasPrefix(url, "/") {
		url += "/"
	}

	targetUrl := url
	if !strings.HasPrefix(url, "https://") {
		targetUrl = fmt.Sprintf("%s%s", azureClient.environment.armUrl, url)
	}

	log.Infof("Running %s %s\n", method, url)

	log.Debugf("Executing %s %s\n", method, targetUrl)

	request, err := http.NewRequest(method, targetUrl, nil)
	if err != nil {
		log.Fatalf("Error creating HTTP request: %v", err)
	}

	request.Header.Set("Authorization", "Bearer "+azureClient.accessToken)
	response, err := azureClient.client.Do(request)
	if err != nil {
		log.Fatalf("Error sending HTTP request: %v", err)
	}

	log.Debugf("Status code: %d\n", response.StatusCode)

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		log.Errorf("Error sending HTTP request with status code: %d\n", response.StatusCode)

		// If unsuccessful HTTP status code, print out the ARM error response
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Fatalf("Error reading body of response: %v", err)
		}

		prettyPrintJson(body)
	}

	return response
}

func (azureClient *AzureClient) getAzureResources(maxContinuation int) []ArmResource {
	// Invoke Azure Resource Manager resource cache API to find all Azure resources on the subscription
	armResourceSlice := make([]ArmResource, 0)
	targetUrl := fmt.Sprintf("/subscriptions/%s/resources?api-version=2017-08-01", azureClient.config.Credentials.SubscriptionID)

	// Follow nextLink continuation tokens
	i := 0
	for len(targetUrl) > 0 && i <= maxContinuation {

		targetUrl = func(getUrl string) string {
			response := azureClient.sendHttpMessage("GET", getUrl)
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

	return armResourceSlice
}

func prettyPrintJson(body []byte) {
	var jsonElement map[string]interface{}
	err := json.Unmarshal(body, &jsonElement)
	if err != nil {
		log.Fatalf("Error unmarshalling error response body: %v", err)
	}

	prettyPrint, _ := json.MarshalIndent(jsonElement, "", "  ")
	fmt.Println(string(prettyPrint))
}
