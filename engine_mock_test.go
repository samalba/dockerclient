package dockerclient

import (
	"log"
	"net/http"
	"net/http/httptest"

	"github.com/gorilla/mux"
)

var (
	testHTTPServer *httptest.Server
)

func init() {
	r := mux.NewRouter()
	baseURL := "/" + APIVersion
	r.HandleFunc(baseURL+"/info", handlerGetInfo).Methods("GET")
	testHTTPServer = httptest.NewServer(handlerAccessLog(r))
}

func handlerAccessLog(handler http.Handler) http.Handler {
	logHandler := func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s \"%s %s\"", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	}
	return http.HandlerFunc(logHandler)
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
