/*
Copyright © 2021 Yale University

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package main

import (
	"bytes"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/YaleSpinup/neptune-api/api"
	"github.com/YaleSpinup/neptune-api/common"

	log "github.com/sirupsen/logrus"
)

var (
	// Version is the main version number
	Version = "0.0.0"

	// Buildstamp is the timestamp the binary was built, it should be set at buildtime with ldflags
	Buildstamp = "No BuildStamp Provided"

	// Githash is the git sha of the built binary, it should be set at buildtime with ldflags
	Githash = "No Git Commit Provided"

	configFileName = flag.String("config", "config/config.json", "Configuration file.")
	version        = flag.Bool("version", false, "Display version information and exit.")
)

func main() {
	flag.Parse()
	if *version {
		vers()
	}

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal("unable to get working directory")
	}
	log.Infof("Starting neptune-api version %s (%s)", Version, cwd)

	config, err := common.ReadConfig(configReader())
	if err != nil {
		log.Fatalf("Unable to read configuration from: %+v", err)
	}

	config.Version = common.Version{
		Version:           Version,
		BuildStamp:        Buildstamp,
		GitHash:           Githash,
	}

	// Set the loglevel, info if it's unset
	switch config.LogLevel {
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}

	if config.LogLevel == "debug" {
		log.Debug("Starting profiler on 127.0.0.1:6080")
		go http.ListenAndServe("127.0.0.1:6080", nil)
	}
	log.Debugf("loaded configuration: %+v", config)

	if err := api.NewServer(config); err != nil {
		log.Fatal(err)
	}
}

func configReader() io.Reader {
	if configEnv := os.Getenv("API_CONFIG"); configEnv != "" {
		log.Infof("reading configuration from API_CONFIG environment")

		c, err := base64.StdEncoding.DecodeString(configEnv)
		if err != nil {
			log.Infof("API_CONFIG is not base64 encoded")
			c = []byte(configEnv)
		}

		return bytes.NewReader(c)
	}

	log.Infof("reading configuration from %s", *configFileName)

	configFile, err := os.Open(*configFileName)
	if err != nil {
		log.Fatalln("unable to open config file", err)
	}

	c, err := ioutil.ReadAll(configFile)
	if err != nil {
		log.Fatalln("unable to read config file", err)
	}

	return bytes.NewReader(c)
}

func vers() {
	fmt.Printf("neptune-api Version: %s\n", Version)
	os.Exit(0)
}
