package main

import (
	"encoding/json"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	config struct {
		Modules      []string `json:"modules"`
		Downloads    string   `json:"downloads"`
		Repositories struct {
			Atom    string `json:"atom"`
			Vagrant string `json:"vagrant"`
		} `json:"repositories"`
	}

	flContinuous bool
	flConfig     string
	flModule     string

	cmdRoot = &cobra.Command{Use: "autoyumzr", Short: "Auto download RPM and create Yum repository", Run: start}
)

func init() {
	cmdRoot.PersistentFlags().BoolVarP(&flContinuous, "continuous", "n", false, "enable continuous mode")
	cmdRoot.PersistentFlags().StringVarP(&flConfig, "config", "c", "config.json", "configuration file")
	cmdRoot.PersistentFlags().StringVarP(&flModule, "module", "m", "all", "module to run")
	cobra.OnInitialize(initialize)
}

func main() {
	cmdRoot.Execute()
}

func initialize() {
	config.Modules = []string{"atom", "vagrant"}
	config.Downloads = "/tmp/downloads/"
	config.Repositories.Atom = "/tmp/repos/atom/"
	config.Repositories.Vagrant = "/tmp/repos/vagrant/"

	f, err := os.Open(flConfig)
	if err != nil {
		logrus.Fatal(err)
	}
	defer f.Close()
	err = json.NewDecoder(f).Decode(&config)
	if err != nil {
		logrus.Fatal(err)
	}
}
