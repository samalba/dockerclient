package dockerclient

import (
	"fmt"
	"testing"
)

func assertEqual(t *testing.T, a interface{}, b interface{}, message string) {
	if a == b {
		return
	}
	if len(message) == 0 {
		message = fmt.Sprintf("%v != %v", a, b)
	}
	t.Fatal(message)
}

func testDockerClient(t *testing.T) *DockerClient {
	client, err := NewDockerClient(testHTTPServer.URL, nil)
	if err != nil {
		t.Fatal("Cannot init the docker client")
	}
	return client
}

func TestInfo(t *testing.T) {
	client := testDockerClient(t)
	info, err := client.Info()
	if err != nil {
		t.Fatal("Cannot get server info")
	}
	assertEqual(t, info.Images, 1, "")
	assertEqual(t, info.Containers, 2, "")
}
