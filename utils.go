package dockerclient

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

// Used for parsing the JSON values from HostConfig and other messages.
const (
	queryTag    = "qparam"
	queryIgnore = "-"
)

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

// Adds a Value to the query string parameter set, if valid.
func __addToValues(param string, v *reflect.Value, values *url.Values) error {
	switch v.Kind() {
	case reflect.String:
		if v.String() != "" {
			values.Add(param, v.String())
		}
	case reflect.Bool:
		if v.Bool() {
			values.Add(param, "1")
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if v.Int() > 0 {
			values.Add(param, strconv.FormatInt(v.Int(), 10))
		}
	case reflect.Float32, reflect.Float64:
		if v.Float() > 0 {
			values.Add(param, strconv.FormatFloat(v.Float(), 'f', -1, 64))
		}
	case reflect.Ptr:
		if !v.IsNil() {
			j, err := json.Marshal(v.Interface())
			if err != nil {
				return errors.New("Error converting " + param + " pointer to JSON")
			}
			values.Add(param, string(j))
		}
	}
	return nil
}

// Adds the Values to the query parameter string 'parser' (effectively the url.Values).
func addToValues(fieldIdx int, value *reflect.Value, values *url.Values) error {
	field := value.Type().Field(fieldIdx)
	if field.PkgPath != "" {
		return nil
	}
	param := field.Tag.Get(queryTag)
	switch param {
	case queryIgnore:
		return nil
	case "":
		param = strings.ToLower(field.Name)
	}
	v := value.Field(fieldIdx)
	return __addToValues(param, &v, values)
}

// generates query parameters from a data structure (presumed json-ish).
func queryParams(params interface{}) (string, error) {
	// handle nil case -- no query params
	if params == nil {
		return "", nil
	}

	// don't deal with pointers -- just drill down until a real value exists
	value := reflect.ValueOf(params)
	for value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	// the query string 'builder'
	values := url.Values{}

	// don't deal with non-structs -- there'd be no tags
	if value.Kind() != reflect.Struct {
		return "", errors.New("Unsupported: cannot convert non-struct type.")
	} else {
		for index := 0; index < value.NumField(); index++ {
			if err := addToValues(index, &value, &values); err != nil {
				return "", err
			}
		}
	}

	return values.Encode(), nil
}
