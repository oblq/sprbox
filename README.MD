# SpareBox 

[![awesome](https://raw.githubusercontent.com/istio/fortio/master/docs/mentioned-badge.svg?sanitize=true)](https://github.com/avelino/awesome-go)
![GitHub tag](https://img.shields.io/github/tag/oblq/sprbox.svg)
[![Build Status](https://travis-ci.org/oblq/sprbox.svg?branch=master)](https://travis-ci.org/oblq/sprbox)
[![codecov](https://codecov.io/gh/oblq/sprbox/branch/master/graph/badge.svg)](https://codecov.io/gh/oblq/sprbox)
[![Go Report Card](https://goreportcard.com/badge/github.com/oblq/sprbox)](https://goreportcard.com/report/github.com/oblq/sprbox)
[![GoDoc](https://godoc.org/github.com/oblq/sprbox?status.svg)](https://godoc.org/github.com/oblq/sprbox)
[![MIT license](https://img.shields.io/badge/License-MIT-blue.svg)](https://lbesson.mit-license.org/)


Dynamically create toolbox singletons with automatic configuration based on your build environment.  
SpareBox is also an agnostic, layered, config parser (supporting YAML, TOML, JSON and Environment vars).  
Keep your projects and their configuration files ordered and maintainable.

## Installation

Using [dep](https://github.com/golang/dep):

```sh
dep ensure -add github.com/oblq/sprbox@master
```

...or go get:

```sh
go get -u github.com/oblq/sprbox
```

## ToolBox autoload (init and config)

##### 1. Define your toolbox.

Fields can be of any type, sprbox will init pointers and pass config files where needed (configurable structs, struct pointers, slices or maps).  
To load a configuration file a field must implement the [configurable](#using-your-package-in-sprbox) interface.  

```go
type ToolBox struct {
	// By default sprbox will look for a file named like the
	// struct field name (Services.*, case sensitive).
	// Services.<environment>.yml will override Services.yml
	// for the given env, if exist.
	Services services.ServicesMap

	// MediaProcessing does not implement the 'configurable' interface
	// so it will be traversed recursively.
	// Recursion only stop when no more embedded elements are found 
	// or when a 'configurable' element is found instead.
	// 'configurable' elements will not be traversed.
	MediaProcessing struct {
		// Optionally pass one or more specific config file name, 
		// separated by the pipe symbol: |.
		// The latest will overrides others, from right to left,
		// you can see that using sprbox.SetDebug(true).
		// sprbox will always try to find the file named
		// like the struct field first (Pictures.*).
		// File extension can be omitted.
		//
		// Configurable elements inside an embedded struct, like this,
		// will search for config files in a sub-directory
		// eg.: "./config/MediaProcessing/Pics.*", not "./config/Pics.*"
		Pictures services.Service `sprbox:"Pics|PicsOverride"`
		Videos   services.Service
	}

	WP  Workerful
	// Workerful implement the 'configurableInCollections' interface,
	// so it can be loaded also directly inside 
	// slices or maps using a single config file.
	WPS []Workerful 

	// will print the error in console
	ToolMissingConfig *Tool

	// Optionally add the 'omit' value to skip a field.
	OmittedTool Tool `sprbox:"omit"`
}

var ToolBox MyToolBox
```

##### 2. Init and configure the toolbox in one line.  

In `sprbox.LoadToolBox()` environment-specific config files (`cfg.<environment>.*`) will override generic ones (`cfg.*`):  

```go
sprbox.PrintInfo()
// Optionally set debug mode.
// sprbox.SetDebug(true)
sprbox.LoadToolBox(&ToolBox, "./config")
```

NOTE: tool's nil anonymous struct fields will be automatically initialized before to call the [configurable](#using-your-package-in-sprbox) interface.

![loading](start.png)

## The build environment
 
The build environment is determined matching a ***tag*** against some predefined environment specific RegEx, since any of the env's RegEx can be edited users have the maximum flexibility on the method to use.  
For instance, the machine hostname (`cat /etc/hostname`) can be used.

sprbox will try to grab that tag in three different ways, in a precise order, if one can't be determined it will check for the next one:

1. The `BUILDENV` var in sprbox package:
    ```go
    sprbox.BUILDENV = "dev"
    ```
    Since it is an exported string, can also be interpolated with `-ldflags` at build/run time:  
    ```bash
    LDFLAGS="-X ${GOPATH:-$HOME/go}/src/github.com/oblq/sprbox.BUILDENV=develop"
    go build -ldflags "${LDFLAGS}" -v -o ./api_bin ./api
    ```

2. The environment variable `'BUILD_ENV'`:
    ```go
    // sprbox.EnvVarKey is 'BUILD_ENV'
    os.Setenv(sprbox.EnvVarKey, "dev")
    ```

3. The Git branch name (Gitflow supported).  
By default the working dir is used, you can pass a different git repository path for this:  
    ```go
    sprbox.VCS = sprbox.NewRepository("path/to/repo")
    println(sprbox.VCS.BranchName) // Commit, Tag, Build, Path and Error
    sprbox.VCS.PrintInfo()
    ```  
4. When you run tests the environment will be set automatically to 'testing' if not set manually and no git repo is found in the project root.

Every environment has a set of default RegEx:

```
Production  = []string{"production", "master"}
Staging     = []string{"staging", "release/*", "hotfix/*"}
Testing     = []string{"testing", "test", "feature/*"}
Development = []string{"development", "develop", "dev"}
Local       = []string{"local"}
```

...and they can be edited:

```go
sprbox.Testing.SetExps([]string{"testing", "test"})
sprbox.Testing.AppendExp("feature/f*")
println("matched:", sprbox.Testing.MatchTag("feature/f5"))
```  

Finally you can check the current env in code with:

```go
if sprbox.Env() == sprbox.Production { 
    doSomething() 
}

sprbox.Env().PrintInfo()
```

##### Working with directories

Sparebox offer two utility funcs to work with directories.

1. `EnvSubDir()`  
...pretty much self-explanatory:
    ```go
    sprbox.EnvSubDir("static") // -> "static/<environment>" (eg.: "static/staging")
    ```
2. `CompiledPath()`  
If the current build-environment has RunCompiled == true `sprbox.CompiledPath()` returns the path base, so static files can stay side by side with the executable while it is possible to have a different location when the program is launched with `go run`.  
This allow to manage multiple packages in one project during development, for instance using a config path in the parent dir, side by side with the packages, while having the same config folder side by side with the executable where needed.
    ```go
    sprbox.BUILDENV = sprbox.Development.ID()   
    
    sprbox.Development.RunCompiled = false   
    sprbox.CompiledPath("../static_files/config") // -> "../static_files/config"
      
    sprbox.Development.RunCompiled = true   
    sprbox.CompiledPath("../static_files/config") // -> "config"
    ```
    A simple usage example is:
    ```go
    sprbox.LoadToolBox(&myToolBox, sprbox.CompiledPath("../config"))
    ```
    By default only Production and Staging environments have `RunCompiled` true.
## Using your package in sprbox

To start using your package in `sprbox` you just need to implement the `configurable` interface:

```go
type configurable interface {
	SpareConfig([]string) error
}

// optional, allow to load the package from a slice or a map directly.
type configurableInCollection interface {
	SpareConfigBytes([]byte) error
}
```

Example:

```go
type MyPackage struct {
	Something string `yaml:"something"`
}

// SpareConfig is the sprbox 'configurable' interface implementation.
// (mp *MyPackage) is automatically initialized with a pointer to MyPackage{}
// so it will never be nil, but needs configuration.
func (mp *MyPackage) SpareConfig(configFiles []string) (err error) {
	var config *MyPackageConfig
	err = sprbox.LoadConfig(&cfg, configFiles...)    	
	mp.DoSomethingWithConfig(config)
	return
}

// SpareConfigBytes optionally allow to load MyPackage inside a slice or a map directly.
func (mp *MyPackage) SpareConfigBytes(configBytes []byte) (err error) {
	var config *MyPackageConfig
	err = sprbox.Unmarshal(configBytes, &cfg)
	mp.DoSomethingWithConfig(config)
	return
}
```

Add `sprbox` in your repo topics and/or the 'sprbox-ready' badge if you like it: [![sprbox](https://img.shields.io/badge/sprbox-ready-green.svg)](https://github.com/oblq/sprbox)  


## Embed third-party packages in sprbox

Suppose we want to embed `packagex.StructX`:

```go
type StructX struct {
	*packagex.StructX
}

func (sx *StructX) SpareConfig(configFiles []string) (err error) {
	var cfg packagex.Config
	err = sprbox.LoadConfig(&cfg, configFiles...)    	
	sx.StructX = packagex.NewStructX(cfg)
	return
}

// SpareConfigBytes optionally allow to load StructX inside a slice or a map directly.
func (sx *StructX) SpareConfigBytes(configBytes []byte) (err error) {
	var cfg packagex.Config
	err = sprbox.Unmarshal(configBytes, &cfg)
	sx.StructX = packagex.NewStructX(cfg)
	return
}
```

From here on you can use the StructX in a toolbox with automatic init/config:

```go
type ToolBox struct {
	SX StructX
}

var App ToolBox

func init() {
	// ./config must contain SX.(yml|yaml|json|toml) config file in that case.
	sprbox.LoadToolBox(&App, "./config") 
	
	// Call any of the packagex.StructX's funcs on SX.
	// Initialized and configured.
	App.SX.DoSomething()
}
```

## Agnostic, layered, config unmarshaling

Given that project structure:
```
├── config
│   ├── pg.yaml
│   └── pg.production.yaml
└── main.go
```

pg.yaml:

```yaml
port: 2222
```

pg.production.yaml:

```yaml
port: 2345
```

...to unmarshal that config files to a struct you just need to call `sprbox.LoadConfig(&pgConfig, "config/pg.yaml")`:

```go
package main

import (
	"fmt"
	"os"
	
	"github.com/oblq/sprbox"
)

type PostgresConfig struct{
	// Environment vars overrides both default values and config file provided values.
	DB       string `sprbox:"env=POSTGRES_DB,default=postgres"`
	User     string `sprbox:"env=POSTGRES_USER,default=postgres"`
	// If no value is found that will return an error: 'required'.
	Password string `sprbox:"env=POSTGRES_PASSWORD,required"`
	Port     int    `sprbox:"default=5432"`
} 

func main() {
	os.Setenv("POSTGRES_PASSWORD", "123_only_known_by_me")
	
	// Setting 'production' build-environment,
	// so 'pg.production.yml' will override 'pg.yml'.
	sprbox.BUILDENV = sprbox.Production.ID() // -> 'production'
	
	var pgConfig PostgresConfig	
	if err := sprbox.LoadConfig(&pgConfig, "config/pg.yaml"); err != nil {
		fmt.Println(err)
	}
	
	fmt.Printf("%#v\n", pgConfig) 
	// Config{
	//      DB:         "postgres"
	//      User:       "postgres"
	//      Password:   "123_only_known_by_me"
	//      Port:       2345
	// }
}
```

Depending on the [build environment](#the-build-environment), trying to load `config/pg.yml` will also load `config/pg.<environment>.yml` (eg.: `cfg.production.yml`).  
If any environment-specific file will be found, for the current environment, that will override the generic one.  

It is possible to load multiple separated config files, also of different type, so components configs can be reused:

```go
func main() {
	// File extension can be omitted:
	var pusherConfig PushNotificationsConfig	
	sprbox.LoadConfig(&pusherConfig, "config/pusher.yml", "config/postgres.json")
}
```

The file extension in the file path can be omitted, since sprbox can load YAML, TOML and JSON files it will search for `cfg.*` using RegEx, the config file itself must have an extension.  

Keep in mind that:
1. YAML files uses lowercased keys by default, unless you define a custom field tag (struct field `Postgres` will become `"postgres"`, while in TOML or JSON it will remain `"Postgres"`).
2. The default map interface is `map[interface{}]interface{}` in YAML, not `map[string]interface{}` as in JSON or TOML.

## Examples
- [example](example)

To start it run: 
```sh 
make example
```
 

## Ready packages

##### Included:
- [`common`](common)
    - [`Services`](common/services) Service/micro-service/monolith abstraction, get services URL, Proxy, Version, Basepath, hold custom Data etc...
    
##### External:
- [`Workerful`](https://github.com/oblq/workerful) Full-featured worker-pool implementation.

## Vendored packages

- [`gopkg.in/yaml.v2`](https://github.com/go-yaml/yaml)  
- [`github.com/BurntSushi/toml`](https://github.com/BurntSushi/toml)

## Author

- [Marco Muratori](mailto:marcomrtr@gmail.com) 

## License

SpareBox is available under the MIT license. See the [LICENSE](./LICENSE) file for more information.
