package services

import (
	"bytes"
	"errors"
	"fmt"
	"net/url"
	"text/template"

	"github.com/oblq/sprbox"
)

// Service is an abstraction of a service (or microservice) or the monolith itself.
// it holds some basic information useful for service discovery.
type Service struct {
	// Name is the service name.
	Name string `yaml:"Name"`

	// Version of the service.
	Version string `sprbox:"default=1" yaml:"Version"`

	// ProxyService is optional, default no proxy.
	// It only works in ServicesSlice:
	// the proxy will be automatically populated based
	// on the proxy service name.
	ProxyService string `yaml:"ProxyService"`

	// Proxy will be automatically populated in ServicesSlice.
	Proxy *Service

	// IPList contains the ip list of the machines running this service
	// in the format <public:private> (e.g. 192.168.1.10: 127.0.0.1 locally).
	// use <tasks.serviceName>:<tasks.serviceName> in Docker swarm
	// use <serviceName>:<serviceName> in Docker
	IPs map[string]string `yaml:"IPs"`

	// Port 443 automatically set https scheme when you get the service url.
	// Port 80 and all the others automatically set http scheme when you get the service url.
	Port int `sprbox:"default=80" yaml:"Port"`

	// Hosts contains the host names pointing to this service.
	// The first one will be used to build the service URL,
	// others may be useful for CORS config or whatever you need
	// and they're used in URLAlternatives().
	Hosts []string `yaml:"Hosts"`

	// BasepathName is optional, it will be parsed by
	// the template package, so you can use placeholders here
	// (eg.: "{{.Name}}/v{{.Version}}")
	Basepath string `yaml:"Basepath"`

	// Data is optional, set custom data here.
	Data map[string]interface{} `yaml:"Data"`
}

// SpareConfig is the sprbox configurable interface.
func (s *Service) SpareConfig(configFiles []string) (err error) {
	if err = sprbox.LoadConfig(s, configFiles...); err == nil {
		if sprbox.Env() == sprbox.Local {
			if len(s.Hosts) > 0 {
				overridenHosts := []string{PublicIP}
				s.Hosts = append(overridenHosts, s.Hosts...)
			} else {
				s.Hosts = append(s.Hosts, PublicIP)
			}
			//fmt.Printf(`
			//Env() == local: service's primary host overriden by outbound IP (%s),
			//that will make possible to connect to it through your local network (wi-fi) also.
			//`, PublicIP)
		}
		return s.parseBasePath()
	}
	return
}

// SpareConfig is the sprbox configurableInCollections interface.
func (s *Service) SpareConfigBytes(configBytes []byte) (err error) {
	if err = sprbox.Unmarshal(configBytes, s); err == nil {
		if sprbox.Env() == sprbox.Local {
			if len(s.Hosts) > 0 {
				overridenHosts := []string{PublicIP}
				s.Hosts = append(overridenHosts, s.Hosts...)
			} else {
				s.Hosts = append(s.Hosts, PublicIP)
			}
			//fmt.Printf(`
			//Env() == local: service's primary host overriden by outbound IP (%s),
			//that will make possible to connect to it through your local network (wi-fi) also.
			//`, PublicIP)
		}
		return s.parseBasePath()
	}
	return
}

func (s *Service) parseBasePath() error {
	basePathTemp, err := template.New("basepath").Parse(s.Basepath)
	if err != nil {
		return errors.New("invalid basepath template: " + s.Basepath)
	}
	buff := &bytes.Buffer{}
	if err := basePathTemp.Execute(buff, s); err != nil {
		return err
	}
	s.Basepath = buff.String()
	return nil
}

// Scheme returns the service scheme (http or https), based on service port.
func (s *Service) scheme() string {
	if s.Port == 443 {
		return "https"
	}
	return "http"
}

// IP returns the service private ip address.
// If not found it returns the public one (or the service name if docker is used).
func (s *Service) IP() string {
	if privateIP, ok := s.IPs[PublicIP]; ok {
		return privateIP
	} else if dockerServiceName, ok := s.IPs[s.Name]; ok {
		return dockerServiceName
	} else if swarmTaskName, ok := s.IPs["tasks."+s.Name]; ok {
		return swarmTaskName
	}
	return PublicIP
}

// Host returns the first host listed in the config file.
func (s *Service) Host() string {
	if len(s.Hosts) > 0 {
		return s.Hosts[0]
	}
	return ""
}

// PortOptional returns the service port for URL.
// If service.Port == ':443' or ':80' returns an empty string.
func (s *Service) portOptional() (port string) {
	if s.Port != 443 && s.Port != 80 {
		port = fmt.Sprintf(":%d", s.Port)
	}
	return
}

// URL returns the service URL by service name.
// If the service use a proxy the proxy scheme and port will be used.
func (s *Service) URL() *url.URL {
	nURL := &url.URL{}
	nURL.Scheme = s.scheme()
	nURL.Host = s.Host() + s.portOptional()
	nURL.Path = s.Basepath
	return nURL
}

// ProxyURL returns the service Proxy URL safely,
// it will fallback to the standard URL if no Proxy is set.
func (s *Service) ProxyURL() *url.URL {
	if s.Proxy != nil {
		nURL := &url.URL{}
		nURL.Scheme = s.Proxy.scheme()
		nURL.Host += s.Proxy.Host() + s.Proxy.portOptional()
		nURL.Path = s.Basepath
		return nURL
	} else {
		return s.URL()
	}
}

// URLAlternatives returns the service alternative URLs by service name.
// If the service use a proxy the proxy scheme and port will be used.
func (s *Service) URLAlternatives() ([]url.URL, []string) {
	urls := make([]url.URL, 0)
	urlsString := make([]string, 0)

	for i := 1; i < len(s.Hosts); i++ {
		nURL := url.URL{}
		nURL.Scheme = s.scheme()
		nURL.Host = s.Hosts[i] + s.portOptional()
		nURL.Path = s.Basepath

		urls = append(urls, nURL)
		urlsString = append(urlsString, nURL.String())
	}

	return urls, urlsString
}
