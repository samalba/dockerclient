package dockerclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"sync/atomic"
	"time"
)

type DockerClient struct {
	URL           *url.URL
	HTTPClient    *http.Client
	monitorEvents int32
}

func NewDockerClient(daemonUrl string) (*DockerClient, error) {
	u, err := url.Parse(daemonUrl)
	if err != nil {
		return nil, err
	}
	httpClient := newHTTPClient(u)
	return &DockerClient{u, httpClient, 0}, nil
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
	b := bytes.NewBuffer(body)
	req, err := http.NewRequest(method, client.URL.String()+path, b)
	if err != nil {
		return nil, err
	}
	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("%s: %s", resp.Status, data)
	}
	return data, nil
}

func (client *DockerClient) ListContainers(all bool) (*[]Container, error) {
	argAll := 0
	if all == true {
		argAll = 1
	}
	args := fmt.Sprintf("?all=%d", argAll)
	data, err := client.doRequest("GET", "/v1.8/containers/json"+args, nil)
	if err != nil {
		return nil, err
	}
	ret := &[]Container{}
	err = json.Unmarshal(data, ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (client *DockerClient) InspectContainer(id string) (*ContainerInfo, error) {
	uri := fmt.Sprintf("/v1.8/containers/%s/json", id)
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

func (client *DockerClient) CreateContainer(config *ContainerConfig) (string, error) {
	data, err := json.Marshal(config)
	if err != nil {
		return "", err
	}
	uri := "/v1.8/containers/create"
	data, err = client.doRequest("POST", uri, data)
	if err != nil {
		return "", err
	}
	fmt.Println(string(data))
	result := make(map[string]string)
	err = json.Unmarshal(data, &result)
	if err != nil {
		return "", err
	}
	return result["Id"], nil
}

func (client *DockerClient) StartContainer(id string) error {
	uri := fmt.Sprintf("/v1.8/containers/%s/start", id)
	_, err := client.doRequest("POST", uri, nil)
	if err != nil {
		return err
	}
	return nil
}

func (client *DockerClient) StopContainer(id string, timeout int) error {
	uri := fmt.Sprintf("/v1.8/containers/%s/stop?t=%d", id, timeout)
	_, err := client.doRequest("POST", uri, nil)
	if err != nil {
		return err
	}
	return nil
}

func (client *DockerClient) RestartContainer(id string, timeout int) error {
	uri := fmt.Sprintf("/v1.8/containers/%s/restart?t=%d", id, timeout)
	_, err := client.doRequest("POST", uri, nil)
	if err != nil {
		return err
	}
	return nil
}

func (client *DockerClient) KillContainer(id string) error {
	uri := fmt.Sprintf("/v1.8/containers/%s/kill", id)
	_, err := client.doRequest("POST", uri, nil)
	if err != nil {
		return err
	}
	return nil
}

func (client *DockerClient) StartMonitorEvents(cb func(*Event)) {
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
			uri := client.URL.String() + "/v1.8/events"
			resp, err := client.HTTPClient.Get(uri)
			fmt.Println("New Request")
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
				_, err = resp.Body.Read(buffer)
				if err != nil {
					resp.Body.Close()
					time.Sleep(wait)
					break
				}
				event := &Event{}
				fmt.Println(string(buffer))
				err = json.Unmarshal(buffer, event)
				if err == nil {
					cb(event)
				}
			}
			time.Sleep(wait)
		}
	}()
}

func (client *DockerClient) StopAllMonitorEvents() {
	atomic.StoreInt32(&client.monitorEvents, 0)
}

func (client *DockerClient) Version() (*Version, error) {
	data, err := client.doRequest("GET", "/v1.8/version", nil)
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
