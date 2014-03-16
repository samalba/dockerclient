// dockerclient
// For the full copyright and license information, please view the LICENSE file.

// Package dockerclient provides Docker client library
//
// References:
//  Attach Protocol: http://docs.docker.io/en/latest/reference/api/docker_remote_api_v1.10/#attach-to-a-container
//
package dockerclient

import (
	//"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
	"time"
)

const (
	APIVERSION = "v1.10"
)

// DockerClient implements a docker client.
type DockerClient struct {
	URL           *url.URL
	HTTPClient    *http.Client
	monitorEvents int32
}

// NewDockerClient returns a new docker client with the given URL.
func NewDockerClient(daemonUrl string) (*DockerClient, error) {
	u, err := url.Parse(daemonUrl)
	if err != nil {
		return nil, err
	}
	httpClient := newHTTPClient(u)
	return &DockerClient{u, httpClient, 0}, nil
}

// newHTTPClient returns a new HTTP client with the given URL.
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

// doRequest makes request to the docker with the given method, path and body.
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

// ListContainers returns the list of the containers.
func (client *DockerClient) ListContainers(all bool) ([]Container, error) {
	argAll := 0
	if all == true {
		argAll = 1
	}
	args := fmt.Sprintf("?all=%d", argAll)
	uri := fmt.Sprintf("/%s/containers/json%s", APIVERSION, args)
	data, err := client.doRequest("GET", uri, nil)
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

// InspectContainer inspects and returns the container information with the given container id.
func (client *DockerClient) InspectContainer(id string) (*ContainerInfo, error) {
	uri := fmt.Sprintf("/%s/containers/%s/json", APIVERSION, id)
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

// CreateContainer creates a container and returns the id.
func (client *DockerClient) CreateContainer(config *ContainerConfig) (string, error) {
	data, err := json.Marshal(config)
	if err != nil {
		return "", err
	}
	uri := fmt.Sprintf("/%s/containers/create", APIVERSION)
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

// StartContainer starts the container with the given id.
func (client *DockerClient) StartContainer(id string) error {
	uri := fmt.Sprintf("/%s/containers/%s/start", APIVERSION, id)
	_, err := client.doRequest("POST", uri, nil)
	if err != nil {
		return err
	}
	return nil
}

// StopContainer stops the container with the given id.
func (client *DockerClient) StopContainer(id string, timeout int) error {
	uri := fmt.Sprintf("/%s/containers/%s/stop?t=%d", APIVERSION, id, timeout)
	_, err := client.doRequest("POST", uri, nil)
	if err != nil {
		return err
	}
	return nil
}

// RestartContainer restarts the container with the given id.
func (client *DockerClient) RestartContainer(id string, timeout int) error {
	uri := fmt.Sprintf("/%s/containers/%s/restart?t=%d", APIVERSION, id, timeout)
	_, err := client.doRequest("POST", uri, nil)
	if err != nil {
		return err
	}
	return nil
}

// KillContainer kills the container with the given id.
func (client *DockerClient) KillContainer(id string) error {
	uri := fmt.Sprintf("/%s/containers/%s/kill", APIVERSION, id)
	_, err := client.doRequest("POST", uri, nil)
	if err != nil {
		return err
	}
	return nil
}

// AttachContainer attaches to the container with the given id and an Attach variable.
func (client *DockerClient) AttachContainer(id string, att Attach) error {

	// Attach protocol stream details:
	// When using the TTY setting is enabled in POST /containers/create,
	// the stream is the raw data from the process PTY and client’s stdin.
	// When the TTY is disabled, then the stream is multiplexed to separate stdout and stderr.
	//
	// HEADER
	// 	header := [8]byte{STREAM_TYPE, 0, 0, 0, SIZE1, SIZE2, SIZE3, SIZE4}
	// PAYLOAD
	// 	The payload is the raw stream.

	// Check container
	info, err := client.InspectContainer(id)
	if err != nil {
		return errors.New("failed to inspect container: " + err.Error())
	}
	att.Tty = info.Config.Tty
	//openStdin := info.Config.OpenStdin
	//exitcode := info.State.ExitCode

	// Init connection
	connNet := client.URL.Scheme
	if connNet == "http" {
		connNet = "tcp"
	}
	conn, err := net.Dial(connNet, client.URL.Host)
	if err != nil {
		return err
	}
	cliConn := httputil.NewClientConn(conn, nil)
	defer cliConn.Close()

	// Request
	// A request without these query parameters doesn't make sense.
	// So they are not optional at the moment.
	qp := "stream=1&stdout=1&stderr=1&stdin=1"
	uri := fmt.Sprintf("/"+APIVERSION+"/containers/%s/attach?%s", id, qp)
	req, err := http.NewRequest("POST", client.URL.String()+uri, nil)
	if err != nil {
		return err
	}
	cliConn.Do(req)

	// Hijack the connection
	hjConn, hjBuf := cliConn.Hijack()
	defer hjConn.Close()

	// Attach
	if att.Tty == true {
		// Tty implementation

		if att.Stdin != nil {
			if att.Stdout != nil {
				go func() {
					// FIX: There is extra output due stdin
					io.Copy(att.Stdout, hjBuf)
				}()
			}

			io.Copy(hjConn, att.Stdin)
		} else if att.StdinPipe != nil {
			io.Copy(hjConn, att.StdinPipe)

			if att.Stdout != nil {
				io.Copy(att.Stdout, hjBuf)
			}
		} else if att.Stdout != nil {
			io.Copy(att.Stdout, hjBuf)
		}
	} else {
		// Multiplexed implementation

		return errors.New("Multiplexed implementation is still under development...")

		// Header
		bufHdr := make([]byte, 8)
		n, err := hjBuf.Read(bufHdr)
		if err != nil && err != io.EOF {
			return err
		} else if n != 8 {
			return errors.New("invalid header: " + fmt.Sprintf("%d / %v", n, bufHdr))
		}

		// Frame 1
		att.Tty = false
		frame1Type := int(bufHdr[0]) // Stream type 0: stdin, 1: stdout, 2: stderr
		frame1Size := int(binary.BigEndian.Uint32(bufHdr[4:]))

		var _ = frame1Type
		var _ = frame1Size

		// TODO: Implement a loop for stream frames
	}

	return nil
}

// StartMonitorEvents provides event monitoring with the given callback function and the arguments.
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
			uri := client.URL.String() + "/" + APIVERSION + "/events"
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

// StopAllMonitorEvents stops the all event monitoring related stuff.
func (client *DockerClient) StopAllMonitorEvents() {
	atomic.StoreInt32(&client.monitorEvents, 0)
}

// Version returns the version information of the Docker.
func (client *DockerClient) Version() (*Version, error) {
	uri := fmt.Sprintf("/%s/version", APIVERSION)
	data, err := client.doRequest("GET", uri, nil)
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
