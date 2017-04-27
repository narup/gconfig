package gconfig

import (
	"fmt"
	"os"
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
	fmt.Printf("app.name: %s\n", n)

	expectedURL := "https://github.com/narup/gconfig-dev"
	u := gcg.GetString("app.url")
	if u != expectedURL {
		t.Errorf("Key app.url didn't match expected value %s\n", expectedURL)
	}

	fmt.Printf("app.url: %s\n", expectedURL)

	envVariable := gcg.GetString("myEnv.variable")
	fmt.Printf(envVariable)
}
