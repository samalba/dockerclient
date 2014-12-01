package dockerclient

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"sync/atomic"
	"time"
)

const (
	APIVersion = "v1.15"
)

var (
	ErrNotFound = errors.New("Not found")

	defaultTimeout = 30 * time.Second
)

type DockerClient struct {
	URL           *url.URL
	HTTPClient    *http.Client
	monitorEvents int32
}

type Error struct {
	StatusCode int
	Status     string
	msg        string
}

func (e Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Status, e.msg)
}

func NewDockerClient(daemonUrl string, tlsConfig *tls.Config) (*DockerClient, error) {
	return NewDockerClientTimeout(daemonUrl, tlsConfig, time.Duration(defaultTimeout))
}

func NewDockerClientTimeout(daemonUrl string, tlsConfig *tls.Config, timeout time.Duration) (*DockerClient, error) {
	u, err := url.Parse(daemonUrl)
	if err != nil {
		return nil, err
	}
	if u.Scheme == "tcp" {
		if tlsConfig == nil {
			u.Scheme = "http"
		} else {
			u.Scheme = "https"
		}
	}
	httpClient := newHTTPClient(u, tlsConfig, timeout)
	return &DockerClient{u, httpClient, 0}, nil
}

func (client *DockerClient) doRequest(method string, path string, body []byte, headers map[string]string) ([]byte, error) {
	b := bytes.NewBuffer(body)
	req, err := http.NewRequest(method, client.URL.String()+path, b)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	if headers != nil {
		for header, value := range headers {
			req.Header.Add(header, value)
		}
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
	if resp.StatusCode == 404 {
		return nil, ErrNotFound
	}
	if resp.StatusCode >= 400 {
		return nil, Error{StatusCode: resp.StatusCode, Status: resp.Status, msg: string(data)}
	}
	return data, nil
}

func (client *DockerClient) Info() (*Info, error) {
	uri := fmt.Sprintf("/%s/info", APIVersion)
	data, err := client.doRequest("GET", uri, nil, nil)
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

func (client *DockerClient) ListContainers(all bool, size bool, filters string) ([]Container, error) {
	argAll := 0
	if all == true {
		argAll = 1
	}
	showSize := 0
	if size == true {
		showSize = 1
	}
	uri := fmt.Sprintf("/%s/containers/json?all=%d&size=%d", APIVersion, argAll, showSize)

	if filters != "" {
		uri += "&filters=" + filters
	}

	data, err := client.doRequest("GET", uri, nil, nil)
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
	uri := fmt.Sprintf("/%s/containers/%s/json", APIVersion, id)
	data, err := client.doRequest("GET", uri, nil, nil)
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
	uri := fmt.Sprintf("/%s/containers/create", APIVersion)
	if name != "" {
		v := url.Values{}
		v.Set("name", name)
		uri = fmt.Sprintf("%s?%s", uri, v.Encode())
	}
	data, err = client.doRequest("POST", uri, data, nil)
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

func (client *DockerClient) ContainerLogs(id string, options *LogOptions) (io.ReadCloser, error) {
	v := url.Values{}
	v.Add("follow", strconv.FormatBool(options.Follow))
	v.Add("stdout", strconv.FormatBool(options.Stdout))
	v.Add("stderr", strconv.FormatBool(options.Stderr))
	v.Add("timestamps", strconv.FormatBool(options.Timestamps))
	if options.Tail > 0 {
		v.Add("tail", strconv.FormatInt(options.Tail, 10))
	}

	uri := fmt.Sprintf("/%s/containers/%s/logs?%s", APIVersion, id, v.Encode())
	req, err := http.NewRequest("GET", client.URL.String()+uri, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func (client *DockerClient) StartContainer(id string, config *HostConfig) error {
	data, err := json.Marshal(config)
	if err != nil {
		return err
	}
	uri := fmt.Sprintf("/%s/containers/%s/start", APIVersion, id)
	_, err = client.doRequest("POST", uri, data, nil)
	if err != nil {
		return err
	}
	return nil
}

func (client *DockerClient) StopContainer(id string, timeout int) error {
	uri := fmt.Sprintf("/%s/containers/%s/stop?t=%d", APIVersion, id, timeout)
	_, err := client.doRequest("POST", uri, nil, nil)
	if err != nil {
		return err
	}
	return nil
}

func (client *DockerClient) RestartContainer(id string, timeout int) error {
	uri := fmt.Sprintf("/%s/containers/%s/restart?t=%d", APIVersion, id, timeout)
	_, err := client.doRequest("POST", uri, nil, nil)
	if err != nil {
		return err
	}
	return nil
}

func (client *DockerClient) KillContainer(id, signal string) error {
	uri := fmt.Sprintf("/%s/containers/%s/kill?signal=%s", APIVersion, id, signal)
	_, err := client.doRequest("POST", uri, nil, nil)
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
	uri := fmt.Sprintf("%s/%s/events", client.URL.String(), APIVersion)
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
	uri := fmt.Sprintf("/%s/version", APIVersion)
	data, err := client.doRequest("GET", uri, nil, nil)
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

func (client *DockerClient) PullImage(name string, auth *AuthConfig) error {
	v := url.Values{}
	v.Set("fromImage", name)
	uri := fmt.Sprintf("/%s/images/create?%s", APIVersion, v.Encode())

	headers := make(map[string]string)
	if auth != nil {
		headers["X-Registry-Auth"] = auth.encode()
	}
	_, err := client.doRequest("POST", uri, nil, headers)
	return err
}

func (client *DockerClient) RemoveContainer(id string, force bool) error {
	argForce := 0
	if force == true {
		argForce = 1
	}
	args := fmt.Sprintf("force=%d", argForce)
	uri := fmt.Sprintf("/%s/containers/%s?%s", APIVersion, id, args)
	_, err := client.doRequest("DELETE", uri, nil, nil)
	return err
}

func (client *DockerClient) ListImages() ([]*Image, error) {
	uri := fmt.Sprintf("/%s/images/json", APIVersion)
	data, err := client.doRequest("GET", uri, nil, nil)
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
	uri := fmt.Sprintf("/%s/images/%s", APIVersion, name)
	_, err := client.doRequest("DELETE", uri, nil, nil)
	return err
}

func (client *DockerClient) PauseContainer(id string) error {
	uri := fmt.Sprintf("/%s/containers/%s/pause", APIVersion, id)
	_, err := client.doRequest("POST", uri, nil, nil)
	if err != nil {
		return err
	}
	return nil
}
func (client *DockerClient) UnpauseContainer(id string) error {
	uri := fmt.Sprintf("/%s/containers/%s/unpause", APIVersion, id)
	_, err := client.doRequest("POST", uri, nil, nil)
	if err != nil {
		return err
	}
	return nil
}
