package registry

import (
	"os"
	"testing"
	"time"

	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/net/context"
)

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	shutdown()
	os.Exit(code)
}

// TestAddRegistry tests the AddRegistry function
func TestParseRegistry(t *testing.T) {

	validRegistryURI := "https://host.domain.com:5000"
	// Create a registy type that contains the expected output from ParseRegistry
	expectedRegistryResponse := Registry{
		Name:   "host.domain.com",
		Scheme: "https",
		Port:   "5000",
	}

	err := AddRegistry(validRegistryURI)
	Convey("When we attempt to add a valid registry "+validRegistryURI+" we should receive no errors and have the registry cached.", t, func() {
		So(err, ShouldBeNil)
		So(ActiveRegistries[expectedRegistryResponse.Name].Name, ShouldEqual, expectedRegistryResponse.Name)
		So(ActiveRegistries[expectedRegistryResponse.Name].Scheme, ShouldEqual, expectedRegistryResponse.Scheme)
		So(ActiveRegistries[expectedRegistryResponse.Name].Port, ShouldEqual, expectedRegistryResponse.Port)
	})

	invalidRegistryURI := "192.168.1.2:5000"
	err = AddRegistry(invalidRegistryURI)
	// Test the response error
	Convey("When we pass an invalid RegistryURI we should get back the registry type with errors", t, func() {
		So(err, ShouldNotBeNil)
	})
}

func setup() {

	// Connect to the local docker client
	defaultHeaders := map[string]string{"User-Agent": "engine-api-cli-1.0"}
	cli, err := client.NewClient("unix:///var/run/docker.sock", "v1.22", nil, defaultHeaders)
	if err != nil {
		panic(err)
	}

	if err = cli.ContainerStart(context.Background(), "registry", types.ContainerStartOptions{}); err != nil {
		panic(err)
	}
}

func shutdown() {

	// Connect to the local docker client
	defaultHeaders := map[string]string{"User-Agent": "engine-api-cli-1.0"}
	cli, err := client.NewClient("unix:///var/run/docker.sock", "v1.22", nil, defaultHeaders)
	if err != nil {
		panic(err)
	}

	timeout, err := time.ParseDuration("10s")

	if err = cli.ContainerStop(context.Background(), "registry", &timeout); err != nil {
		panic(err)
	}

}
