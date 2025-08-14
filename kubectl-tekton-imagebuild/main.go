package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kraema/kubectl-plugins/kubectl-tekton-imagebuild/pkg/pipelinerun"
	"github.com/spf13/pflag"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

const (
	labelKey = "imagebuild.ba.de/imagebuildschedule"
)

func main() {
	var (
		kubeconfig    string
		namespace     string
		allNamespaces bool
		scheduleValue string
		outputFormat  string
		help          bool
		noHeaders     bool
	)

	flags := pflag.NewFlagSet("kubectl-tekton-imagebuild", pflag.ExitOnError)

	if home := homedir.HomeDir(); home != "" {
		flags.StringVar(&kubeconfig, "kubeconfig", filepath.Join(home, ".kube", "config"), "path to the kubeconfig file")
	} else {
		flags.StringVar(&kubeconfig, "kubeconfig", "", "path to the kubeconfig file")
	}

	flags.StringVarP(&namespace, "namespace", "n", "", "namespace to list PipelineRuns from (default: current context namespace)")
	flags.BoolVarP(&allNamespaces, "all-namespaces", "A", false, "list PipelineRuns from all namespaces")
	flags.StringVarP(&scheduleValue, "schedule", "s", "", "filter by specific schedule value")
	flags.StringVarP(&outputFormat, "output", "o", "table", "output format (table, wide, json, yaml)")
	flags.BoolVar(&noHeaders, "no-headers", false, "don't print headers (only for table output)")
	flags.BoolVarP(&help, "help", "h", false, "show help message")

	flags.Usage = func() {
		fmt.Fprintf(os.Stderr, `kubectl-tekton-imagebuild - List Tekton PipelineRuns with imagebuild schedule labels

Usage:
  kubectl tekton-imagebuild [flags]

Examples:
  # List all PipelineRuns with any imagebuild schedule label in current namespace
  kubectl tekton-imagebuild

  # List PipelineRuns with specific schedule value
  kubectl tekton-imagebuild --schedule=daily

  # List PipelineRuns from all namespaces
  kubectl tekton-imagebuild -A

  # List PipelineRuns in specific namespace
  kubectl tekton-imagebuild -n production

  # Output in JSON format
  kubectl tekton-imagebuild -o json

Flags:
`)
		flags.PrintDefaults()
	}

	if err := flags.Parse(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	if help {
		flags.Usage()
		os.Exit(0)
	}

	// Build the config from kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building kubeconfig: %v\n", err)
		os.Exit(1)
	}

	// If namespace is not specified and not all-namespaces, get from current context
	if namespace == "" && !allNamespaces {
		loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
		configOverrides := &clientcmd.ConfigOverrides{}
		kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
		namespace, _, err = kubeConfig.Namespace()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting namespace from context: %v\n", err)
			os.Exit(1)
		}
	}

	// Create the lister
	lister, err := pipelinerun.NewLister(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating lister: %v\n", err)
		os.Exit(1)
	}

	// List PipelineRuns
	options := pipelinerun.ListOptions{
		Namespace:     namespace,
		AllNamespaces: allNamespaces,
		LabelKey:      labelKey,
		LabelValue:    scheduleValue,
		OutputFormat:  outputFormat,
		NoHeaders:     noHeaders,
	}

	if err := lister.List(options); err != nil {
		fmt.Fprintf(os.Stderr, "Error listing PipelineRuns: %v\n", err)
		os.Exit(1)
	}
}
