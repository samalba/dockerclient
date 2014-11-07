package dockerclient

import (
	"io"

	"github.com/stretchr/testify/mock"
)

type DockerClientMock struct {
	mock.Mock
}

func NewDockerClientMock() *DockerClientMock {
	return &DockerClientMock{}
}

func (client *DockerClientMock) Info() (*Info, error) {
	args := client.Mock.Called()
	return args.Get(0).(*Info), args.Error(1)
}

func (client *DockerClientMock) ListContainers(all bool) ([]Container, error) {
	args := client.Mock.Called(all)
	return args.Get(0).([]Container), args.Error(1)
}

func (client *DockerClientMock) InspectContainer(id string) (*ContainerInfo, error) {
	args := client.Mock.Called(id)
	return args.Get(0).(*ContainerInfo), args.Error(1)
}

func (client *DockerClientMock) CreateContainer(config *ContainerConfig, name string) (string, error) {
	args := client.Mock.Called(config, name)
	return args.String(0), args.Error(1)
}

func (client *DockerClientMock) ContainerLogs(id string, stdout bool, stderr bool) (io.ReadCloser, error) {
	args := client.Mock.Called(id, stdout, stderr)
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func (client *DockerClientMock) StartContainer(id string, config *HostConfig) error {
	args := client.Mock.Called(id, config)
	return args.Error(0)
}

func (client *DockerClientMock) StopContainer(id string, timeout int) error {
	args := client.Mock.Called(id, timeout)
	return args.Error(0)
}

func (client *DockerClientMock) RestartContainer(id string, timeout int) error {
	args := client.Mock.Called(id, timeout)
	return args.Error(0)
}

func (client *DockerClientMock) KillContainer(id string) error {
	args := client.Mock.Called(id)
	return args.Error(0)
}

func (client *DockerClientMock) StartMonitorEvents(cb Callback, args ...interface{}) {
	client.Mock.Called(cb, args)
}

func (client *DockerClientMock) StopAllMonitorEvents() {
	client.Mock.Called()
}

func (client *DockerClientMock) Version() (*Version, error) {
	args := client.Mock.Called()
	return args.Get(0).(*Version), args.Error(1)
}

func (client *DockerClientMock) PullImage(name, tag string) error {
	args := client.Mock.Called(name, tag)
	return args.Error(0)
}

func (client *DockerClientMock) RemoveContainer(id string, force bool) error {
	args := client.Mock.Called(id, force)
	return args.Error(0)
}

func (client *DockerClientMock) ListImages() ([]*Image, error) {
	args := client.Mock.Called()
	return args.Get(0).([]*Image), args.Error(1)
}

func (client *DockerClientMock) RemoveImage(name string) error {
	args := client.Mock.Called(name)
	return args.Error(0)
}
