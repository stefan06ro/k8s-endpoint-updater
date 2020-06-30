package main

import (
	"os"

	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/k8s-endpoint-updater/command"
)

var (
	description string = "Command line tool for updating Kubernetes endpoints based on given configuration."
	gitCommit   string = "n/a"
	name        string = "k8s-endpoint-updater"
	source      string = "https://github.com/giantswarm/k8s-endpoint-updater"
)

func main() {
	var err error

	// Create a new logger which is used by all packages.
	var newLogger micrologger.Logger
	{
		loggerConfig := micrologger.Config{
			IOWriter: os.Stdout,
		}
		newLogger, err = micrologger.New(loggerConfig)
		if err != nil {
			panic(err)
		}
	}

	var newCommand *command.Command
	{
		commandConfig := command.DefaultConfig()

		commandConfig.Logger = newLogger

		commandConfig.Description = description
		commandConfig.GitCommit = gitCommit
		commandConfig.Name = name
		commandConfig.Source = source

		newCommand, err = command.New(commandConfig)
		if err != nil {
			panic(err)
		}
	}

	err = newCommand.CobraCommand().Execute()
	if err != nil {
		panic(err)
	}
}
