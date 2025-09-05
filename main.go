package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/opensourceways/image-scanning/common/infrastructure/postgresql"
	"github.com/opensourceways/image-scanning/config"
	"github.com/opensourceways/image-scanning/scanning"
)

const (
	port        = 8888
	gracePeriod = 180
)

type options struct {
	service     ServiceOptions
	enableDebug bool
}

// ServiceOptions defines configuration parameters for the service.
type ServiceOptions struct {
	Port        int
	ConfigFile  string
	Cert        string
	Key         string
	GracePeriod time.Duration
	RemoveCfg   bool
}

// Validate checks if the ServiceOptions are valid.
// It returns an error if the config file is missing.
func (o *ServiceOptions) Validate() error {
	if o.ConfigFile == "" {
		return fmt.Errorf("missing config-file")
	}

	return nil
}

// AddFlags adds flags for ServiceOptions to the provided FlagSet.
// It includes flags for port, remove-config, config-file, cert, key, and grace-period.
func (o *ServiceOptions) AddFlags(fs *flag.FlagSet) {
	fs.IntVar(&o.Port, "port", port, "Port to listen on.")
	fs.BoolVar(&o.RemoveCfg, "rm-cfg", false, "whether remove the cfg file after initialized .")

	fs.StringVar(&o.ConfigFile, "config-file", "", "Path to config file.")
	fs.StringVar(&o.Cert, "cert", "", "Path to tls cert file.")
	fs.StringVar(&o.Key, "key", "", "Path to tls key file.")
	fs.DurationVar(&o.GracePeriod, "grace-period", time.Duration(gracePeriod)*time.Second,
		"On shutdown, try to handle remaining events for the specified duration.")
}

func (o *options) Validate() error {
	return o.service.Validate()
}

func gatherOptions(fs *flag.FlagSet, args ...string) options {
	var o options

	o.service.AddFlags(fs)

	fs.BoolVar(
		&o.enableDebug, "enable_debug", false, "whether to enable debug model.",
	)

	fs.Parse(args)

	return o
}

func main() {
	o := gatherOptions(
		flag.NewFlagSet(os.Args[0], flag.ExitOnError),
		os.Args[1:]...,
	)
	if err := o.Validate(); err != nil {
		logrus.Errorf("Invalid options, err:%s", err.Error())

		return
	}

	if o.enableDebug {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.Debug("debug enabled.")
	}

	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		DisableColors: true,
		DisableQuote:  true,
	})

	// cfg
	cfg := new(config.Config)
	if err := config.LoadConfig(o.service.ConfigFile, cfg, false); err != nil {
		logrus.Errorf("load config, err:%s", err.Error())

		return
	}

	// postgresql
	if err := postgresql.Init(&cfg.Postgresql, o.service.RemoveCfg); err != nil {
		logrus.Errorf("init db failed, err:%s", err.Error())

		return
	}

	go healthCheck()

	scanning.Run(cfg)
}

func healthCheck() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "i am ok")
	})
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}