package dockerclient

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
)

// AuthConfig hold parameters for authenticating with the docker registry
type AuthConfig struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Email    string `json:"email,omitempty"`
}

// encode the auth configuration struct into base64 for the X-Registry-Auth header
func (c *AuthConfig) encode() string {
	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(c)
	return base64.URLEncoding.EncodeToString(buf.Bytes())
}

// ConfigFile holds parameters for authenticating during a BuildImage request
type ConfigFile struct {
	Configs  map[string]AuthConfig `json:"configs,omitempty"`
	rootPath string
}

// encode the configuration struct into base64 for the X-Registry-Config header
func (c *ConfigFile) encode() string {
	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(c)
	return base64.URLEncoding.EncodeToString(buf.Bytes())
}
