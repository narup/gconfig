// Package gconfig - Spring boot style configuration manager. It can load properties files.
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
	s "strings"

	"path/filepath"
	"strconv"

	"github.com/pkg/errors"
)

const (
	// PropertiesExtension defines extension for properties file
	PropertiesExtension string = ".properties"
	// StandardPropFileName standard properties file if default is not defined
	StandardPropFileName string = "application.properties"
)

//Gcg is a global variable that represents configuration
var Gcg *GConfig

// Command line profile and path flags that can be passed when running the application
var cpath *string
var profile *string

// ErrConfigFileRequired represents file required error
var ErrConfigFileRequired = errors.New("At least one configuration file is required")

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
	if cf.Name() == StandardPropFileName {
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

// GetString returns string value for the given key
func (c GConfig) GetString(key string) string {
	return c.getStringValue(key)
}

// GetInt returns int value for the given key
func (c GConfig) GetInt(key string) int {
	i, _ := strconv.Atoi(c.getStringValue(key))
	return i
}

// GetFloat returns float value for the given key
func (c GConfig) GetFloat(key string) float64 {
	v, _ := strconv.ParseFloat(c.getStringValue(key), 32)
	return v
}

// GetBool returns bool value for the given key
func (c GConfig) GetBool(key string) bool {
	b, _ := strconv.ParseBool(c.getStringValue(key))
	return b
}

// Exists checks if key exists
func (c GConfig) Exists(key string) bool {
	v := c.getValue(key)
	if v != nil {
		return true
	}
	return false
}

// getStringValue returns a value for a given key as type interface which is converted
// to actual return type by individual Get* functions.
func (c GConfig) getStringValue(key string) string {
	v := c.getValue(key)
	strV := v.(string)
	if s.HasPrefix(strV, "${") && s.HasSuffix(strV, "}") {
		return os.ExpandEnv(strV)
	}

	return strV
}

// getValue gets the raw value for a given key
func (c GConfig) getValue(key string) interface{} {
	v := c.defaultConfig.configs[key]
	if c.profileConfig.fileInfo != nil && s.Contains(c.profileConfig.fileInfo.Name(), c.Profile) {
		v = c.profileConfig.configs[key]
	}
	if v == nil {
		v = c.defaultConfig.configs[key]
	}

	return v
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

func init() {
	cpath = flag.String("path", "", "-path=/Users/puran/myserver/config")
	profile = flag.String("profile", "", "-profile=dev")
}

// Load reads all the properties and creates GConfig representation. It loads
// config data based on passed in flags or environment variables. If none is
// defined it uses default values.
func Load() (*GConfig, error) {

	flag.Parse()

	gc := new(GConfig)
	gc.Profile = loadProfile()

	p, err := loadPath()
	if err != nil {
		return configError(err, "Error reading config directory path %s", p)
	}

	files, err := ioutil.ReadDir(p)
	if err != nil {
		return configError(err, "Error reading config directory in path %s", cpath)
	}
	if len(files) == 0 {
		return configError(ErrConfigFileRequired, "Config file not found in path %s", cpath)
	}

	//read individual config file
	for _, f := range files {
		cfpath := filepath.Join(p, f.Name())
		pf := fmt.Sprintf("application-%s.properties", gc.Profile)
		if f.Name() == StandardPropFileName || pf == f.Name() {
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
	p := ""
	if len(*profile) == 0 {
		//Load application profile from environment variable
		p = os.Getenv("GC_PROFILE")
	} else {
		p = *profile
	}
	return s.ToLower(p)
}

//Check if location of config or properties file is set in the env variable
//if no path is specified it will use the current directory
func loadPath() (string, error) {
	path := ""
	if len(*cpath) == 0 {
		path = os.Getenv("GC_PATH")
	} else {
		path = *cpath
	}

	//if empty, load default config path
	if len(path) == 0 {
		gp, err := getGoPath()
		if err != nil {
			return "", err
		}
		path = gp + "/config"
	}

	log.Printf("Loading configuration file from path %s\n", path)

	return path, nil
}

// getGoPath returns a single GOPATH. If there are multiple defined in the users
// $GOPATH env variable, then getGoPath validates that the working directory is
// part of one of the GOPATHs, and uses the first one it finds that does.
func getGoPath() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		return "", fmt.Errorf("GOPATH not set, it's needed to run go program")
	}

	// Split out multiple GOPATHs if necessary
	if s.Contains(gopath, ":") {
		paths := s.Split(gopath, ":")
		for _, path := range paths {
			if s.Contains(wd, path) {
				gopath = path
				break
			}
		}
	}

	if !s.Contains(wd, gopath) {
		return "", fmt.Errorf("gconfig can only be executed within a directory in"+
			" the GOPATH, wd: %s, gopath: %s", wd, gopath)
	}
	return gopath, nil
}
