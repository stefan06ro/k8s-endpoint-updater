// Package update implements the update command for the command line tool.
package update

import (
	"fmt"
	"net/url"
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/cenk/backoff"
	microerror "github.com/giantswarm/microkit/error"
	micrologger "github.com/giantswarm/microkit/logger"
	microstorage "github.com/giantswarm/microkit/storage"
	"github.com/spf13/cobra"

	"github.com/giantswarm/k8s-endpoint-updater/command/update/flag"
	"github.com/giantswarm/k8s-endpoint-updater/service/provider"
	"github.com/giantswarm/k8s-endpoint-updater/service/provider/bridge"
	"github.com/giantswarm/k8s-endpoint-updater/service/provider/env"
	"github.com/giantswarm/k8s-endpoint-updater/service/provider/etcd"
	"github.com/giantswarm/k8s-endpoint-updater/service/updater"
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

	newCommand.CobraCommand().PersistentFlags().StringVar(&f.Kubernetes.Address, "service.kubernetes.address", "http://127.0.0.1:6443", "Address used to connect to Kubernetes. When empty in-cluster config is created.")
	newCommand.CobraCommand().PersistentFlags().StringVar(&f.Kubernetes.Cluster.Namespace, "service.kubernetes.cluster.namespace", "default", "Namespace of the guest cluster which endpoints should be updated.")
	newCommand.CobraCommand().PersistentFlags().BoolVar(&f.Kubernetes.InCluster, "service.kubernetes.inCluster", false, "Whether to use the in-cluster config to authenticate with Kubernetes.")
	newCommand.CobraCommand().PersistentFlags().StringVar(&f.Kubernetes.TLS.CaFile, "service.kubernetes.tls.caFile", "", "Certificate authority file path to use to authenticate with Kubernetes.")
	newCommand.CobraCommand().PersistentFlags().StringVar(&f.Kubernetes.TLS.CrtFile, "service.kubernetes.tls.crtFile", "", "Certificate file path to use to authenticate with Kubernetes.")
	newCommand.CobraCommand().PersistentFlags().StringVar(&f.Kubernetes.TLS.KeyFile, "service.kubernetes.tls.keyFile", "", "Key file path to use to authenticate with Kubernetes.")

	newCommand.cobraCommand.PersistentFlags().StringVar(&f.Provider.Bridge.Name, "provider.bridge.name", "", "Bridge name of the guest cluster VM on the host network.")
	newCommand.cobraCommand.PersistentFlags().StringVar(&f.Provider.Env.Prefix, "provider.env.prefix", "K8S_ENDPOINT_UPDATER_POD_", "Prefix of environment variables providing pod names.")
	newCommand.cobraCommand.PersistentFlags().StringVar(&f.Provider.Etcd.Address, "provider.etcd.address", "", "Address used to connect to etcd.")
	newCommand.cobraCommand.PersistentFlags().StringVar(&f.Provider.Etcd.Kind, "provider.etcd.kind", "etcdv2", "Etcd storage client version to use.")
	newCommand.cobraCommand.PersistentFlags().StringVar(&f.Provider.Etcd.Prefix, "provider.etcd.prefix", "", "Prefix of etcd paths providing pod names.")
	newCommand.cobraCommand.PersistentFlags().StringVar(&f.Provider.Kind, "provider.kind", "env", "Provider used to lookup pod IPs.")

	newCommand.cobraCommand.PersistentFlags().StringSliceVar(&f.Updater.Pod.Names, "updater.pod.names", nil, "List of pod names used to lookup pod IPs.")

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

	// Sleep forver to make Kubernetes happy. It does not like terminated pods.
	select {}
}

func (c *Command) execute() error {
	var err error

	// At first we have to sort out which provider to use. This is based on the
	// flags given to the updater.
	var newProvider provider.Provider
	{
		k := f.Provider.Kind
		switch k {
		case bridge.Kind:
			if len(f.Updater.Pod.Names) != 1 {
				return microerror.MaskAnyf(invalidConfigError, "bridge provider expects 1 pod name")
			}
			podName := f.Updater.Pod.Names[0]

			bridgeConfig := bridge.DefaultConfig()
			bridgeConfig.BridgeName = f.Provider.Bridge.Name
			bridgeConfig.Logger = c.logger
			bridgeConfig.PodName = podName
			newProvider, err = bridge.New(bridgeConfig)
			if err != nil {
				return microerror.MaskAny(err)
			}
		case env.Kind:
			envConfig := env.DefaultConfig()
			envConfig.Logger = c.logger
			envConfig.PodNames = f.Updater.Pod.Names
			envConfig.Prefix = f.Provider.Env.Prefix
			newProvider, err = env.New(envConfig)
			if err != nil {
				return microerror.MaskAny(err)
			}
		case etcd.Kind:
			var storageService microstorage.Service
			{
				storageConfig := microstorage.DefaultConfig()
				storageConfig.EtcdAddress = f.Provider.Etcd.Address
				storageConfig.EtcdPrefix = f.Provider.Etcd.Prefix
				storageConfig.Kind = f.Provider.Etcd.Kind
				storageService, err = microstorage.New(storageConfig)
				if err != nil {
					return microerror.MaskAny(err)
				}
			}

			etcdConfig := etcd.DefaultConfig()
			etcdConfig.Logger = c.logger
			etcdConfig.PodNames = f.Updater.Pod.Names
			etcdConfig.Storage = storageService
			newProvider, err = etcd.New(etcdConfig)
			if err != nil {
				return microerror.MaskAny(err)
			}
		default:
			return microerror.MaskAnyf(invalidConfigError, "unsupported provider kind '%s'", k)
		}
	}

	// We also need to create the updater which is able to update Kubernetes
	// endpoints.
	var newUpdater *updater.Updater
	{
		var kubernetesClient *kubernetes.Clientset
		{
			var restConfig *rest.Config

			if f.Kubernetes.InCluster {
				c.logger.Log("debug", "creating in-cluster config")
				restConfig, err = rest.InClusterConfig()
				if err != nil {
					return microerror.MaskAny(err)
				}

				if f.Kubernetes.Address != "" {
					c.logger.Log("debug", "using explicit api server")
					restConfig.Host = f.Kubernetes.Address
				}
			} else {
				if f.Kubernetes.Address == "" {
					return microerror.MaskAnyf(invalidConfigError, "kubernetes address must not be empty")
				}

				c.logger.Log("debug", "creating out-cluster config")

				// Kubernetes listen URL.
				u, err := url.Parse(f.Kubernetes.Address)
				if err != nil {
					return microerror.MaskAny(err)
				}

				restConfig = &rest.Config{
					Host: u.String(),
					TLSClientConfig: rest.TLSClientConfig{
						CAFile:   f.Kubernetes.TLS.CaFile,
						CertFile: f.Kubernetes.TLS.CrtFile,
						KeyFile:  f.Kubernetes.TLS.KeyFile,
					},
				}
			}

			kubernetesClient, err = kubernetes.NewForConfig(restConfig)
			if err != nil {
				return microerror.MaskAny(err)
			}
		}

		updaterConfig := updater.DefaultConfig()
		updaterConfig.KubernetesClient = kubernetesClient
		updaterConfig.Logger = c.logger
		newUpdater, err = updater.New(updaterConfig)
		if err != nil {
			return microerror.MaskAny(err)
		}
	}

	// Once we know which provider to use we execute it to lookup the pod
	// information we are interested in.
	var podInfos []provider.PodInfo
	{
		action := func() error {
			podInfos, err = newProvider.Lookup()
			if err != nil {
				return microerror.MaskAny(err)
			}

			return nil
		}

		err := backoff.Retry(action, backoff.NewExponentialBackOff())
		if err != nil {
			return microerror.MaskAny(err)
		}

		for _, pi := range podInfos {
			c.logger.Log("debug", "found pod info", "IP", pi.IP.String(), "pod", pi.Name)
		}
	}

	// Use the updater to actually update the endpoints identified by the provided
	// flags.
	{
		err = newUpdater.Update(f.Kubernetes.Cluster.Namespace, podInfos)
		if err != nil {
			return microerror.MaskAny(err)
		}
	}

	return nil
}
