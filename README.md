# gconfig 
## Spring boot style configuration management for Go

Note: Only supports *.properties file, working on YAML support

### Code example
```go
    // Profile can be set using 2 ways:
    // 1. Environment variable 'GC_PROFILE' eg: export GC_PROFILE='dev'
    // 2. Command line argument 'profile' eg: go run myserver.go -profile=dev

    //Path
    // 1. Environment variable 'GC_PATH' eg: export GC_PATH='./config' config directory in $GOPATH folder
    // 2. Command line argument 'path' eg: -path=/Users/puran/myserver/config

    import "github.com/narup/gconfig"

    //load configuration
	if _, err := gconfig.Load(); err != nil {
		fmt.Printf("Error::%s\n", err.Error())
	}
	cfg = gconfig.Gcg

	//read config
	host := cfg.GetString("maindb.host")
	port := cfg.GetInt("maindb.port")

```

### Usage: command line flags
```go
	go run main.go -profile=stage -path=/Users/puran/server/config
```
   
   
