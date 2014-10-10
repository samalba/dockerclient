package dockerclient

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sync/atomic"
)

var (
	ErrNotFound = errors.New("Not found")
)

type DockerClient struct {
	URL           *url.URL
	HTTPClient    *http.Client
	monitorEvents int32
}

type Callback func(*Event, ...interface{})

type Error struct {
	StatusCode int
	Status     string
	msg        string
}

func (e Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Status, e.msg)
}

func NewDockerClient(daemonUrl string, tlsConfig *tls.Config) (*DockerClient, error) {
	u, err := url.Parse(daemonUrl)
	if err != nil {
		return nil, err
	}
	if u.Scheme == "tcp" {
		u.Scheme = "http"
	}
	httpClient := newHTTPClient(u, tlsConfig)
	return &DockerClient{u, httpClient, 0}, nil
}

func (client *DockerClient) doRequest(method string, path string, body []byte) ([]byte, error) {
	b := bytes.NewBuffer(body)
	req, err := http.NewRequest(method, client.URL.String()+path, b)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 404 {
		return nil, ErrNotFound
	}
	if resp.StatusCode >= 400 {
		return nil, Error{StatusCode: resp.StatusCode, Status: resp.Status, msg: string(data)}
	}
	return data, nil
}

func (client *DockerClient) Info() (*Info, error) {
	data, err := client.doRequest("GET", "/v1.10/info", nil)
	if err != nil {
		return nil, err
	}
	ret := &Info{}
	err = json.Unmarshal(data, &ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (client *DockerClient) ListContainers(all bool) ([]Container, error) {
	argAll := 0
	if all == true {
		argAll = 1
	}
	args := fmt.Sprintf("?all=%d", argAll)
	data, err := client.doRequest("GET", "/v1.10/containers/json"+args, nil)
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

func (client *DockerClient) InspectContainer(id string) (*ContainerInfo, error) {
	uri := fmt.Sprintf("/v1.10/containers/%s/json", id)
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

func (client *DockerClient) CreateContainer(config *ContainerConfig, name string) (string, error) {
	data, err := json.Marshal(config)
	if err != nil {
		return "", err
	}
	uri := "/v1.10/containers/create"

	if name != "" {
		v := url.Values{}
		v.Set("name", name)
		uri = fmt.Sprintf("%s?%s", uri, v.Encode())
	}
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

func (client *DockerClient) StartContainer(id string, config *HostConfig) error {
	data, err := json.Marshal(config)
	if err != nil {
		return err
	}
	uri := fmt.Sprintf("/v1.10/containers/%s/start", id)
	_, err = client.doRequest("POST", uri, data)
	if err != nil {
		return err
	}
	return nil
}

func (client *DockerClient) StopContainer(id string, timeout int) error {
	uri := fmt.Sprintf("/v1.10/containers/%s/stop?t=%d", id, timeout)
	_, err := client.doRequest("POST", uri, nil)
	if err != nil {
		return err
	}
	return nil
}

func (client *DockerClient) RestartContainer(id string, timeout int) error {
	uri := fmt.Sprintf("/v1.10/containers/%s/restart?t=%d", id, timeout)
	_, err := client.doRequest("POST", uri, nil)
	if err != nil {
		return err
	}
	return nil
}

func (client *DockerClient) KillContainer(id string) error {
	uri := fmt.Sprintf("/v1.10/containers/%s/kill", id)
	_, err := client.doRequest("POST", uri, nil)
	if err != nil {
		return err
	}
	return nil
}

func (client *DockerClient) StartMonitorEvents(cb Callback, args ...interface{}) {
	atomic.StoreInt32(&client.monitorEvents, 1)
	go client.getEvents(cb, args...)
}

func (client *DockerClient) getEvents(cb Callback, args ...interface{}) {
	uri := client.URL.String() + "/v1.10/events"
	resp, err := client.HTTPClient.Get(uri)
	if err != nil {
		log.Printf("GET %s failed: %v", uri, err)
		return
	}
	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)
	for atomic.LoadInt32(&client.monitorEvents) > 0 {
		var event *Event
		if err := dec.Decode(&event); err != nil {
			log.Printf("Event decoding failed: %v", err)
			return
		}
		cb(event, args...)
	}
}

func (client *DockerClient) StopAllMonitorEvents() {
	atomic.StoreInt32(&client.monitorEvents, 0)
}

func (client *DockerClient) Version() (*Version, error) {
	data, err := client.doRequest("GET", "/v1.10/version", nil)
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

func (client *DockerClient) PullImage(name, tag string) error {
	v := url.Values{}
	v.Set("fromImage", name)
	if tag != "" {
		v.Set("tag", tag)
	}
	_, err := client.doRequest("POST", "/v1.10/images/create?"+v.Encode(), nil)
	return err
}

func (client *DockerClient) RemoveContainer(id string, force bool) error {
	argForce := 0
	if force == true {
		argForce = 1
	}
	args := fmt.Sprintf("force=%d", argForce)

	_, err := client.doRequest("DELETE", fmt.Sprintf("/v1.10/containers/%s?%s", id, args), nil)
	return err
}

func (client *DockerClient) ListImages() ([]*Image, error) {
	data, err := client.doRequest("GET", "/v1.10/images/json", nil)
	if err != nil {
		return nil, err
	}

	var images []*Image
	if err := json.Unmarshal(data, &images); err != nil {
		return nil, err
	}

	return images, nil
}

func (client *DockerClient) RemoveImage(name string) error {
	_, err := client.doRequest("DELETE", fmt.Sprintf("/v1.10/images/%s", name), nil)
	return err
}
