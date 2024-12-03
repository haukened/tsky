package config

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	yaml2 "gopkg.in/yaml.v2"
)

var k = koanf.New(".")

type Config struct {
	Did         string `koanf:"did" yaml:"did"`
	Identifier  string `koanf:"identifier" yaml:"identifier"`
	RefreshJwt  string `koanf:"refresh_jwt" yaml:"refresh_jwt"`
	AccessJwt   string `koanf:"-" yaml:"-"` // do not marshal this field
	AppPassword string `koanf:"-" yaml:"-"` // do not marshal this field
	Path        string `koanf:"-" yaml:"-"` // do not marshal this field
	Server      string `koanf:"server,omitempty" yaml:"server,omitempty"`
	Debug       bool   `koanf:"debug,omitempty" yaml:"debug,omitempty"`
}

func New(path string) (*Config, error) {
	expanded, err := expandHomeDir(path)
	if err != nil {
		return nil, err
	}
	return &Config{
		Path: expanded,
	}, nil
}

func (c *Config) checkFilePermissions() error {
	fileInfo, err := os.Stat(c.Path)
	if err != nil {
		return err
	}

	// Check if the file is readable by others
	if fileInfo.Mode().Perm()&0044 != 0 {
		return fmt.Errorf("your config file located at %s permissions are to permissive, and readable to other users", c.Path)
	}

	return nil
}

// Checks if the file exists at the specified path.
func (c *Config) Exists() bool {
	// check if the file exists
	if _, err := os.Stat(c.Path); err != nil {
		return false
	}
	return true
}

// Load loads the configuration from the specified file path and environment variables.
// if the file does not exist, it will return an empty Config struct.
//
// Parameters:
//   - path: The file path to the YAML configuration file.
//
// Returns:
//   - *Config: A pointer to the loaded Config struct.
//   - error: An error if any occurred during the loading process.
func (c *Config) Load() error {
	if c.Exists() {
		// check the file permissions
		if err := c.checkFilePermissions(); err != nil {
			return err
		}

		// load the yaml config file
		if err := k.Load(file.Provider(c.Path), yaml.Parser()); err != nil {
			return err
		}
	}

	// load the environment variables that begin with TSKY_
	// and replace the underscores with dots so that they are the same as the yaml keys
	// strip the TSKY_ prefix so that only the key remains
	if err := k.Load(env.Provider("TSKY_", ".", func(s string) string {
		return strings.Replace(strings.ToLower(
			strings.TrimPrefix(s, "TSKY_")), "_", ".", -1)
	}), nil); err != nil {
		return err
	}

	// unmarshal the config into the Config struct
	if err := k.Unmarshal("", &c); err != nil {
		return err
	}

	// set the default server if it is not set
	if c.Server == "" {
		c.Server = "bsky.social"
	}

	return nil
}

func (c *Config) Save() error {
	// marshal the data into yaml
	data, err := yaml2.Marshal(&c)
	if err != nil {
		return err
	}

	// ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(c.Path), 0700); err != nil {
		return err
	}

	// write the data to the file
	if err := os.WriteFile(c.Path, data, 0600); err != nil {
		return err
	}

	return nil
}

func expandHomeDir(path string) (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return usr.HomeDir + path[1:], nil
}
