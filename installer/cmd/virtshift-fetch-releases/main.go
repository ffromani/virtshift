package main

import (
	"fmt"
	"os"
	"text/tabwriter"
	"text/template"

	"github.com/spf13/pflag"

	"github.com/fromanirh/virtshift/installer/pkg/buildscore"
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
	var skipBuildScore bool
	var bestScore bool

	pflag.Usage = showHelp
	pflag.StringVarP(&graphURL, "url", "u", defaultGraphURL, "URL from where to get the graph data")
	pflag.StringVarP(&OCPVersion, "ocpversion", "V", defaultOCPVersion, "OCP version to be considered")
	pflag.BoolVarP(&skipBuildScore, "score", "S", false, "do not compute the build score")
	pflag.BoolVarP(&bestScore, "best", "B", false, "show only data about the best scored build")
	pflag.Parse()

	if bestScore && skipBuildScore {
		fmt.Fprintf(os.Stderr, "no scores available\n")
		os.Exit(0)
	}

	var err error
	bi, err := graphinfo.NewFromURL(graphURL, OCPVersion)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	var scores buildscore.BuildScore
	if !skipBuildScore {
		scores, err = buildscore.NewFromBuildInfo(bi)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(2)
		}
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 2, 1, ' ', 0)
	if bestScore {
		item := findHighestScoreBuildInfo(bi, scores)
		fmt.Fprintf(w, "%s\t\t%s\t\t%v\n", item.Version, item.Payload, scores[item.Version])
	} else {
		for _, item := range bi {
			fmt.Fprintf(w, "%s\t\t%s\t\t%v\n", item.Version, item.Payload, scores[item.Version])
		}
	}
	w.Flush()
}

func findHighestScoreBuildInfo(bi []graphinfo.BuildInfo, scores buildscore.BuildScore) graphinfo.BuildInfo {
	highScoreVal := 0
	var res graphinfo.BuildInfo
	for _, item := range bi {
		if scores[item.Version] > highScoreVal {
			highScoreVal = scores[item.Version]
			res = item
		}

	}
	return res
}
