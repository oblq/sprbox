# SpareBox 

![GitHub tag](https://img.shields.io/github/tag/oblq/sprbox.svg)
[![Build Status](https://travis-ci.org/oblq/sprbox.svg?branch=master)](https://travis-ci.org/oblq/sprbox)
[![codecov](https://codecov.io/gh/oblq/sprbox/branch/master/graph/badge.svg)](https://codecov.io/gh/oblq/sprbox)
[![Go Report Card](https://goreportcard.com/badge/github.com/oblq/sprbox)](https://goreportcard.com/report/github.com/oblq/sprbox)
![dependencies Status](https://img.shields.io/badge/dependencies-none-brightgreen.svg)
[![GoDoc](https://godoc.org/github.com/oblq/sprbox?status.svg)](https://godoc.org/github.com/oblq/sprbox)
[![MIT license](https://img.shields.io/badge/License-MIT-blue.svg)](https://lbesson.mit-license.org/)

**SpareBox** is a toolbox factory with automatic package (struct or struct pointers) initialization and/or configuration based on your build environment.

## Installation

```sh
go get -u github.com/oblq/sprbox
```

## Getting Started

1. Define the build environment.  
    The build environment is determined matching a ***tag*** against some RegEx.  
    SpareBox will try to grab that tag in three different ways, in a precise order, if one can't be determined it will check for the next one:
    
    - The `BUILDENV` var in sprbox package, which can be defined by code:
        ```go
        sprbox.BUILDENV = "dev"
        ```
        ...or, since it is an exported string, can also be interpolated with `-ldflags` at build/run time:  
        ```bash
        LDFLAGS="-X ${GOPATH:-$HOME/go}/src/github.com/oblq/sprbox.BUILDENV=develop"
        go build -ldflags "${LDFLAGS}" -v -o ./api_bin ./api
        ```
    
    - The environment variable `'BUILD_ENV'`:
        ```go
        // sprbox.EnvVarKey is 'BUILD_ENV'
        os.Setenv(sprbox.EnvVarKey, "dev")
        ```
    
    - The Git branch name (Gitflow for CI/CD are supported).  
    By default the working dir is used, you can pass a different git repository path for this:  
        ```go
        sprbox.VCS = sprbox.NewRepository("path/to/repo")
        println(sprbox.VCS.BranchName) // Commit, Tag, Build, Path and Error
        sprbox.VCS.PrintInfo()
        ```  
    
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
    
    Be able to edit the RegEx means that it is possible to use everything to determine the build env, may be the machine hostname (`cat /etc/hostname`) or whatever you prefer.  
    SpareBox will grab the correct config file for any given build environment.  
    
    Finally you can check the current env in code with:
    
    ```go
    if sprbox.Env() == sprbox.Production { 
    	doSomething() 
    }
    println(sprbox.Env().String())
    sprbox.Env().PrintInfo()
    
    sprbox.SubPathByEnv("static") // -> "static/<environment>"
    ```
     
2. Prepare your config files.  
    Given a config path `'config/'` in your project, searching for a config file 'tool.yaml' in SpareBox is equal to searching for (in that order):
    - `'config/tool.<environment>.yaml'` (overrides `'config/tool.yaml'` for a specific environment only)
    - `'config/tool.yaml'` (valid for any environment, overrides `'config/<environment>/tool.yaml'`)
    - `'config/<environment>/tool.yaml'` (valid for a specific environment only)

    ...so it is possible to use `'config/tool.yaml'` for any build environment or override it for a specific environment only with `'config/tool.<environment>.yaml'` or create a subfolder for every environment having all the configs completely separated:  
    ```
    ├── config
    │   ├── development
    │   ├── local
    │   ├── production
    │   ├── staging
    │   └── testing
    └── main.go
    ```
    *SpareBox is agnostic about the config file extension, so it is also possible to pass just the file name without extension: `'tool'`*.
    
3. Define your toolbox (struct), fields can be of any type, sprbox will init nil struct pointers and pass config files where needed (structs or struct pointers).  
To load a configuration file your struct must implement the [configurable](#using-your-package-in-sprbox) interface, you can optionally pass the config file name in tags:  
    ```go
    type MyToolBox struct {
        // By default sprbox will look for a file named like the struct field (ATool),
        // case sensitive and with arbitrary extension, 
        // could be .yml, .json, .toml or any other, SpareBox is agnostic.
        ATool Tool
    
        // Define the config file explicitly, extension can be omitted.    
        WPool *workerful.Workerful `config:"workerpool.yml"`
    
        // ...or you can skip initialization and initialize it later by code
        AnOmittedTool Tool `omit:"true"`
    }
    
    var ToolBox MyToolBox
    ```

4. Init and configure the toolbox in one line.  

    ```go
    sprbox.PrintInfo(false) // optionally
    sprbox.Load(&ToolBox, "./config")
    ```

5. ...and the result is:  
![loading](./infoscreen.png)

## Using your package in sprbox

To start using your package in `sprbox` you just need to implement the `configurable` interface:

```go
type configurable interface {
	SBConfig(string) error
}
```

For instance:

```go
type MyPackage struct {
	Something string `yaml:"something"`
}

// SBConfig is the sprbox 'configurable' interface implementation.
func (mp *MyPackage) SBConfig(configPath string) error {
	if configFile, err := ioutil.ReadFile(configPath); err != nil {
		return fmt.Errorf("wrong config path: %s", err.Error())
	} else if err = yaml.Unmarshal(configFile, mp); err != nil {
		return fmt.Errorf("can't unmarshal config file: %s", err.Error())
	}
	return nil
}
```
    
Add `sprbox` in your repo topics if you like it.

## Examples
- [example](example)

## Compatible tools

- [`workerful`](https://github.com/oblq/workerful) Full-featured worker-pool implementation.

## Author

- [Marco Muratori](mailto:marcomrtr@gmail.com) 

## License

Workerful is available under the MIT license. See the [LICENSE](./LICENSE) file for more information.
