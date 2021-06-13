package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/haveachin/infrared"
	"github.com/haveachin/infrared/connection"
	"github.com/haveachin/infrared/gateway"
	"github.com/haveachin/infrared/proxy"
	"github.com/haveachin/infrared/server"
)

const (
	envPrefix     = "INFRARED_"
	envConfigPath = envPrefix + "CONFIG_PATH"
)

const (
	clfConfigPath = "config-path"
)

var (
	configPath = "./configs"
)

func envBool(name string, value bool) bool {
	envString := os.Getenv(name)
	if envString == "" {
		return value
	}

	envBool, err := strconv.ParseBool(envString)
	if err != nil {
		return value
	}

	return envBool
}

func envString(name string, value string) string {
	envString := os.Getenv(name)
	if envString == "" {
		return value
	}

	return envString
}

func initEnv() {
	configPath = envString(envConfigPath, configPath)
}

func initFlags() {
	flag.StringVar(&configPath, clfConfigPath, configPath, "path of all proxy configs")
	flag.Parse()
}

func init() {
	initEnv()
	initFlags()
}

func main() {
	fmt.Println("starting going to setup proxylane")
	serverCfgs := []server.ServerConfig{
		{
			NumberOfInstances: 1,
			DomainName:        "localhost",
			ProxyTo:           ":25560",
			RealIP:            false,
			OnlineStatus:      infrared.StatusConfig{},
			OfflineStatus:     infrared.StatusConfig{VersionName: "Infrared-1"},
		},
		{
			NumberOfInstances: 2,
			DomainName:        "127.0.0.1",
			ProxyTo:           ":25560",
			RealIP:            false,
			OnlineStatus:      infrared.StatusConfig{},
			OfflineStatus:     infrared.StatusConfig{VersionName: "Infrared-2"},
		},
	}

	connFactoryFactory := func(timeout time.Duration) (connection.ServerConnFactory, error) {
		return func(addr string) (connection.ServerConn, error) {
			c, err := net.DialTimeout("tcp", addr, timeout)
			if err != nil {
				return connection.ServerConn{}, err
			}
			return connection.NewServerConn(c), nil
		}, nil
	}
	outerListenerFactory := func(addr string) gateway.OuterListener {
		return gateway.NewBasicOuterListener(addr)
	}

	proxyCfg := proxy.ProxyLaneConfig{
		NumberOfListeners: 2,
		NumberOfGateways:  4,

		Timeout:  250,
		ListenTo: ":25565",
		Servers:  serverCfgs,

		ServerConnFactory:    connFactoryFactory,
		OuterListenerFactory: outerListenerFactory,
	}

	proxyLane := proxy.ProxyLane{Config: proxyCfg}
	proxyLane.StartupProxy()

	fmt.Println("finished setting up proxylane")

	wg := sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()

}

// func main() {
// 	log.Println("Loading proxy configs")

// 	cfgs, err := infrared.LoadProxyConfigsFromPath(configPath, false)
// 	if err != nil {
// 		log.Printf("Failed loading proxy configs from %s; error: %s", configPath, err)
// 		return
// 	}

// 	var proxies []*infrared.Proxy
// 	for _, cfg := range cfgs {
// 		proxies = append(proxies, &infrared.Proxy{
// 			Config: cfg,
// 		})
// 	}

// 	outCfgs := make(chan *infrared.ProxyConfig)
// 	go func() {
// 		if err := infrared.WatchProxyConfigFolder(configPath, outCfgs); err != nil {
// 			log.Println("Failed watching config folder; error:", err)
// 			log.Println("SYSTEM FAILURE: CONFIG WATCHER FAILED")
// 		}
// 	}()

// 	gateway := infrared.Gateway{}
// 	go func() {
// 		for {
// 			cfg, ok := <-outCfgs
// 			if !ok {
// 				return
// 			}

// 			proxy := &infrared.Proxy{Config: cfg}
// 			proxy.ServerFactory = func (p *infrared.Proxy) infrared.MCServer {
// 				timeout := p.Timeout()
// 				serverAddr := p.ProxyTo()
// 				return &infrared.BasicServer{
// 					ServerAddr: serverAddr,
// 					Timeout: timeout,
// 				}
// 			}
// 			if err := gateway.RegisterProxy(proxy); err != nil {
// 				log.Println("Failed registering proxy; error:", err)
// 			}
// 		}
// 	}()

// 	log.Println("Starting Infrared")
// 	if err := gateway.ListenAndServe(proxies); err != nil {
// 		log.Fatal("Gateway exited; error:", err)
// 	}

// 	gateway.KeepProcessActive()
// }
