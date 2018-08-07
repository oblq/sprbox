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
// (mp *MyPackage) is automatically initialized with a pointer to MyPackage{}
// so it will never be nil, but needs configuration.
func (s *ServicesMap) SpareConfig(configFiles []string) (err error) {
	if err = sprbox.LoadConfig(s, configFiles...); err == nil {
		return s.parseServices()
	}
	return
}

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
			//fmt.Printf(`
			//Env() == local: service's primary host overriden by outbound IP (%s),
			//that will make possible to connect to it through your local network (wi-fi) also.
			//`, PublicIP)
		}

		proxy, _ := (*s)[service.ProxyService]
		service.Proxy = proxy

		if err := service.parseBasePath(); err != nil {
			return err
		}
	}
	return nil
}
