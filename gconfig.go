// Package gconfig - Spring boot style configuration manager. It can load either properties or yaml formatted files.
// properties file should follow the naming convention:
//
// 1. application.properties: this holds all the default configuration values as key/value pair.
// 2. application-{profile}.properties. contains all the environment specific configuration values.
//    eg: for prod environment, application-prod.properties

package gconfig

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	s "strings"

	"github.com/pkg/errors"
	"path/filepath"
	"strconv"
)

const (
	PROP_EXTENSION         string = ".properties"
	DEFAULT_PROP_FILENAME  string = "application-default.properties"
	STANDARD_PROP_FILENAME string = "application.properties"
)

var Gcg *GConfig
var ConfigFileRequired = errors.New("At least one configuration file is required")

// configFile is a internal representation of individual configurations for default and env specific
// configuration values.
type configFile struct {
	fileInfo os.FileInfo
	configs  map[string]interface{}
}

func (cf configFile) Name() string {
	return cf.fileInfo.Name()
}

func (cf *configFile) addProperty(key, value string) {
	k := s.Trim(key, " ")
	v := s.Trim(value, " ")

	cf.configs[k] = v
}

func (cf configFile) isDefault() bool {
	if cf.Name() == DEFAULT_PROP_FILENAME || cf.Name() == STANDARD_PROP_FILENAME {
		return true
	}
	return false
}

// GConfig is the representation of all the configuration properties. It loads 2 types of data: default and environment
// specific. One out of 2 must be present otherwise, error is returned during the Load operation
type GConfig struct {
	Profile                      string
	profileConfig, defaultConfig configFile
}

func (c GConfig) GetString(key string) string {
	return c.getValue(key)
}

func (c GConfig) GetInt(key string) int {
	i, _ := strconv.Atoi(c.getValue(key))
	return i
}

func (c GConfig) GetFloat(key string) float64 {
	v, _ := strconv.ParseFloat(c.getValue(key), 32)
	return v
}

func (c GConfig) GetBool(key string) bool {
	b, _ := strconv.ParseBool(c.getValue(key))
	return b
}

// getValue returns a value for a given key as type interface which is converted to actual return type by
// individual Get* functions.
func (c GConfig) getValue(key string) string {
	v := c.profileConfig.configs[key]
	if v == nil {
		v = c.defaultConfig.configs[key]
	}

	return v.(string)
}

func (c *GConfig) addConfigFile(cf configFile) {
	if cf.isDefault() {
		c.defaultConfig = cf
	} else {
		c.profileConfig = cf
	}
}

func (c GConfig) isEmpty() bool {
	return len(c.profileConfig.configs) == 0 && len(c.defaultConfig.configs) == 0
}

func configError(cause error, format string, args ...interface{}) (*GConfig, error) {
	return new(GConfig), errors.Wrap(cause, fmt.Sprintf(format, args...))
}

func Load() (*GConfig, error) {
	gc := new(GConfig)
	gc.Profile = loadProfile()

	p := loadPath()
	files, err := ioutil.ReadDir(p)
	if err != nil {
		return configError(err, "Error reading config directory in path %s", p)
	}
	if len(files) == 0 {
		return configError(ConfigFileRequired, "Config file not found in path %s", p)
	}

	//read individual config file
	for _, f := range files {
		cfpath := filepath.Join(p, f.Name())
		if path.Ext(f.Name()) == PROP_EXTENSION {
			cf, err := readPropertyFile(f, cfpath)
			if err != nil {
				return configError(err, "Error opening config file %s", f)
			}
			gc.addConfigFile(cf)
		}
	}

	Gcg = gc

	//do a final check if loaded config has any values
	if gc.isEmpty() {
		log.Printf("Configuration loaded, but empty for profile: '%s'\n", Gcg.Profile)
	} else {
		log.Printf("Configuration loaded for profile %s\n", Gcg.Profile)
	}

	return gc, nil
}

// readPropertyFile opens the configuration file and creates configuration struct with all the key/value pair info.
// It ignores any line that begins with # and silently ignores line without correct key/value pair format.
func readPropertyFile(fi os.FileInfo, cfpath string) (configFile, error) {
	cf := configFile{fileInfo: fi, configs: make(map[string]interface{})}

	f, err := os.Open(cfpath)
	if err != nil {
		return configFile{}, err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	sc.Split(bufio.ScanLines)
	for sc.Scan() {
		l := sc.Text()
		kv := s.Split(l, "=")
		if len(kv) == 2 {
			cf.addProperty(kv[0], kv[1])
		}
	}

	return cf, nil
}

// If no profile is specified then it uses the default profile and load the config
// data that's marked as default.
// Profile can be set using 2 ways:
// 1. Environment variable 'GC_PROFILE' eg: export GC_PROFILE='dev'
// 2. Command line argument 'profile' eg: go run myserver.go -profile=dev
func loadProfile() string {
	//Load application profile from environment variable
	profile := os.Getenv("GC_PROFILE")
	if len(profile) == 0 {
		p := flag.String("profile", "local", "-profile=dev")
		profile = *p
	}

	return s.ToLower(profile)
}

//Check if location of config or properties file is set in the env variable
//if no path is specified it will use the current directory
func loadPath() string {
	path := os.Getenv("GC_PATH")
	if len(path) == 0 {
		p := flag.String("path", "./config", "-path=/Users/puran/myserver/config")
		path = *p
	}
	return path
}
