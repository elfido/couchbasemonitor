package main

import (
	"cbmonitor/internal/config"
	"cbmonitor/internal/monitor"
	"flag"
	"log"
	"os"
	"time"
)

// ToDo:
//  - Add prometheus metrics
//  - Create docker file

func exitOnError(message string, err error) {
	if err != nil {
		log.Printf("%s: %s\n", message, err)
		os.Exit(1)
	}
}

func main() {
	configFile := flag.String("config", "./config.json", "Configuration file path")
	scrapInterval := flag.Duration("interval", 15*time.Second, "Monitoring interval")
	callsTimeout := flag.Duration("timeout", 3*time.Second, "Monitoring call timeout")
	defaultPassword := flag.String("password", "", "Default password (if you don't want to set one in config file)")
	flag.Parse()
	configuration, err := config.NewFileConfiguration(*configFile)
	exitOnError("Cannot read configuration", err)
	log.Printf("Using configuration from file: %s, found %d clusters", *configFile, len(configuration))
	monitors := make([]*monitor.Monitor, len(configuration))
	for i, cluster := range configuration {
		pass := cluster.Credentials.Password
		if pass == "" {
			pass = *defaultPassword
		}
		log.Printf("\t- %s @ %s://%s:%s\n", cluster.Name, cluster.Protocol, cluster.Hostname, cluster.Port)
		monitor, err := monitor.NewMonitor(cluster.Hostname, cluster.Name, cluster.Credentials.Username,
			pass, cluster.Protocol, cluster.Port)
		exitOnError("Cannot create monitor", err)
		monitor.SetTimeout(*callsTimeout)
		monitors[i] = monitor.Build()
	}

	for {
		log.Println("Processing")
		for _, monitor := range monitors {
			monitor.Check()
		}
		time.Sleep(*scrapInterval)
	}
}
