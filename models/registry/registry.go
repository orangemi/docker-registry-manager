package registry

import (
	"net"
	"net/http"
	"net/url"
	"sync"

	"github.com/Sirupsen/logrus"
	_ "github.com/go-sql-driver/mysql" // need to initialize mysql before making a connection
	"github.com/stefannaglee/docker-registry-manager/utilities"
)

// ActiveRegistries contains a map of all active registries identified by their name
var ActiveRegistries map[string]*Registry

func init() {
	// Create the active registries map
	ActiveRegistries = make(map[string]*Registry, 0)
}

// Registry contains all identifying information for communicating with a registry
type Registry struct {
	Name    string
	IP      string
	Scheme  string
	Port    string
	Version string

	Metadata
}

// Metadata contains registry metadata information such as the current repository count and number of tags
type Metadata struct {
	Status           string
	RepoCount        int
	TagCount         int
	RepoTotalSize    int64
	RepoTotalSizeStr string

	sync.Mutex
}

// GetURI returns the full url path for communicating with this registry
func (r *Registry) GetURI() string {
	return r.Scheme + "://" + r.Name + ":" + r.Port + "/v2"
}

// GetRegistryStatus takes in a registry URL and checks for communication errors
//
// Create and execute basic GET request to test if each registry can be reached
// To determine registry status we test the base registry route of /v2/ and check
// the HTTP response code for a 200 response (200 is a successful request)
func (r *Registry) GetRegistryStatus() error {

	// Create and execute a plain get request and check the http status code
	// If we don't receive a 200 response code or get an hour, we cannot communicate with the registry
	if response, err := http.Get(r.GetURI()); err != nil || response.StatusCode != 200 {
		utils.Log.WithFields(logrus.Fields{
			"Registry URL":  r.GetURI(),
			"Error":         err,
			"HTTP Response": response,
			"Possible Fix":  "Check to see if your registry is up, and serving on the correct port with 'docker ps'.",
		}).Error("Get request to registry timed out/failed! Is the URL correct, and is the registry active?")
		r.Lock()
		r.Status = "unavailable"
		r.Unlock()
		return err
	}

	// We've successfully connected to the registry so mark it as available
	r.Lock()
	r.Status = "available"
	r.Unlock()
	return nil
}

// AddRegistry takes in a registry URI string and converts it into a registry object
func AddRegistry(registryURI string) error {

	// Create an empty Registry
	r := Registry{
		Version: "v2",
	}

	// Parse the URL and get the scheme
	// e.g https, http, etc.
	u, err := url.Parse(registryURI)
	if err != nil {
		utils.Log.Error(err)
		return err
	}

	// Set scheme
	r.Scheme = u.Scheme

	// Get the host and port
	// e.g test.domain.com and 5000, etc.
	host, port, err := net.SplitHostPort(u.Host)
	if err != nil {
		utils.Log.Error(err)
		return err
	}

	// Set name and port
	r.Name = host
	r.Port = port

	// Lookup the ip for the passed host
	// Using the host name try looking up the IP for informational purposes
	ip, err := net.LookupHost(host)
	if err != nil {
		utils.Log.Error(err)
		// We do not need to return an error since we don't "need" the IP of the host
	}
	// Set IP if we have it
	if ip != nil && len(ip) > 0 {
		r.IP = ip[0]
	}

	// Add the registry to the map of registries
	ActiveRegistries[r.Name] = &r

	return nil
}
