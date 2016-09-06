package registry

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/Sirupsen/logrus"
	"github.com/stefannaglee/docker-registry-manager/utilities"
)

// Repository contains information on the name and encoded name
type Repository struct {
	Name       string
	EncodedURI string
	TagCount   int

	Tags []Tag
}

// GetRepositories returns a slice of repositories with their names and encoded names
func GetRepositories(registry *Registry) ([]Repository, error) {
	// Create and execute Get request for the catalog of repositores
	// https://github.com/docker/distribution/blob/master/docs/spec/api.md#catalog
	response, err := http.Get(registry.GetURI() + "/_catalog")
	defer response.Body.Close()
	if err != nil {
		utils.Log.WithFields(logrus.Fields{
			"Registry URL": string(registry.GetURI()),
			"Error":        err,
			"Possible Fix": "Check to see if your registry is up, and serving on the correct port with 'docker ps'. ",
		}).Error("Get request to registry failed for the /_catalog endpoint! Is your registry active?")
		return nil, err
	} else if response.StatusCode != 200 {
		utils.Log.WithFields(logrus.Fields{
			"Error":       err,
			"Status Code": response.StatusCode,
			"Response":    response,
		}).Error("Did not receive an ok status code!")
		return nil, err
	}

	// Read response into byte body
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		utils.Log.WithFields(logrus.Fields{
			"Error": err,
			"Body":  body,
		}).Error("Unable to read response into body!")
	}

	// Unmarshal JSON into a slice of repository names
	var repositoryNames map[string][]string
	if err := json.Unmarshal(body, &repositoryNames); err != nil {
		utils.Log.WithFields(logrus.Fields{
			"Error":         err,
			"Response Body": string(body),
		}).Error("Unable to unmarshal JSON!")
		return nil, err
	}

	// Create the new repository objects keyed off of their name
	var repositories []Repository
	for _, name := range repositoryNames["repositories"] {
		r := Repository{
			EncodedURI: url.QueryEscape(name),
			Name:       name,
		}
		repositories = append(repositories, r)
	}

	// Update the registry metadata information
	registry.RepoCount = len(repositoryNames)

	return repositories, nil
}
