//spring boot style configuration managemer
package gconfig

import (
	"bufio"
	"flag"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	s "strings"
	"fmt"
)

type GConfig struct {
	Profile  string
	FileName string
	configs  map[string]interface{}
}

func (c *GConfig) GetString(key string) string {
	return ""
}

func (c *GConfig) GetInt(key string) int {
	return 0
}

func (c *GConfig) GetBool(key string) string {
	return ""
}

func (c *GConfig) PrettyPrint(key string) string {
	return ""
}

func configError(cause error, format string, args ...interface{}) (*GConfig, error) {
	return new(GConfig), errors.Wrap(cause, fmt.Sprintf(format, args...))
}

func Load() (*GConfig, error) {
	gc := new(GConfig)
	gc.configs = make(map[string]interface{})
	gc.Profile = loadProfile()
	gc.FileName = "application-" + gc.Profile + ".properties"

	p := loadPath()
	files, err := ioutil.ReadDir(p)
	if err != nil {
		return configError(err, "Error reading config directory at path:: %s", p)
	}

	fis := make([]os.FileInfo, 2)
	for _, file := range files {
		if file.Name() == "application.properties" || file.Name() == gc.FileName {
			fis = append(fis, file)
		}
	}

	for _, fi := range files {
		f, err := os.Open(fi.Name())
		if err != nil {
			return configError(err, "Error opening config file:: %s", fi)
		}

		sc := bufio.NewScanner(f)
		sc.Split(bufio.ScanLines)
		for sc.Scan() {
			l := sc.Text()
			kv := s.Split(l, "=")
			fmt.Printf("Key:%s\n", kv[0])
			fmt.Printf("Value:%s\n", kv[1])

		}

	}
	return gc, nil
}

//If no profile is specified then it uses the default profile and load the config data that's marked as default.
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
