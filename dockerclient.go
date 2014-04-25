package dockerclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"sync/atomic"
	"time"
)

const (
	DockerBaseURL = "/v1.10"
)

type DockerClient struct {
	URL           *url.URL
	HTTPClient    *http.Client
	Debug         bool
	monitorEvents int32
}

// Return a new dockerclient for use in subsequent calls to the remote Docker API.
func NewDockerClient(daemonUrl string) (*DockerClient, error) {
	u, err := url.Parse(daemonUrl)
	if err != nil {
		return nil, err
	}
	httpClient := newHTTPClient(u)
	return &DockerClient{u, httpClient, false, 0}, nil
}

func newHTTPClient(u *url.URL) *http.Client {
	httpTransport := &http.Transport{}
	if u.Scheme == "unix" {
		socketPath := u.Path
		unixDial := func(proto string, addr string) (net.Conn, error) {
			return net.Dial("unix", socketPath)
		}
		httpTransport.Dial = unixDial
		// Override the main URL object so the HTTP lib won't complain
		u.Scheme = "http"
		u.Host = "unix.sock"
	}
	u.Path = ""
	return &http.Client{Transport: httpTransport}
}

func (client *DockerClient) doRequest(method string, path string, body []byte) ([]byte, error) {
	if client.Debug {
		bodyStr := ""
		if body != nil {
			bodyStr = string(body)
		}
		log.Printf("doRequest: path, body:\n%s\n%s", path, bodyStr)
	}

	b := bytes.NewBuffer(body)
	req, err := http.NewRequest(method, client.URL.String()+path, b)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	} else if client.Debug {
		log.Printf("doRequest: response data: %s\n", data)
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("%s: %s", resp.Status, data)
	}
	return data, nil
}

// List all containers on the Docker host, including those stopped if 'all' is true.
func (client *DockerClient) ListContainers(all bool) ([]Container, error) {
	argAll := 0
	if all == true {
		argAll = 1
	}
	args := fmt.Sprintf("?all=%d", argAll)
	data, err := client.doRequest("GET", DockerBaseURL+"/containers/json"+args, nil)
	if err != nil {
		return nil, err
	}
	ret := []Container{}
	err = json.Unmarshal(data, &ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

// Return all information about a container by id.
func (client *DockerClient) InspectContainer(id string) (*ContainerInfo, error) {
	uri := fmt.Sprintf(DockerBaseURL+"/containers/%s/json", id)
	data, err := client.doRequest("GET", uri, nil)
	if err != nil {
		return nil, err
	}
	info := &ContainerInfo{}
	err = json.Unmarshal(data, info)
	if err != nil {
		return nil, err
	}
	return info, nil
}

// Create a container based on the provided CongainerConfig.
func (client *DockerClient) CreateContainer(config *ContainerConfig) (string, error) {
	data, err := json.Marshal(config)
	if err != nil {
		return "", err
	}

	uri := DockerBaseURL + "/containers/create"
	data, err = client.doRequest("POST", uri, data)
	if err != nil {
		return "", err
	}
	result := &RespContainersCreate{}
	err = json.Unmarshal(data, result)
	if err != nil {
		return "", err
	}
	return result.Id, nil
}

// Start a container with HostConfig 'host.'
func (client *DockerClient) StartContainer(id string, config *HostConfig) error {
	data, err := json.Marshal(config)
	if err != nil {
		return err
	}

	uri := fmt.Sprintf(DockerBaseURL+"/containers/%s/start", id)
	_, err = client.doRequest("POST", uri, data)
	if err != nil {
		return err
	}
	return nil
}

// Stop a container by id with timeout.
func (client *DockerClient) StopContainer(id string, timeout int) error {
	uri := fmt.Sprintf(DockerBaseURL+"/containers/%s/stop?t=%d", id, timeout)
	_, err := client.doRequest("POST", uri, nil)
	if err != nil {
		return err
	}
	return nil
}

// Restart a container by id with timeout.
func (client *DockerClient) RestartContainer(id string, timeout int) error {
	uri := fmt.Sprintf(DockerBaseURL+"/containers/%s/restart?t=%d", id, timeout)
	_, err := client.doRequest("POST", uri, nil)
	if err != nil {
		return err
	}
	return nil
}

// Kill container by id.
func (client *DockerClient) KillContainer(id string) error {
	uri := fmt.Sprintf(DockerBaseURL+"/containers/%s/kill", id)
	_, err := client.doRequest("POST", uri, nil)
	if err != nil {
		return err
	}
	return nil
}

// Start monitoring the Docker service for events, calling the given callback for each.
func (client *DockerClient) StartMonitorEvents(cb func(*Event, ...interface{}), args ...interface{}) {
	atomic.StoreInt32(&client.monitorEvents, 1)
	wait := 100 * time.Millisecond
	buffer := make([]byte, 4096)
	var running int32 = 1
	go func() {
		for running > 0 {
			running = atomic.LoadInt32(&client.monitorEvents)
			if running == 0 {
				break
			}
			uri := client.URL.String() + DockerBaseURL + "/events"
			resp, err := client.HTTPClient.Get(uri)
			if err != nil {
				time.Sleep(wait)
				continue
			}
			if resp.StatusCode >= 300 {
				resp.Body.Close()
				time.Sleep(wait)
				continue
			}
			for {
				nBytes, err := resp.Body.Read(buffer)
				if err != nil {
					resp.Body.Close()
					time.Sleep(wait)
					break
				}
				event := &Event{}
				err = json.Unmarshal(buffer[:nBytes], event)
				if err == nil {
					cb(event, args...)
				}
			}
			time.Sleep(wait)
		}
	}()
}

// Stop monitoring the Docker service for events.
func (client *DockerClient) StopAllMonitorEvents() {
	atomic.StoreInt32(&client.monitorEvents, 0)
}

// Get the Docker version.
func (client *DockerClient) Version() (*Version, error) {
	data, err := client.doRequest("GET", DockerBaseURL+"/version", nil)
	if err != nil {
		return nil, err
	}
	version := &Version{}
	err = json.Unmarshal(data, version)
	if err != nil {
		return nil, err
	}
	return version, nil
}

// Get the Docker version.
func (client *DockerClient) RemoveContainer(id string) error {
	_, err := client.doRequest("DELETE", DockerBaseURL+"/containers/"+id, nil)
	return err
}
