package env

import (
	"net"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"

	microerror "github.com/giantswarm/microkit/error"
	micrologger "github.com/giantswarm/microkit/logger"

	"github.com/giantswarm/k8s-endpoint-updater/service/provider"
)

const (
	Kind = "env"
)

// Config represents the configuration used to create a new provider.
type Config struct {
	// Dependencies.
	Logger micrologger.Logger

	// Settings.
	PodNames []string
	Prefix   string
}

// DefaultConfig provides a default configuration to create a new provider
// by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		Logger: nil,

		// Settings.
		PodNames: nil,
		Prefix:   "",
	}
}

// New creates a new provider.
func New(config Config) (*Provider, error) {
	// Dependencies.
	if config.Logger == nil {
		return nil, microerror.MaskAnyf(invalidConfigError, "logger must not be empty")
	}

	// Settings.
	if len(config.PodNames) == 0 {
		return nil, microerror.MaskAnyf(invalidConfigError, "pod names must not be empty")
	}
	if config.Prefix == "" {
		return nil, microerror.MaskAnyf(invalidConfigError, "prefix must not be empty")
	}

	newProvider := &Provider{
		// Dependencies.
		logger: config.Logger,

		// Settings.
		podNames: config.PodNames,
		prefix:   config.Prefix,
	}

	return newProvider, nil
}

type Provider struct {
	// Dependencies.
	logger micrologger.Logger

	// Settings.
	podNames []string
	prefix   string
}

func (p *Provider) Lookup() ([]provider.PodInfo, error) {
	var podInfos []provider.PodInfo

	for i, ev := range podNamesToEnvVars(p.podNames, p.prefix) {
		podInfo := provider.PodInfo{
			IP:   net.ParseIP(os.Getenv(ev)),
			Name: p.podNames[i],
		}

		podInfos = append(podInfos, podInfo)
	}

	return podInfos, nil
}

func filterParts(list []string) []string {
	var newList []string

	for _, v := range list {
		if v == "-" || v == "_" || strings.TrimSpace(v) == "" {
			continue
		}

		newList = append(newList, v)
	}

	return newList
}

func partsToEnvVar(list []string) string {
	var newList []string
	envVar := strings.Join(list, "_")

	for _, c := range envVar {
		newList = append(newList, string(unicode.ToUpper(c)))
	}

	newEnvVar := strings.Join(newList, "")

	return newEnvVar
}

func podNamesToEnvVars(podNames []string, prefix string) []string {
	var envVars []string

	for _, pn := range podNames {
		parts := filterParts(splitPodName(pn))
		envVar := partsToEnvVar(parts)

		envVars = append(envVars, prefix+envVar)
	}

	return envVars
}

// splitPodName is shamelessly copied from https://github.com/fatih/camelcase.
// There is only one single modification to make our own code work. We do not
// split by numbers.
func splitPodName(src string) (entries []string) {
	// don't split invalid utf8
	if !utf8.ValidString(src) {
		return []string{src}
	}
	entries = []string{}
	var runes [][]rune
	lastClass := 0
	class := 0
	// split into fields based on class of unicode character
	for _, r := range src {
		switch true {
		case unicode.IsLower(r):
			class = 1
		case unicode.IsUpper(r):
			class = 2
		case unicode.IsDigit(r):
			// Note the original code used the following. We want to treat numbers
			// like usual lower case characters.
			//
			//     class = 3
			//
			class = 1
		default:
			class = 4
		}
		if class == lastClass {
			runes[len(runes)-1] = append(runes[len(runes)-1], r)
		} else {
			runes = append(runes, []rune{r})
		}
		lastClass = class
	}
	// handle upper case -> lower case sequences, e.g.
	// "PDFL", "oader" -> "PDF", "Loader"
	for i := 0; i < len(runes)-1; i++ {
		if unicode.IsUpper(runes[i][0]) && unicode.IsLower(runes[i+1][0]) {
			runes[i+1] = append([]rune{runes[i][len(runes[i])-1]}, runes[i+1]...)
			runes[i] = runes[i][:len(runes[i])-1]
		}
	}
	// construct []string from results
	for _, s := range runes {
		if len(s) > 0 {
			entries = append(entries, string(s))
		}
	}
	return
}
