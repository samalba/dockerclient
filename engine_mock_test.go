package dockerclient

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strconv"
	"time"

	"github.com/docker/docker/pkg/ioutils"
	"github.com/docker/docker/pkg/jsonlog"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/gorilla/mux"
)

var (
	testHTTPServer *httptest.Server
)

func init() {
	r := mux.NewRouter()
	baseURL := "/" + APIVersion
	r.HandleFunc(baseURL+"/info", handlerGetInfo).Methods("GET")
	r.HandleFunc(baseURL+"/containers/json", handlerGetContainers).Methods("GET")
	r.HandleFunc(baseURL+"/containers/{id}/logs", handleContainerLogs).Methods("GET")
	r.HandleFunc(baseURL+"/containers/{id}/changes", handleContainerChanges).Methods("GET")
	r.HandleFunc(baseURL+"/containers/{id}/kill", handleContainerKill).Methods("POST")
	r.HandleFunc(baseURL+"/containers/{id}/wait", handleWait).Methods("POST")
	r.HandleFunc(baseURL+"/images/create", handleImagePull).Methods("POST")
	r.HandleFunc(baseURL+"/images/json", handleGetImages).Methods("GET")
	r.HandleFunc(baseURL+"/events", handleEvents).Methods("GET")
	testHTTPServer = httptest.NewServer(handlerAccessLog(r))
}

func handlerAccessLog(handler http.Handler) http.Handler {
	logHandler := func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s \"%s %s\"", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	}
	return http.HandlerFunc(logHandler)
}

func handleContainerKill(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "{%q:%q", "Id", "421373210afd132")
}

func handleWait(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if vars["id"] == "valid-id" {
		fmt.Fprintf(w, `{"StatusCode":0}`)
	} else {
		http.Error(w, "failed", 500)
	}
}

func handleImagePull(w http.ResponseWriter, r *http.Request) {
	imageName := r.URL.Query()["fromImage"][0]
	responses := []map[string]interface{}{{
		"status": fmt.Sprintf("Pulling repository mydockerregistry/%s", imageName),
	}}
	switch imageName {
	case "busybox":
		responses = append(responses, map[string]interface{}{
			"status": "Status: Image is up to date for mydockerregistry/busybox",
		})
	case "haproxy":
		fmt.Fprintf(w, haproxyPullOutput)
		return
	default:
		errorMsg := fmt.Sprintf("Error: image %s not found", imageName)
		responses = append(responses, map[string]interface{}{
			"errorDetail": map[string]interface{}{
				"message": errorMsg,
			},
			"error": errorMsg,
		})
	}
	for _, response := range responses {
		json.NewEncoder(w).Encode(response)
	}
}

func handleContainerLogs(w http.ResponseWriter, r *http.Request) {
	var outStream, errStream io.Writer
	outStream = ioutils.NewWriteFlusher(w)

	// not sure how to test follow
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), 500)
	}
	stdout, stderr := getBoolValue(r.Form.Get("stdout")), getBoolValue(r.Form.Get("stderr"))
	if stderr {
		errStream = stdcopy.NewStdWriter(outStream, stdcopy.Stderr)
	}
	if stdout {
		outStream = stdcopy.NewStdWriter(outStream, stdcopy.Stdout)
	}
	var i int
	if tail, err := strconv.Atoi(r.Form.Get("tail")); err == nil && tail > 0 {
		i = 50 - tail
		if i < 0 {
			i = 0
		}
	}
	for ; i < 50; i++ {
		line := fmt.Sprintf("line %d", i)
		if getBoolValue(r.Form.Get("timestamps")) {
			l := &jsonlog.JSONLog{Log: line, Created: time.Now().UTC()}
			line = fmt.Sprintf("%s %s", l.Created.Format(jsonlog.RFC3339NanoFixed), line)
		}
		if i%2 == 0 && stderr {
			fmt.Fprintln(errStream, line)
		} else if i%2 == 1 && stdout {
			fmt.Fprintln(outStream, line)
		}
	}
}

func handleContainerChanges(w http.ResponseWriter, r *http.Request) {
	writeHeaders(w, 200, "changes")
	body := `[
          {
            "Path": "/dev",
            "Kind": 0
          },
          {
            "Path": "/dev/kmsg",
            "Kind": 1
          },
          {
            "Path": "/test",
            "Kind": 1
          }
        ]`
	w.Write([]byte(body))
}

func getBoolValue(boolString string) bool {
	switch boolString {
	case "1":
		return true
	case "True":
		return true
	case "true":
		return true
	default:
		return false
	}
}

func writeHeaders(w http.ResponseWriter, code int, jobName string) {
	h := w.Header()
	h.Add("Content-Type", "application/json")
	if jobName != "" {
		h.Add("Job-Name", jobName)
	}
	w.WriteHeader(code)
}

func handlerGetInfo(w http.ResponseWriter, r *http.Request) {
	writeHeaders(w, 200, "info")
	body := `{
	"Containers": 2,
	 "Debug": 1,
	 "Driver": "aufs",
	 "DriverStatus": [["Root Dir", "/mnt/sda1/var/lib/docker/aufs"],
	  ["Dirs", "0"]],
	 "ExecutionDriver": "native-0.2",
	 "IPv4Forwarding": 1,
	 "Images": 1,
	 "IndexServerAddress": "https://index.docker.io/v1/",
	 "InitPath": "/usr/local/bin/docker",
	 "InitSha1": "",
	 "KernelVersion": "3.16.4-tinycore64",
	 "MemoryLimit": 1,
	 "NEventsListener": 0,
	 "NFd": 10,
	 "NGoroutines": 11,
	 "OperatingSystem": "Boot2Docker 1.3.1 (TCL 5.4); master : a083df4 - Thu Jan 01 00:00:00 UTC 1970",
	 "SwapLimit": 1}`
	w.Write([]byte(body))
}

func handlerGetContainers(w http.ResponseWriter, r *http.Request) {
	writeHeaders(w, 200, "containers")
	body := `[
          {
            "Status": "Up 39 seconds",
            "Ports": [
              {
                "Type": "tcp",
                "PublicPort": 49163,
                "PrivatePort": 8080,
                "IP": "0.0.0.0"
              }
            ],
            "Names": [
              "/trusting_heisenberg"
            ],
            "Image": "foo:latest",
            "Id": "332375cfbc23edb921a21026314c3497674ba8bdcb2c85e0e65ebf2017f688ce",
            "Created": 1415720105,
            "Command": "/bin/go-run"
          }
        ]`
	if v, ok := r.URL.Query()["size"]; ok {
		if v[0] == "1" {
			body = `[
          {
            "Status": "Up 39 seconds",
            "Ports": [
              {
                "Type": "tcp",
                "PublicPort": 49163,
                "PrivatePort": 8080,
                "IP": "0.0.0.0"
              }
            ],
            "Names": [
              "/trusting_heisenberg"
            ],
            "Image": "foo:latest",
            "Id": "332375cfbc23edb921a21026314c3497674ba8bdcb2c85e0e65ebf2017f688ce",
            "Created": 1415720105,
            "SizeRootFs": 12345,
            "SizeRW": 123,
            "Command": "/bin/go-run"
          }
        ]`
		}
	}
	if v, ok := r.URL.Query()["filters"]; ok {
		if v[0] != "{'id':['332375cfbc23edb921a21026314c3497674ba8bdcb2c85e0e65ebf2017f688ce']}" {
			body = "[]"
		}
	}
	w.Write([]byte(body))
}

func handleGetImages(w http.ResponseWriter, r *http.Request) {
	writeHeaders(w, 200, "images")
	body := `[
			{
   				"Id":"d39c3fa09ced9473c6e4fffa9d26c63cb531d60552abdaf58dd6e703b6662ecd",
   				"ParentId":"8eb845d8d7f6c255db9550bc1f6031f57d7e0cc599945644dc6f676928b9fec8",
   				"RepoTags":[
      				"mysql:latest"
   				],
   				"RepoDigests":[

   				],
   				"Created":1453142946,
   				"Size":0,
   				"VirtualSize":360299416,
   				"Labels":null
				},
				{
					"Id":"0750f4b0fe91d69123565609507f4dd38e19a0469b56d080bd033d89675007ac",
   				"ParentId":"d77a278aa5b811a93644a73ada70eb941cbe1c7b269fa860723f8bb61e38961c",
   				"RepoTags":[
      			"\u003cnone\u003e:\u003cnone\u003e"
   				],
   				"RepoDigests":[
      			"\u003cnone\u003e@\u003cnone\u003e"
   				],
   				"Created":1452457247,
   				"Size":0,
   				"VirtualSize":655675014,
   				"Labels":{

   				}
				}
]`

	if v, ok := r.URL.Query()["filters"]; ok {
		if v[0] == "{'dangling':['true']}" {
			body = `[{
				"Id":"0750f4b0fe91d69123565609507f4dd38e19a0469b56d080bd033d89675007ac",
				"ParentId":"d77a278aa5b811a93644a73ada70eb941cbe1c7b269fa860723f8bb61e38961c",
				"RepoTags":[
					"\u003cnone\u003e:\u003cnone\u003e"
				],
				"RepoDigests":[
					"\u003cnone\u003e@\u003cnone\u003e"
				],
				"Created":1452457247,
				"Size":0,
				"VirtualSize":655675014,
				"Labels":{

				}
			}]`
		}
	}
	w.Write([]byte(body))
}

func handleEvents(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(eventsResp))
}
