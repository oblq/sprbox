# SpareBox 

[![GitHub tag](https://img.shields.io/github/tag/oblq/sprbox.svg)](https://github.com/oblq/sprbox)
[![Build Status](https://travis-ci.org/oblq/sprbox.svg?branch=master)](https://travis-ci.org/oblq/sprbox)
[![codecov](https://codecov.io/gh/oblq/sprbox/branch/master/graph/badge.svg)](https://codecov.io/gh/oblq/sprbox)
[![Go Report Card](https://goreportcard.com/badge/github.com/oblq/sprbox)](https://goreportcard.com/report/github.com/oblq/sprbox)
![dependencies Status](https://img.shields.io/badge/dependencies-none-brightgreen.svg)
[![GoDoc](https://godoc.org/github.com/oblq/sprbox?status.svg)](https://godoc.org/github.com/oblq/sprbox)
[![MIT license](https://img.shields.io/badge/License-MIT-blue.svg)](https://lbesson.mit-license.org/)



**SpareBox** is a toolbox factory with automatic package configuration based on your build environment.

## Installation

```sh
go get -u github.com/oblq/sprbox
```

## Getting Started

1. In your project, create a config path with a subfolder for every environment you plan to use:  
    ```
    ├── config
    │   ├── development
    │   ├── local
    │   ├── production
    │   ├── staging
    │   └── testing
    └── main.go
    ```
    
2. Define your toolbox composed by [`boxable`](./#Using-your-package-in-sprbox) elements (see the [`boxable`](./#Using-your-package-in-sprbox) interface) 
:  
    ```go
    type MyToolBox struct {
        WPool workerful.Workerful `sprbox:"workerpool.yml"`
    
        // By default sprbox will look for a file named like the
        // struct field name (ATool.yml, case insensitive).
        ATool    Tool
        AToolPtr *Tool `sprbox:"atool.yml"`
    
        // ...or you can skip initialization and initialize it later by code
        AnOmittedTool Tool `sprbox:"omit"`
    }
    
    var ToolBox MyToolBox
    ```

3. Init and configure the toolbox for the right build environment in one line:  
    ```go
    sprbox.PrintInfo(false) // optionally
    sprbox.InitAndConfig(&ToolBox, "path/to/config")
    ```

4. ...and the result is:  
![loading](./infoscreen.png)

## The build environment

The build environment is inferred from a ***tag***, using RegEx to match it.  
SpareBox will try to grab that tag in three different ways, in a precise order, if one can't be determined it will check for the next one:

1. The `BUILDENV` var in sprbox package, which can be defined by code:
    ```go
    sprbox.BUILDENV = "dev"
    ```
    ...or, since it is an exported string, can also be interpolated with `-ldflags` at build/run time:  
    ```bash
    LDFLAGS="-X ${GOPATH:-$HOME/go}/src/github.com/oblq/sprbox.BUILDENV=develop"
    go build -ldflags "${LDFLAGS}" -v -o ./api_bin ./api
    ```

2. The environment variable `BUILD_ENV`:
    ```go
    os.Setenv(sprbox.EnvVarKey, "dev")
    ```

3. The branch name (Gitflow for CI/CD are supported).  
By default the working dir is used, you can pass a different git repository path for this:  
    ```go
    sprbox.VCS = sprbox.NewRepository("path/to/repo")
    println(sprbox.VCS.BranchName) // Commit, Tag, Build, Path and Error
    sprbox.VCS.PrintInfo()
    ```  

Every environment has a set of default patterns:

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

Be able to edit the RegEx means that you can use everything to determine the build env, may be the machine hostname (`cat /etc/hostname`) or whatever you prefer.  
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

## Using your package in sprbox

To start using your package in `sprbox` you just need to implement the `boxable` interface:

```go
type boxable interface {
	Go2Box(string) error
}
```

For instance:

```go
type MyPackage struct {
	Something string `yaml:"something"`
}

// Go2Box is the sprbox 'boxable' interface implementation.
func (mp *MyPackage) Go2Box(configPath string) error {
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
