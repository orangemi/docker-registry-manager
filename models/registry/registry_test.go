package registry

import (
	"os"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/net/context"
)

var ip string

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	shutdown()
	os.Exit(code)
}

// TestAddRegistry tests the AddRegistry function
func TestAddRegistry(t *testing.T) {

	validRegistryURI := "http://" + ip + ":5000/v2"
	// Create a registy type that contains the expected output from ParseRegistry
	expectedRegistryResponse := Registry{
		Name:   ip,
		Scheme: "http",
		Port:   "5000",
	}

	err := AddRegistry(validRegistryURI)
	Convey("When we attempt to add a valid registry "+validRegistryURI+" we should receive no errors and have the registry cached.", t, func() {
		So(err, ShouldBeNil)
		So(ActiveRegistries[expectedRegistryResponse.Name].Name, ShouldEqual, expectedRegistryResponse.Name)
		So(ActiveRegistries[expectedRegistryResponse.Name].Scheme, ShouldEqual, expectedRegistryResponse.Scheme)
		So(ActiveRegistries[expectedRegistryResponse.Name].Port, ShouldEqual, expectedRegistryResponse.Port)
	})

	reg := ActiveRegistries[expectedRegistryResponse.Name]
	Convey("When we add a valid registry "+validRegistryURI+" the returned URI should equal our expected URI.", t, func() {
		So(reg.GetURI(), ShouldEqual, validRegistryURI)
	})

	Convey("When we add a registry we should be able to get the status of the registry.", t, func() {
		So(reg.GetRegistryStatus(), ShouldBeNil)
	})

	invalidRegistryURI := "192.168.1.2:5000"
	err = AddRegistry(invalidRegistryURI)
	// Test the response error
	Convey("When we pass an invalid RegistryURI we should get back the registry type with errors", t, func() {
		So(err, ShouldNotBeNil)
	})
}

func setup() {

	// TODO: create a new container config definition, pull the image if needed etc before proceeding further

	// Connect to the local docker client
	defaultHeaders := map[string]string{"User-Agent": "engine-api-cli-1.0"}
	cli, err := client.NewClient("unix:///var/run/docker.sock", "v1.24", nil, defaultHeaders)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"Error": err.Error(),
		}).Fatal("Failed to connect to the docker engine api using unix:///var/run/docker.sock")
	}

	// Start the registry container
	if err = cli.ContainerStart(context.Background(), "registry", types.ContainerStartOptions{}); err != nil {
		logrus.WithFields(logrus.Fields{
			"Error": err.Error(),
		}).Fatal("Failed to start the docker registry container")
	}

	// Get the current IP of the container
	containerJSON, err := cli.ContainerInspect(context.Background(), "registry")
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"Error": err.Error(),
		}).Fatal("Failed to inspect the recently created registry to get the IP address")
	}
	ip = containerJSON.NetworkSettings.IPAddress
	time.Sleep(time.Second * 1)

	//containers, _ := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	imgs, _ := cli.ImageList(context.Background(), types.ImageListOptions{})

	for _, img := range imgs {
		err = cli.ImageTag(context.Background(), img.ID, ip+":5000/test1")
		if err != nil {
			logrus.Fatal(err)
		}
		resp, err := cli.ImagePush(context.Background(), ip+":5000/test1", types.ImagePushOptions{RegistryAuth: "test"})
		if err != nil {
			logrus.Fatal(err)
		}
		defer resp.Close()
		if err != nil {
			logrus.Fatal(err)
		}
	}
}

func shutdown() {

	// Connect to the local docker client
	defaultHeaders := map[string]string{"User-Agent": "engine-api-cli-1.0"}
	cli, err := client.NewClient("unix:///var/run/docker.sock", "v1.24", nil, defaultHeaders)
	if err != nil {
		panic(err)
	}

	timeout, err := time.ParseDuration("10s")
	if err = cli.ContainerStop(context.Background(), "registry", &timeout); err != nil {
		panic(err)
	}

}
