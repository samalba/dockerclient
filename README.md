Docker client library in Go
===========================

This library supports few API calls but it will get extended over time.

Example:

```go
package main

import (
	"github.com/samalba/dockerclient"
	"log"
	"time"
)

// Callback used to listen to Docker's events
func eventCallback(event *dockerclient.Event) {
	log.Printf("Received event: %#v\n", *event)
}

func main() {
	// Init the client
	docker, _ := dockerclient.NewDockerClient("unix:///var/run/docker.sock")

	// Get only running containers
	containers, err := docker.ListContainers(false)
	if err != nil {
		log.Fatal(err)
	}
	for _, c := range *containers {
		log.Println(c.Id, c.Names)
	}

	// Inspect the first container returned
	id := (*containers)[0].Id
	info, _ := docker.InspectContainer(id)
	log.Println(info)

	// Create a container
	containerConfig := &dockerclient.ContainerConfig{
		Image: "ubuntu", Cmd: []string{"bash"}}
	containerId, err := docker.CreateContainer(containerConfig)
	if err != nil {
		log.Fatal(err)
	}

	// Start the container
	err = docker.StartContainer(containerId)
	if err != nil {
		log.Fatal(err)
	}

	// Stop the container (with 5 seconds timeout)
	docker.StartContainer(containerId, 5)

	// Listen to events
	docker.StartMonitorEvents(eventCallback)
	time.Sleep(3600 * time.Second)
}
```
