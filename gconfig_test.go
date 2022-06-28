package gconfig

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func setup(profile string, t *testing.T) *GConfig {
	wd, err := os.Getwd()
	if err != nil {
		t.Errorf("Test failed:%s", err)
	}

	p := wd + "/config"
	if profile != "" {
		os.Args = []string{"cmd", "-path=" + p, "-profile=dev"}
	} else {
		os.Args = []string{"cmd", "-path=" + p}
	}
	gcg, loadErr := Load()
	if loadErr != nil {
		t.Fatal(loadErr)
	}

	return gcg
}

func TestLoad(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	gcg := setup("", t)

	expectedName := "gconfig test"
	n := gcg.GetString("app.name")
	println(n)
	if n != expectedName {
		t.Errorf("Key app.name didn't match expected value %s\n", expectedName)
	}
	fmt.Printf("app.name: %s\n", n)

	expectedURL := "https://github.com/narup/gconfig"
	u := gcg.GetString("app.url")
	if u != expectedURL {
		t.Errorf("Key app.url didn't match expected value %s\n", expectedURL)
	}

	fmt.Printf("app.url: %s\n", expectedURL)
}

func TestProfileLoad(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	gcg := setup("dev", t)

	expectedName := "gconfig dev profile"
	n := gcg.GetString("app.name")
	if n != expectedName {
		t.Errorf("Key app.name didn't match expected value %s\n", expectedName)
	}

	expectedURL := "https://github.com/narup/gconfig-dev"
	u := gcg.GetString("app.url")
	if u != expectedURL {
		t.Errorf("Key app.url didn't match expected value %s\n", expectedURL)
	}

	envVariable := gcg.GetString("myEnv.variable")
	if strings.Trim(envVariable, " ") != "" {
		t.Errorf("Key value should be absent but it is %s", envVariable)
	}
}

func TestGetStringOrDefault(t *testing.T) {
	gcg := setup("dev", t)
	envVariable := gcg.GetStringOrDefault("myEnv.variable.withDefault")
	if envVariable != "default" {
		t.Errorf("Key app.name didn't match expected value %s\n", "default")
	}
}

func TestGetStringOrDefaultInCommaSeparator(t *testing.T) {
	gcg := setup("dev", t)
	envVariable := gcg.GetStringOrDefaultInCommaSeparator("myEnv.variable.listwithDefault")
	expectedString := "http://localhost:9292, http://localhost:3000, http://localhost:8080, https://int.dev.phil.us/, https://my.dev.phil.us, https://ops.dev.phil.us, https://web.dev.phil.us, https://md.dev.phil.us, https://us.dev.phil.us,https://workflow.dev.phil.us, https://data.dev.phil.us, https://gifted-goldberg-e5cca9.netlify.app"
	if strings.Trim(envVariable, " ") != strings.Trim(expectedString, " ") {
		t.Errorf("Key app.name didn't match expected value %s\n", expectedString)
	}
}

func TestGetStringOrDefaultInCommaSeparatorWithEnvValue(t *testing.T) {
	os.Setenv("DATA_DASHBOARD_ENDPOINT", "http://dataDashTest")
	gcg := setup("dev", t)

	envVariable := gcg.GetStringOrDefaultInCommaSeparator("myEnv.variable.listwithDefault")
	expectedString := "http://localhost:9292, http://localhost:3000, http://localhost:8080, https://int.dev.phil.us/, https://my.dev.phil.us, https://ops.dev.phil.us, https://web.dev.phil.us, https://md.dev.phil.us, https://us.dev.phil.us,https://workflow.dev.phil.us, http://dataDashTest, https://gifted-goldberg-e5cca9.netlify.app"
	if strings.Trim(envVariable, " ") != strings.Trim(expectedString, " ") {
		t.Errorf("Key app.name didn't match expected value %s\n", expectedString)
	}
}

func TestGetStringOrDefaultInCommaSeparatorWithEnvValueForAllTypes(t *testing.T) {
	expectedString := "CAPIAPI"
	os.Setenv("CAPI_API_KEY", expectedString)
	os.Setenv("CAPI_API_KEY_REG", "CAPI_API_KEY_REG")

	gcg := setup("dev", t)

	envVariable := gcg.GetStringOrDefaultInCommaSeparator("myEnv.variable.withDefault")
	if strings.Trim(envVariable, " ") != strings.Trim(expectedString, " ") {
		t.Errorf("Key app.name didn't match expected value %s\n", expectedString)
	}

	envVariable = gcg.GetStringOrDefaultInCommaSeparator("myEnv.variable")
	if envVariable != "CAPI_API_KEY_REG" {
		t.Errorf("Key app.name didn't match expected value %s\n", expectedString)
	}

	envVariable = gcg.GetStringOrDefaultInCommaSeparator("app.url")
	if envVariable != "https://github.com/narup/gconfig-dev" {
		t.Errorf("Key app.name didn't match expected value %s\n", expectedString)
	}
}
