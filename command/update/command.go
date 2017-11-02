// Package update implements the update command for the command line tool.
package update

import (
	"fmt"
	"os"

	"github.com/cenk/backoff"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/client/k8sclient"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/k8s-endpoint-updater/command/update/flag"
	"github.com/giantswarm/k8s-endpoint-updater/service/provider"
	"github.com/giantswarm/k8s-endpoint-updater/service/provider/bridge"
	"github.com/giantswarm/k8s-endpoint-updater/service/updater"
	"net"
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
		return nil, microerror.Maskf(invalidConfigError, "logger must not be empty")
	}

	newCommand := &Command{
		// Dependencies.
		logger: config.Logger,

		// Internals.
		cobraCommand: nil,
	}

	newCommand.cobraCommand = &cobra.Command{
		Use:   "update",
		Short: "Update annotations on KVM pod based on given configuration.",
		Long:  "Update annotations on KVM pod based on given configuration.",
		Run:   newCommand.Execute,
	}

	newCommand.CobraCommand().PersistentFlags().StringVar(&f.Kubernetes.Address, "service.kubernetes.address", "http://127.0.0.1:6443", "Address used to connect to Kubernetes. When empty in-cluster config is created.")
	newCommand.CobraCommand().PersistentFlags().StringVar(&f.Kubernetes.Cluster.Namespace, "service.kubernetes.cluster.namespace", "default", "Namespace of the guest cluster which endpoints should be updated.")
	newCommand.CobraCommand().PersistentFlags().StringVar(&f.Kubernetes.Cluster.Service, "service.kubernetes.cluster.service", "", "Name of the service which endpoints should be updated.")
	newCommand.CobraCommand().PersistentFlags().BoolVar(&f.Kubernetes.InCluster, "service.kubernetes.inCluster", false, "Whether to use the in-cluster config to authenticate with Kubernetes.")
	newCommand.CobraCommand().PersistentFlags().StringVar(&f.Kubernetes.TLS.CaFile, "service.kubernetes.tls.caFile", "", "Certificate authority file path to use to authenticate with Kubernetes.")
	newCommand.CobraCommand().PersistentFlags().StringVar(&f.Kubernetes.TLS.CrtFile, "service.kubernetes.tls.crtFile", "", "Certificate file path to use to authenticate with Kubernetes.")
	newCommand.CobraCommand().PersistentFlags().StringVar(&f.Kubernetes.TLS.KeyFile, "service.kubernetes.tls.keyFile", "", "Key file path to use to authenticate with Kubernetes.")
	newCommand.CobraCommand().PersistentFlags().StringVar(&f.Kubernetes.Pod.Name, "service.kubernetes.pod.name", "", "Name of the guest cluster kvm Kubernetes pod.")

	newCommand.cobraCommand.PersistentFlags().StringVar(&f.Provider.Bridge.Name, "provider.bridge.name", "", "Bridge name of the guest cluster VM on the host network.")
	newCommand.cobraCommand.PersistentFlags().StringVar(&f.Provider.Env.Prefix, "provider.env.prefix", "K8S_ENDPOINT_UPDATER_POD_", "Prefix of environment variables providing pod names.")
	newCommand.cobraCommand.PersistentFlags().StringVar(&f.Provider.Etcd.Address, "provider.etcd.address", "", "Address used to connect to etcd.")
	newCommand.cobraCommand.PersistentFlags().StringVar(&f.Provider.Etcd.Kind, "provider.etcd.kind", "etcdv2", "Etcd storage client version to use.")
	newCommand.cobraCommand.PersistentFlags().StringVar(&f.Provider.Etcd.Prefix, "provider.etcd.prefix", "", "Prefix of etcd paths providing pod names.")
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

func (c *Command) Execute(cmd *cobra.Command, args []string) {
	c.logger.Log("info", "start adding annotations to KVM pod")

	err := f.Validate()
	if err != nil {
		c.logger.Log("error", fmt.Sprintf("%#v", microerror.Mask(err)))
		os.Exit(1)
	}

	err = c.execute()
	if err != nil {
		c.logger.Log("error", fmt.Sprintf("%#v", microerror.Mask(err)))
		os.Exit(1)
	}

	c.logger.Log("info", "finished adding annotations to KVM pod")
}

func (c *Command) execute() error {
	var err error

	var k8sClient kubernetes.Interface
	{
		k8sConfig := k8sclient.DefaultConfig()

		k8sConfig.Address = f.Kubernetes.Address
		k8sConfig.Logger = c.logger
		k8sConfig.InCluster = f.Kubernetes.InCluster
		k8sConfig.TLS.CAFile = f.Kubernetes.TLS.CaFile
		k8sConfig.TLS.CrtFile = f.Kubernetes.TLS.CrtFile
		k8sConfig.TLS.KeyFile = f.Kubernetes.TLS.KeyFile

		k8sClient, err = k8sclient.New(k8sConfig)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var newProvider provider.Provider
	{
		bridgeConfig := bridge.DefaultConfig()

		bridgeConfig.Logger = c.logger

		bridgeConfig.BridgeName = f.Provider.Bridge.Name

		newProvider, err = bridge.New(bridgeConfig)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	// We need to create the updater which is able to update Kubernetes endpoints.
	var newUpdater *updater.Updater
	{
		updaterConfig := updater.DefaultConfig()

		updaterConfig.K8sClient = k8sClient
		updaterConfig.Logger = c.logger

		newUpdater, err = updater.New(updaterConfig)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	// Here we lookup the VM IP we are interested in.
	var podIP net.IP
	{
		action := func() error {
			podIP, err = newProvider.Lookup()

			if err != nil {
				return microerror.Mask(err)
			}

			return nil
		}

		err := backoff.Retry(action, backoff.NewExponentialBackOff())
		if err != nil {
			return microerror.Mask(err)
		}

		c.logger.Log("debug", fmt.Sprintf("found pod info of service '%s'", f.Kubernetes.Cluster.Service), "ip", podIP.String())

	}

	// Use the updater to actually add annotations to the kvm pod.
	{
		action := func() error {
			err := newUpdater.AddAnnotations(f.Kubernetes.Cluster.Namespace, f.Kubernetes.Cluster.Service, f.Kubernetes.Pod.Name, podIP)
			if err != nil {
				return microerror.Mask(err)
			}

			return nil
		}

		err := backoff.Retry(action, backoff.NewExponentialBackOff())
		if err != nil {
			return microerror.Mask(err)
		}

		c.logger.Log("debug", fmt.Sprintf("added annotations to the KVM pod '%s'", f.Kubernetes.Pod.Name))
	}

	// wait forever
	select {}
}
