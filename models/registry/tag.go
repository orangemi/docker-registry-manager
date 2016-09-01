package registry

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/pivotal-golang/bytefmt"
	"github.com/stefannaglee/docker-registry-manager/utilities"
)

// Tags contains a slice of tags for the given repository
type Tags struct {
	Name string
	Tags []Tag
}

type Tag struct {
	ID              string
	Name            string
	UpdatedTime     time.Time
	UpdatedTimeUnix int64
	TimeAgo         string
	Layers          int
	Size            string
	SizeInt         int64
}

// GetTags returns the tags for the given repository
// https://github.com/docker/distribution/blob/master/docs/spec/api.md#listing-image-tags
func GetTags(registry *Registry, repository *Repository) (Tags, error) {

	ts := Tags{}

	// Create and execute Get request
	response, err := http.Get(registry.GetURI() + "/" + repository.Name + "/tags/list")
	defer response.Body.Close()
	if err != nil {
		utils.Log.WithFields(logrus.Fields{
			"Registry URL": registry.GetURI(),
			"Error":        err,
			"Possible Fix": "Check to see if your registry is up, and serving on the correct port with 'docker ps'. ",
		}).Error("Get request to registry failed for the tags endpoint.")
		return ts, err
	} else if response.StatusCode != 200 {
		utils.Log.WithFields(logrus.Fields{
			"Error":       err,
			"Status Code": response.StatusCode,
			"Response":    response,
		}).Error("Did not receive an ok status code!")
		return ts, err
	}

	// Read response into byte body
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		utils.Log.WithFields(logrus.Fields{
			"Error": err,
			"Body":  body,
		}).Error("Unable to read response into body!")
		return ts, err
	}

	// Unmarshal JSON into the tag response struct containing a slice of tags
	if err := json.Unmarshal(body, &ts); err != nil {
		utils.Log.WithFields(logrus.Fields{
			"Error":         err,
			"Response Body": string(body),
		}).Error("Unable to unmarshal JSON!")
		return ts, err
	}

	// Get the tag metadata information
	tagChan := make(chan *Tag)
	for _, tag := range ts.Tags {
		go func(tag *Tag) {
			// Created a new tag for view type to fill
			var tempSize int64
			var maxTime time.Time

			// Get the image information for each tag
			img, _ := GetImage(registry, repository, tag)

			for _, layer := range img.FsLayers {
				// Create and execute Get request
				response, _ := http.Head(registry.GetURI() + "/" + repository.Name + "/blobs/" + layer.BlobSum)
				if err != nil {
					utils.Log.Error(err)
				}
				tempSize += response.ContentLength
			}
			// Get the latest creation time and total the size for the tag image
			for _, history := range img.History {
				if history.V1Compatibility.Created.After(maxTime) {
					maxTime = history.V1Compatibility.Created
				}
			}

			// Set the fields
			tag.Size = bytefmt.ByteSize(uint64(tempSize))
			tag.SizeInt = tempSize
			tag.UpdatedTime = maxTime
			tag.UpdatedTimeUnix = maxTime.Unix()
			tag.Layers = len(img.History)
			tag.TimeAgo = utils.TimeAgo(maxTime)

			tagChan <- tag
		}(&tag)

	}

	// Wait for each of the requests and append to the returned tag information
	for i := 0; i < len(ts.Tags); i++ {
		<-tagChan
	}
	close(tagChan)

	return ts, nil
}

// DeleteTag deletes the tag by first getting its docker-content-digest, and then using
// the digest received the function deletes the manifest
//
// Documentation:
// DELETE	/v2/<name>/manifests/<reference>	Manifest	Delete the manifest identified by name and reference. Note that a manifest can only be deleted by digest.
func DeleteTag(registry *Registry, repository *Repository, tag Tag) error {

	// Note When deleting a manifest from a registry version 2.3 or later, the following header must be used when HEAD or GET-ing the manifest to obtain the correct digest to delete:
	// Accept: application/vnd.docker.distribution.manifest.v2+json
	client := &http.Client{}
	req, _ := http.NewRequest("HEAD", registry.GetURI()+"/"+repository.Name+"/manifests/"+tag.Name, nil)
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")

	// Execute the request
	resp, existsErr := client.Do(req)
	if existsErr != nil {
		utils.Log.WithFields(logrus.Fields{
			"Request":  resp.Request,
			"Error":    existsErr,
			"Tag":      tag,
			"Response": resp,
		}).Error("Could not delete tag! Could not head the tag.")
		return existsErr
	}

	// Make sure the digest exists in the header. If it does, attempt the deletion
	if _, ok := resp.Header["Docker-Content-Digest"]; ok {

		if len(resp.Header["Docker-Content-Digest"]) > 0 {
			// Create and execute DELETE request
			digest := resp.Header["Docker-Content-Digest"][0]
			client := &http.Client{}
			req, _ := http.NewRequest("DELETE", registry.GetURI()+"/"+repository.Name+"/manifests/"+digest, nil)
			req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")
			resp, err := client.Do(req)
			if err != nil || resp.StatusCode != 200 {
				utils.Log.WithFields(logrus.Fields{
					"Error":    err,
					"Tag":      tag,
					"Response": resp,
				}).Error("Could not delete tag!")
				return err
			}
		}

		// Error if there was nothing in the Docker-Content-Digest field
		utils.Log.WithFields(logrus.Fields{
			"Error":    errors.New("No digest gotten from response header"),
			"Tag":      tag,
			"Response": resp,
		}).Error("Could not delete tag!")
		return errors.New("No digest gotten from response header")
	}

	return nil
}
