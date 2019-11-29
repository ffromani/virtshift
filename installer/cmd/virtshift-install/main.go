package main

import (
	"fmt"
	"os"
	"os/exec"
	"text/template"

	"github.com/fromanirh/virtshift/installer/pkg/checkpoint"
	"github.com/fromanirh/virtshift/installer/pkg/debug"
)

const helpMessage string = `{{ .Executable }} wraps openshift-installer to create a known working cluster.

use this tool IFF you want to run a OCP cluster from CI using the libvirt provider.
{{ .Executable }} is configured using environment variables only.
Do not set them if unsuere, virtshift-install has sane defaults.

VIRTSHIFT_CHECKPOINTS_URL: URL to checkpoints.json ({{ .CheckpointsURL }})
VIRTSHIFT_DEBUG:           set this variable (any value) to enable debug messages (disabled)
`

func showHelp() {
	type Help struct {
		Executable     string
		CheckpointsURL string
	}
	help := Help{
		Executable:     "virtshift-install",
		CheckpointsURL: checkpoint.CheckpointsURL,
	}
	tmpl := template.Must(template.New("help").Parse(helpMessage))
	err := tmpl.Execute(os.Stderr, help)
	if err != nil {
		panic(err)
	}
}

func main() {
	if len(os.Args) >= 2 && os.Args[1] == "help" {
		showHelp()
	}
	debug.Printf("virtshift-install start")
	defer debug.Printf("virtshift-install done")

	var err error

	path, err := exec.LookPath("openshift-install")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	debug.Printf("openshift-install found at %s", path)

	/*installerInfo*/
	_, err = checkpoint.NewCheckpointFromExe(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(2)
	}

	jsonLocation := checkpoint.CheckpointsURL
	if val, ok := os.LookupEnv("VIRTSHIFT_CHECKPOINTS_URL"); ok {
		jsonLocation = val
	}
	cp, err := checkpoint.LastFromURL(jsonLocation)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(4)
	}

	if !cp.IsValid() {
		fmt.Fprintf(os.Stderr, "invalid checkpoint: %v\n", cp)
		os.Exit(8)
	}
	overrideVar := fmt.Sprintf("OPENSHIFT_INSTALL_RELEASE_IMAGE_OVERRIDE=%s", cp.ReleaseImageURL)

	debug.Printf("running openshift-install with %s", overrideVar)

	cmd := exec.Command(path, os.Args[1:]...)
	cmd.Env = append(os.Environ(), overrideVar)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		if x, ok := err.(*exec.ExitError); ok {
			os.Exit(x.ExitCode())
		}
		debug.Printf("%v not ExitErro but %T?!", err, err)
		os.Exit(255)
	}
}
