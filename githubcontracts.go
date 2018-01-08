package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
)

type GitHubContentItem struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	DownloadUrl string `json:"download_url"`
	Url         string `json:"url"`
}

type GitHubDashboardTemplate struct {
	Name     string
	Contents string
}

const (
	GitHubGrafanaTemplateRootUrl = "https://api.github.com/repos/asheniam/azure-grafana-dashboard-templates/contents/"
)

func getGitHubGrafanaTemplates(resourceType string, resourceKind string) []GitHubDashboardTemplate {

	var httpClient *http.Client = &http.Client{}

	// Generate the root URL for the given ARM resource type.  This is the folder that contains the Grafana dashboard templates
	// for this ARM resource type.
	encodedResourceType := resourceType

	if len(resourceKind) > 0 {
		resourceType = resourceType + "/kind/" + resourceKind
	}

	encodedResourceType = strings.Replace(resourceType, ".", "-", -1)
	encodedResourceType = strings.Replace(encodedResourceType, "/", "-", -1)
	githubUrl := fmt.Sprintf("%s%s?ref=master", GitHubGrafanaTemplateRootUrl, encodedResourceType)

	// Get all the Grafana dashboard template subfolders
	githubContentItems := httpGetGitHubContentItems(httpClient, githubUrl)

	gitHubDashboardTemplates := make([]GitHubDashboardTemplate, 0)

	// For each sub-folder, find the template.json
	for _, githubContentItem := range githubContentItems {
		if strings.EqualFold(githubContentItem.Type, "dir") &&
			len(githubContentItem.Url) > 0 {
			githubContentDashboardFolderItems := httpGetGitHubContentItems(httpClient, githubContentItem.Url)
			for _, githubContentDashboardFolderItem := range githubContentDashboardFolderItems {
				if strings.EqualFold(githubContentDashboardFolderItem.Name, "template.json") {

					// Read the template.json
					templateJson := httpGetGitHubDashboardTemplateJson(httpClient, githubContentDashboardFolderItem.DownloadUrl)
					dashboardTemplate := GitHubDashboardTemplate{
						Name:     githubContentItem.Name,
						Contents: templateJson,
					}
					gitHubDashboardTemplates = append(gitHubDashboardTemplates, dashboardTemplate)
				}
			}
		}
	}

	return gitHubDashboardTemplates
}

func httpGetGitHubContentItems(httpClient *http.Client, targetUrl string) []GitHubContentItem {
	request, err := http.NewRequest("GET", targetUrl, nil)
	if err != nil {
		log.Fatalf("Error creating HTTP request: %v", err)
	}

	log.Debugf("Executing GET %s\n", targetUrl)
	response, err := httpClient.Do(request)
	if err != nil {
		log.Fatalf("Error sending HTTP request: %v", err)
	}

	log.Debugf("Status code: %d\n", response.StatusCode)
	defer response.Body.Close()

	githubContentResponse := make([]GitHubContentItem, 0)
	if response.StatusCode == 200 {
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Fatalf("Error reading body of response: %v", err)
		}

		err = json.Unmarshal(body, &githubContentResponse)
		if err != nil {
			log.Fatalf("Error unmarshalling ARM reource response body: %v", err)
		}
	}

	return githubContentResponse
}

func httpGetGitHubDashboardTemplateJson(httpClient *http.Client, targetUrl string) string {
	request, err := http.NewRequest("GET", targetUrl, nil)
	if err != nil {
		log.Fatalf("Error creating HTTP request: %v", err)
	}

	log.Debugf("Executing GET %s\n", targetUrl)
	response, err := httpClient.Do(request)
	if err != nil {
		log.Fatalf("Error sending HTTP request: %v", err)
	}

	log.Debugf("Status code: %d\n", response.StatusCode)
	defer response.Body.Close()

	templateJson := ""
	if response.StatusCode == 200 {
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Fatalf("Error reading body of response: %v", err)
		}

		templateJson = string(body)
	}

	return templateJson
}
