package main

import (
	"flag"
	"narwhal/internal"
	"narwhal/service"
	"os"

	log "github.com/sirupsen/logrus"
)

func checkErr(err error) {
	if err != nil {
		log.Panic(err)
		os.Exit(1)
	}
}

func registrySingal() error {
	return nil
}

func main() {
	//Parse command line parms and config file
	confFile := flag.String("config", "/etc/narwhal/narwhal.yaml", "Narwhal confige file")
	debug := flag.Bool("debug", false, "Show debug info, default False")
	flag.Parse()
	iconf, debug_enable, err := internal.ParseConfig(*confFile)
	checkErr(err)

	// Setup logger
	internal.NewLogger(debug_enable || *debug)

	// Handle signal
	checkErr(registrySingal())

	// Launch service
	switch conf := iconf.(type) {
	case *internal.ServerConf:
		err := service.RunServer(conf)
		checkErr(err)
	case *internal.ClientConf:
		err := service.RunClient(conf)
		checkErr(err)
	}
}
