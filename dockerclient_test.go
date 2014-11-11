package dockerclient

import (
	"bytes"
	"fmt"
	"github.com/docker/docker/pkg/stdcopy"
	"strings"
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

func TestListContainers(t *testing.T) {
	client := testDockerClient(t)
	containers, err := client.ListContainers(true, false)
	if err != nil {
		t.Fatal("cannot get containers: %s", err)
	}
	assertEqual(t, len(containers), 1, "")
	cnt := containers[0]
	assertEqual(t, cnt.SizeRw, 0, "")
}

func TestListContainersWithSize(t *testing.T) {
	client := testDockerClient(t)
	containers, err := client.ListContainers(true, true)
	if err != nil {
		t.Fatal("cannot get containers: %s", err)
	}
	assertEqual(t, len(containers), 1, "")
	cnt := containers[0]
	assertEqual(t, cnt.SizeRw, 123, "")
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
