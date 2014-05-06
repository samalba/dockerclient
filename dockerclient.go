package dockerclient

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sync/atomic"
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

// Status info the flows from things like 'pull' requests.
type jsonStatusInfo struct {
	Status   string `json:"status,omitempty"`
	Progress string `json:"progress,omitempty"`
	Error    string `json:"error,omitempty"`
	Stream   string `json:"stream,omitempty"`
}

// AuthConfiguration: auth options for Docker image/registry api.
type AuthConfiguration struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Email    string `json:"email,omitempty"`
}

// PullImageParams: params available for pulling an image from a registry.
type PullImageParams struct {
	FromImage    string    `qparam:"fromImage"`
	Registry     string    `qparam:"registry"`
	Tag          string    `qparam:"tag"`
	OutputStream io.Writer `qparam:"-"`
}

// Return a new dockerclient for use in subsequent calls to the remote Docker API.
type Callback func(*Event, ...interface{})

// Return a new dockerclient for use in subsequent calls to the remote Docker API.
func NewDockerClient(daemonUrl string) (*DockerClient, error) {
	u, err := url.Parse(daemonUrl)
	if err != nil {
		return nil, err
	}
	httpClient := newHTTPClient(u)
	return &DockerClient{u, httpClient, false, 0}, nil
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

func (client *DockerClient) doStream(method, path string, headers map[string]string,
	in io.Reader, out io.Writer) error {
	if (method == "POST" || method == "PUT") && in == nil {
		in = bytes.NewReader(nil)
	}
	req, err := http.NewRequest(method, client.URL.String()+path, in)
	if err != nil {
		return err
	}
	for key, val := range headers {
		req.Header.Set(key, val)
	}
	if out == nil {
		out = ioutil.Discard
	}
	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return NewDockerClientError(resp.StatusCode, string(body))
	}
	if resp.Header.Get("Content-Type") == "application/json" {
		dec := json.NewDecoder(resp.Body)
		for {
			var msg jsonStatusInfo
			if err := dec.Decode(&msg); err == io.EOF {
				break
			} else if err != nil {
				return err
			}
			if msg.Stream != "" {
				fmt.Fprint(out, msg.Stream)
			} else if msg.Progress != "" {
				fmt.Fprintf(out, "%s %s\r", msg.Status, msg.Progress)
			} else if msg.Error != "" {
				return errors.New(msg.Error)
			}
			fmt.Fprintln(out, msg.Status)
		}
	} else {
		if _, err := io.Copy(out, resp.Body); err != nil {
			return err
		}
	}
	return nil
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
func (client *DockerClient) StartMonitorEvents(cb Callback, args ...interface{}) {
	atomic.StoreInt32(&client.monitorEvents, 1)
	go client.getEvents(cb, args...)
}

func (client *DockerClient) getEvents(cb Callback, args ...interface{}) {
	uri := client.URL.String() + DockerBaseURL + "/events"
	resp, err := client.HTTPClient.Get(uri)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)
	for atomic.LoadInt32(&client.monitorEvents) > 0 {
		var event *Event
		if err := dec.Decode(&event); err != nil {
			log.Println(err)
			return
		}
		cb(event, args...)
	}
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

//-------------------------------------------------------
// Pull an image from a remote Docker index and log progress to output stream.
func (client *DockerClient) PullImage(params *PullImageParams, auth *AuthConfiguration) error {
	if params.Registry == "" {
		return errors.New("Registry cannot be empty")
	}

	var h = make(map[string]string)
	if auth != nil {
		var buf bytes.Buffer
		json.NewEncoder(&buf).Encode(auth)
		h["X-Registry-Auth"] = base64.URLEncoding.EncodeToString(buf.Bytes())
	}

	qparams, error := queryParams(&params)
	if error != nil {
		return error
	}

	return client.doStream("POST", "/images/create?"+qparams, h, nil, params.OutputStream)
}
