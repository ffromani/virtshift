package main

import (
	"fmt"
	"os"
	"text/tabwriter"
	"text/template"

	"github.com/spf13/pflag"

	"github.com/fromanirh/virtshift/installer/pkg/graphinfo"
)

const (
	helpMessage = `{{ .Executable }} fetches the last builds from OCP CI.

Usage: {{ .Executable }} [flags]

Environment variables:
VIRTSHIFT_DEBUG: set this variable (any value) to enable debug messages (disabled)

`
	// ref: https://github.com/openshift/cincinnati
	defaultGraphURL   = "https://openshift-release.svc.ci.openshift.org/graph"
	defaultOCPVersion = "4.3"
)

func showHelp() {
	type Help struct {
		Executable string
	}
	help := Help{
		Executable: "virtshift-fetch-releases",
	}
	tmpl := template.Must(template.New("help").Parse(helpMessage))
	err := tmpl.Execute(os.Stderr, help)
	if err != nil {
		panic(err)
	}
	pflag.PrintDefaults()
}

func main() {
	var graphURL string
	var OCPVersion string

	pflag.Usage = showHelp
	pflag.StringVarP(&graphURL, "url", "u", defaultGraphURL, "URL from where to get the graph data")
	pflag.StringVarP(&OCPVersion, "ocpversion", "V", defaultOCPVersion, "OCP version to be considered")
	pflag.Parse()

	bi, err := graphinfo.NewFromURL(graphURL, OCPVersion)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 2, 1, ' ', 0)
	for _, item := range bi {
		fmt.Fprintf(w, "%s\t\t%s\n", item.Version, item.Payload)
	}
	w.Flush()
}
