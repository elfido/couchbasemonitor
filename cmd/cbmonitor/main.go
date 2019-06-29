package main

import (
	"cbmonitor/internal/config"
	"flag"
	"log"
	"os"
)

func exitOnError(message string, err error) {
	if err != nil {
		log.Printf("%s: %s\n", message, err)
		os.Exit(1)
	}
}

func main() {
	configFile := flag.String("config", "./config.json", "Configuration file path")
	flag.Parse()
	configuration, err := config.NewFileConfiguration(*configFile)
	exitOnError("Cannot read configuration", err)
	log.Printf("Using configuration from file: %s, found %d clusters", *configFile, len(configuration))
	for _, cluster := range configuration {
		log.Printf("\t- %s @ %s\n", cluster.Name, cluster.Hostname)
	}
}
