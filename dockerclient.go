package dockerclient

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
)

type DockerClient struct {
	URL        *url.URL
	HTTPClient *http.Client
}

func NewDockerClient(daemonUrl string) (*DockerClient, error) {
	u, err := url.Parse(daemonUrl)
	if err != nil {
		return nil, err
	}
	httpClient := newHTTPClient(u)
	return &DockerClient{u, httpClient}, nil
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

func (client *DockerClient) doRequest(method string, path string) ([]byte, error) {
	req, err := http.NewRequest(method, client.URL.String() + path, nil)
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
	return data, nil
}

func (client *DockerClient) ListContainers(all bool) (*[]Container, error) {
	argAll := 0
	if all == true {
		argAll = 1
	}
	args := fmt.Sprintf("?all=%d", argAll)
	data, err := client.doRequest("GET", "/v1.8/containers/json" + args)
	if err != nil {
		return nil, err
	}
	ret := &[]Container{}
	json.Unmarshal(data, ret)
	return ret, nil
}

func (client *DockerClient) InspectContainer(id string) (*ContainerInfo, error) {
	uri := fmt.Sprintf("/v1.8/containers/%s/json", id)
	data, err := client.doRequest("GET", uri)
	if err != nil {
		return nil, err
	}
	info := &ContainerInfo{}
	json.Unmarshal(data, info)
	return info, nil
}
