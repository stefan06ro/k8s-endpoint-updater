// Package update implements the update command for the command line tool.
package update

import (
	"fmt"
	"os"

	microerror "github.com/giantswarm/microkit/error"
	micrologger "github.com/giantswarm/microkit/logger"
	"github.com/spf13/cobra"

	"github.com/giantswarm/k8s-endpoint-updater/command/update/flag"
	"github.com/giantswarm/k8s-endpoint-updater/service/provider"
	"github.com/giantswarm/k8s-endpoint-updater/service/provider/env"
)

var (
	f = &flag.Flag{}
)

// Config represents the configuration used to create a new update command.
type Config struct {
	// Dependencies.
	Logger micrologger.Logger
}

// DefaultConfig provides a default configuration to create a new update
// command by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		Logger: nil,
	}
}

// New creates a new configured update command.
func New(config Config) (*Command, error) {
	// Dependencies.
	if config.Logger == nil {
		return nil, microerror.MaskAnyf(invalidConfigError, "logger must not be empty")
	}

	newCommand := &Command{
		// Dependencies.
		logger: config.Logger,

		// Internals.
		cobraCommand: nil,
	}

	newCommand.cobraCommand = &cobra.Command{
		Use:   "update",
		Short: "Update Kubernetes endpoints based on given configuration.",
		Long:  "Update Kubernetes endpoints based on given configuration.",
		Run:   newCommand.Execute,
	}

	newCommand.cobraCommand.PersistentFlags().StringSliceVar(&f.Pod.Names, "pod.names", nil, "List of pod names used to lookup pod IPs.")
	newCommand.cobraCommand.PersistentFlags().StringVar(&f.Provider.Env.Prefix, "provider.env.prefix", "K8S_ENDPOINT_UPDATER_POD_", "Prefix of environment variables providing pod names.")
	newCommand.cobraCommand.PersistentFlags().StringVar(&f.Provider.Kind, "provider.kind", "env", "Provider used to lookup pod IPs.")

	return newCommand, nil
}

type Command struct {
	// Dependencies.
	logger micrologger.Logger

	// Internals.
	cobraCommand *cobra.Command
}

func (c *Command) CobraCommand() *cobra.Command {
	return c.cobraCommand
}

// TODO
func (c *Command) Execute(cmd *cobra.Command, args []string) {
	c.logger.Log("info", "start updating Kubernetes endpoint")

	err := f.Validate()
	if err != nil {
		c.logger.Log("error", fmt.Sprintf("%#v", microerror.MaskAny(err)))
		os.Exit(1)
	}

	err = c.execute()
	if err != nil {
		c.logger.Log("error", fmt.Sprintf("%#v", microerror.MaskAny(err)))
		os.Exit(1)
	}

	c.logger.Log("info", "finished updating Kubernetes endpoint")
}

func (c *Command) execute() error {
	var err error

	var newProvider provider.Provider
	{
		k := f.Provider.Kind
		switch k {
		case env.Kind:
			envConfig := env.DefaultConfig()
			envConfig.Logger = c.logger
			envConfig.PodNames = f.Pod.Names
			envConfig.Prefix = f.Provider.Env.Prefix
			newProvider, err = env.New(envConfig)
			if err != nil {
				return microerror.MaskAny(err)
			}
		}
	}

	var podInfos []provider.PodInfo
	{
		podInfos, err = newProvider.Lookup()
		if err != nil {
			return microerror.MaskAny(err)
		}
	}

	fmt.Printf("%#v\n", podInfos)

	// TODO
	//{
	//	err = c.updater.Update(podInfos)
	//	if err != nil {
	//		return microerror.MaskAny(err)
	//	}
	//}

	return nil
}
