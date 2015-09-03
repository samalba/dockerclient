package dockerclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/pkg/stdcopy"
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
	assertEqual(t, info.Images, int64(1), "")
	assertEqual(t, info.Containers, int64(2), "")
}

func TestKillContainer(t *testing.T) {
	client := testDockerClient(t)
	if err := client.KillContainer("23132acf2ac", "5"); err != nil {
		t.Fatal("cannot kill container: %s", err)
	}
}

func TestWait(t *testing.T) {
	client := testDockerClient(t)

	// This provokes an error on the server.
	select {
	case wr := <-client.Wait("1234"):
		assertEqual(t, wr.ExitCode, int(-1), "")
	case <-time.After(2 * time.Second):
		t.Fatal("Timed out!")
	}

	// Valid case.
	select {
	case wr := <-client.Wait("valid-id"):
		assertEqual(t, wr.ExitCode, int(0), "")
	case <-time.After(2 * time.Second):
		t.Fatal("Timed out!")
	}
}

func TestPullImage(t *testing.T) {
	client := testDockerClient(t)
	err := client.PullImage("busybox", nil)
	if err != nil {
		t.Fatal("unable to pull busybox")
	}

	err = client.PullImage("haproxy", nil)
	if err != nil {
		t.Fatal("unable to pull haproxy")
	}

	err = client.PullImage("wrongimg", nil)
	if err == nil {
		t.Fatal("should return error when it fails to pull wrongimg")
	}
}

func TestListContainers(t *testing.T) {
	client := testDockerClient(t)
	containers, err := client.ListContainers(true, false, "")
	if err != nil {
		t.Fatal("cannot get containers: %s", err)
	}
	assertEqual(t, len(containers), 1, "")
	cnt := containers[0]
	assertEqual(t, cnt.SizeRw, int64(0), "")
}

func TestContainerChanges(t *testing.T) {
	client := testDockerClient(t)
	changes, err := client.ContainerChanges("foobar")
	if err != nil {
		t.Fatal("cannot get container changes: %s", err)
	}
	assertEqual(t, len(changes), 3, "unexpected number of changes")
	c := changes[0]
	assertEqual(t, c.Path, "/dev", "unexpected")
	assertEqual(t, c.Kind, 0, "unexpected")
}

func TestListContainersWithSize(t *testing.T) {
	client := testDockerClient(t)
	containers, err := client.ListContainers(true, true, "")
	if err != nil {
		t.Fatal("cannot get containers: %s", err)
	}
	assertEqual(t, len(containers), 1, "")
	cnt := containers[0]
	assertEqual(t, cnt.SizeRw, int64(123), "")
}
func TestListContainersWithFilters(t *testing.T) {
	client := testDockerClient(t)
	containers, err := client.ListContainers(true, true, "{'id':['332375cfbc23edb921a21026314c3497674ba8bdcb2c85e0e65ebf2017f688ce']}")
	if err != nil {
		t.Fatal("cannot get containers: %s", err)
	}
	assertEqual(t, len(containers), 1, "")

	containers, err = client.ListContainers(true, true, "{'id':['332375cfbc23edb921a21026314c3497674ba8bdcb2c85e0e65ebf2017f688cf']}")
	if err != nil {
		t.Fatal("cannot get containers: %s", err)
	}
	assertEqual(t, len(containers), 0, "")
}

func TestContainerLogs(t *testing.T) {
	client := testDockerClient(t)
	containerId := "foobar"
	logOptions := &LogOptions{
		Follow:     true,
		Stdout:     true,
		Stderr:     true,
		Timestamps: true,
		Tail:       10,
	}
	logsReader, err := client.ContainerLogs(containerId, logOptions)
	if err != nil {
		t.Fatal("cannot read logs from server")
	}

	stdoutBuffer := new(bytes.Buffer)
	stderrBuffer := new(bytes.Buffer)
	if _, err = stdcopy.StdCopy(stdoutBuffer, stderrBuffer, logsReader); err != nil {
		t.Fatal("cannot read logs from logs reader")
	}
	stdoutLogs := strings.TrimSpace(stdoutBuffer.String())
	stderrLogs := strings.TrimSpace(stderrBuffer.String())
	stdoutLogLines := strings.Split(stdoutLogs, "\n")
	stderrLogLines := strings.Split(stderrLogs, "\n")
	if len(stdoutLogLines) != 5 {
		t.Fatalf("wrong number of stdout logs: len=%d", len(stdoutLogLines))
	}
	if len(stderrLogLines) != 5 {
		t.Fatalf("wrong number of stderr logs: len=%d", len(stdoutLogLines))
	}
	for i, line := range stdoutLogLines {
		expectedSuffix := fmt.Sprintf("Z line %d", 41+2*i)
		if !strings.HasSuffix(line, expectedSuffix) {
			t.Fatalf("expected stdout log line \"%s\" to end with \"%s\"", line, expectedSuffix)
		}
	}
	for i, line := range stderrLogLines {
		expectedSuffix := fmt.Sprintf("Z line %d", 40+2*i)
		if !strings.HasSuffix(line, expectedSuffix) {
			t.Fatalf("expected stderr log line \"%s\" to end with \"%s\"", line, expectedSuffix)
		}
	}
}

func TestMonitorEvents(t *testing.T) {
	client := testDockerClient(t)
	decoder := json.NewDecoder(bytes.NewBufferString(eventsResp))
	var expectedEvents []Event
	for {
		var event Event
		if err := decoder.Decode(&event); err != nil {
			if err == io.EOF {
				break
			} else {
				t.Fatalf("cannot parse expected resp: %s", err.Error())
			}
		} else {
			expectedEvents = append(expectedEvents, event)
		}
	}

	// test passing stop chan
	stopChan := make(chan struct{})
	eventInfoChan, err := client.MonitorEvents(nil, stopChan)
	if err != nil {
		t.Fatalf("cannot get events from server: %s", err.Error())
	}

	eventInfo := <-eventInfoChan
	if eventInfo.Error != nil || eventInfo.Event != expectedEvents[0] {
		t.Fatalf("got:\n%#v\nexpected:\n%#v", eventInfo, expectedEvents[0])
	}
	close(stopChan)
	for i := 0; i < 3; i++ {
		_, ok := <-eventInfoChan
		if i == 2 && ok {
			t.Fatalf("read more than 2 events successfully after closing stopChan")
		}
	}

	// test when you don't pass stop chan
	eventInfoChan, err = client.MonitorEvents(nil, nil)
	if err != nil {
		t.Fatalf("cannot get events from server: %s", err.Error())
	}

	for i, expectedEvent := range expectedEvents {
		t.Logf("on iter %d\n", i)
		eventInfo := <-eventInfoChan
		if eventInfo.Error != nil || eventInfo.Event != expectedEvent {
			t.Fatalf("index %d, got:\n%#v\nexpected:\n%#v", i, eventInfo, expectedEvent)
		}
		t.Logf("done with iter %d\n", i)
	}
}

func TestDockerClientInterface(t *testing.T) {
	iface := reflect.TypeOf((*Client)(nil)).Elem()
	test := testDockerClient(t)

	if !reflect.TypeOf(test).Implements(iface) {
		t.Fatalf("DockerClient does not implement the Client interface")
	}
}

func TestLibnetwork(t *testing.T) {
	docker, _ := NewDockerClient("unix:///var/run/docker.sock", nil)

	// Create Network
	netConfig := &NetworkConfig{
		Name:        "netfoo",
		NetworkType: "overlay",
	}
	netId, err := docker.CreateNetwork(netConfig)
	if err != nil {
		t.Fatalf("cannot create network 'netfoo': %v", err)
	}
	if netId == "" {
		t.Fatalf("CreateEndpoint should return a non empty value for the network Id")
	}

	// List Networks
	networks, err := docker.ListNetworks("", "")
	if err != nil {
		t.Fatalf("cannot list network: %v", err)
	}
	if networks == nil {
		t.Fatalf("ListNetworks should return a non nil network list")
	}
	for _, n := range networks {
		if n.Id == "" {
			t.Fatalf("invalid entry for network")
		}
	}

	// Get Network
	network, err := docker.GetNetwork(netId)
	if err != nil {
		t.Fatalf("cannot get network 'netfoo': %v", err)
	}
	if network == nil {
		t.Fatalf("GetNetwork should return a non nil network object")
	}

	// Create Endpoint
	endpointConfig := &EndpointConfig{
		Name: "myendpoint",
		ExposedPorts: []TransportPort{
			TransportPort{
				Proto: 6,
				Port:  7890,
			},
		},
	}
	eId, err := docker.CreateEndpoint(netId, endpointConfig)
	if err != nil {
		t.Fatalf("cannot create endpoint myendpoint for network 'netfoo': %v", err)
	}
	if eId == "" {
		t.Fatalf("CreateEndpoint should return a non empty value for the endpoint Id")
	}

	// List Endpoints
	endpoints, err := docker.ListEndpoints(netId)
	if err != nil {
		t.Fatalf("cannot list endpoints for 'netfoo': %s", err)
	}
	if len(endpoints) == 0 {
		t.Fatalf("endpoint list is empty for ListEndpoints, should contain at least one entry")
	}

	// Get Network with endpoint
	network, err = docker.GetNetwork(netId)
	if err != nil {
		t.Fatalf("cannot get network 'netfoo': %v", err)
	}
	if network == nil {
		t.Fatalf("network is nil")
	}
	if len(network.Endpoints) == 0 {
		t.Fatalf("endpoint list is empty for GetNetwork, should contain at least one entry")
	}

	// Get Endpoint
	endpoint, err := docker.GetEndpoint(netId, eId)
	if err != nil {
		t.Fatalf("cannot get endpoint myendpoint: %s", err)
	}
	if endpoint == nil {
		t.Fatalf("endpoint is nil")
	}

	// Create a container
	containerConfig := &ContainerConfig{
		Image:       "ubuntu:14.04",
		Cmd:         []string{"bash"},
		AttachStdin: true,
		Tty:         true,
	}
	containerId, err := docker.CreateContainer(containerConfig, "bar")
	if err != nil {
		t.Fatalf("CreateContainer should successfully create a container: %v", err)
	}
	if containerId == "" {
		t.Fatalf("container ID should not be empty")
	}

	// Attach endpoint to container
	joinConfig := &JoinConfig{
		ContainerID:       containerId,
		UseDefaultSandbox: true,
	}
	joinStr, err := docker.JoinEndpoint(netId, eId, joinConfig)
	if err != nil {
		t.Fatalf("JoinEndpoint should successfully join the container to the endpoint: %v", err)
	}
	if joinStr == "" {
		t.Fatalf("result of JoinEndpoint should not be empty")
	}

	// Start the container
	hostConfig := &HostConfig{}
	err = docker.StartContainer(containerId, hostConfig)
	if err != nil {
		t.Fatal("The container should start successfully", err)
	}

	// Publish Service
	serviceConfig := &ServiceConfig{
		Name:        "myservice",
		NetworkName: network.Name,
	}
	sId, err := docker.PublishService(serviceConfig)
	if err != nil {
		t.Fatalf("cannot publish service myservice: %s", err)
	}
	if sId == "" {
		t.Fatalf("service Id should not be empty")
	}

	// Attach Backend to Service
	joinConfig = &JoinConfig{
		ContainerID:       containerId,
		UseDefaultSandbox: true,
	}
	attachStr, err := docker.AttachBackendToService(sId, joinConfig)
	if err != nil {
		t.Fatalf("cannot attach backend to service 'myservice': %v", err)
	}
	if attachStr == "" {
		t.Fatalf("result of AttachBackendToService should not be empty")
	}

	// List Service Backends
	backends, err := docker.ListBackends(sId)
	if err != nil {
		t.Fatalf("cannot list backends of service 'myservice': %v", err)
	}
	for _, b := range backends {
		if b.Id == "" {
			t.Fatalf("incorrect entry for backends list in return of ListBackends")
		}
	}

	// Detach Backend from Service
	err = docker.DetachBackendFromService(sId, containerId)
	if err != nil {
		t.Fatalf("cannnot detach backend from service 'myservice': %v", err)
	}

	// Unpublish Service
	err = docker.UnpublishService(sId)
	if err != nil {
		t.Fatalf("cannot unpublish service 'myservice': %v", err)
	}

	// Detach endpoint from the container
	err = docker.DetachEndpoint(netId, eId, containerId)
	if err != nil {
		t.Fatalf("cannot detach endpoint 'myendpoint' from network 'netfoo': %v", err)
	}

	// Delete Endpoint
	err = docker.DeleteEndpoint(netId, eId)
	if err != nil {
		t.Fatalf("cannot delete endpoint 'myendpoint': %v", err)
	}

	// Delete Network
	err = docker.DeleteNetwork(netId)
	if err != nil {
		t.Fatalf("cannot delete network 'netfoo': %v", err)
	}

	// Delete Container
	err = docker.RemoveContainer(containerId, true, false)
	if err != nil {
		t.Fatalf("failed to remove container: %v", err)
	}
}
