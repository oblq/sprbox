package services

import (
	"fmt"
	"net"

	"github.com/oblq/sprbox"
)

type ServicesMap map[string]*Service

// The preferred outbound ip of this machine
// http://stackoverflow.com/a/37382208/3079922.
var PublicIP string

func init() {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	if err := conn.Close(); err != nil {
		fmt.Println(err.Error())
	}
	PublicIP = localAddr.IP.String()
}

// SpareConfig is the sprbox 'configurable' interface implementation.
func (s *ServicesMap) SpareConfig(configFiles []string) (err error) {
	if err = sprbox.LoadConfig(s, configFiles...); err == nil {
		return s.parseServices()
	}
	return
}

// if Env() == local the service's primary host
// will be overriden by outbound IP,
// that will make possible to connect to it
// through your local network (wi-fi) also.
func (s *ServicesMap) parseServices() error {
	for sName, service := range *s {
		service.Name = sName
		if sprbox.Env() == sprbox.Local {
			if len(service.Hosts) > 0 {
				overridenHosts := []string{PublicIP}
				service.Hosts = append(overridenHosts, service.Hosts...)
			} else {
				service.Hosts = append(service.Hosts, PublicIP)
			}
		}

		service.Proxy, _ = (*s)[service.ProxyService]

		// sprbox template parsing works only on struct, doing it manually
		if err := service.parseBasePath(); err != nil {
			return err
		}
	}
	return nil
}
