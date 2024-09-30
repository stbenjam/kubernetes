package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/openshift-eng/openshift-tests-extension/pkg/cmd"
	e "github.com/openshift-eng/openshift-tests-extension/pkg/extension"
	g "github.com/openshift-eng/openshift-tests-extension/pkg/ginkgo"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	utilflag "k8s.io/component-base/cli/flag"
	"k8s.io/component-base/logs"
	// initialize framework extensions
	_ "k8s.io/kubernetes/test/e2e/framework/debug/init"
	_ "k8s.io/kubernetes/test/e2e/framework/metrics/init"
)

func main() {
	logs.InitLogs()
	defer logs.FlushLogs()
	rand.Seed(time.Now().UTC().UnixNano())
	pflag.CommandLine.SetNormalizeFunc(utilflag.WordSepNormalizeFunc)

	// Initialize test framework
	if err := initializeTestFramework(os.Getenv("TEST_PROVIDER")); err != nil {
		panic(err)
	}

	// Create openshift-tests extension
	registry := e.NewRegistry()
	ext := e.NewExtension("openshift", "payload", "hyperkube")
	ext.AddSuite(e.Suite{Name: "k8s", Parents: []string{"openshift/conformance/parallel"}})

	// Cobra stuff
	root := &cobra.Command{
		Long: "OpenShift Tests compatible wrapper",
	}

	root.AddCommand(
		cmd.DefaultExtensionCommands(registry)...,
	)

	specs, err := g.BuildExtensionTestSpecsFromOpenShiftGinkgoSuite()
	if err != nil {
		panic(err)
	}

	ext.AddSpecs(specs)
	registry.Register(ext)

	if err := func() error {
		return root.Execute()
	}(); err != nil {
		if ex, ok := err.(ExitError); ok {
			fmt.Fprintf(os.Stderr, "Ginkgo exit error %d: %v\n", ex.Code, err)
			os.Exit(ex.Code)
		}
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
