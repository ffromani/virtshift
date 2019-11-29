package checkpoint

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/fromanirh/virtshift/installer/pkg/debug"
)

const (
	CheckpointsURL string = "https://raw.githubusercontent.com/fromanirh/virtshift/wrapinstall/installer/checkpoints.json"
)

type OpenshiftInstallVersion struct {
	Version     string `json:"version"`
	BuildCommit string `json:"commit"`
}

type Checkpoint struct {
	Installer       OpenshiftInstallVersion `json:"installer"`
	ReleaseImageURL string                  `json:"release_image"`
}

func (cp Checkpoint) IsValid() bool {
	/* everything else is optional atm */
	return cp.ReleaseImageURL != ""
}

func NewCheckpointFromExe(path string) (Checkpoint, error) {
	var info Checkpoint
	cmd := exec.Command(path, "version")
	out, err := cmd.Output()
	if err != nil {
		return info, err
	}
	rd := bufio.NewReader(bytes.NewBuffer(out))
	err = parseVersionOutput(rd, &info)
	return info, err
}

func LastFromURL(location string) (Checkpoint, error) {
	u, err := url.Parse(location)
	if err != nil {
		return Checkpoint{}, err
	}
	if u.Scheme == "" {
		// assume local path
		return lastFromFile(u.Path)
	}
	return lastFromHTTP(location)
}

func lastFromFile(path string) (Checkpoint, error) {
	debug.Printf("reading checkpoints from file: %s", path)
	fd, err := os.Open(path)
	if err != nil {
		return Checkpoint{}, err
	}
	defer fd.Close()
	return parseCheckpointsFile(fd)
}

func lastFromHTTP(location string) (Checkpoint, error) {
	debug.Printf("reading checkpoints from http: %s", location)
	resp, err := http.Get(location)
	if err != nil {
		return Checkpoint{}, err
	}
	defer resp.Body.Close()
	return parseCheckpointsFile(resp.Body)
}

func parseCheckpointsFile(r io.Reader) (Checkpoint, error) {
	debug.Printf("parsing checkpoints")
	cpts := []Checkpoint{}
	dec := json.NewDecoder(r)
	err := dec.Decode(&cpts)
	if err != nil {
		return Checkpoint{}, err
	}
	return pickLast(cpts)
}

func pickLast(cpts []Checkpoint) (Checkpoint, error) {
	return cpts[0], nil
}

/*
 * example:
 * /some/path/to/openshift-install unreleased-master-2147-g1d7ed7af26a804b229924633b22a3ea013cf9cae
 * built from commit 1d7ed7af26a804b229924633b22a3ea013cf9cae
 * release image registry.svc.ci.openshift.org/ocp/release:4.3
 */
func parseVersionOutput(rd *bufio.Reader, info *Checkpoint) error {
	var line string
	var data []byte
	var err error

	data, _, err = rd.ReadLine()
	if err != nil {
		return err
	}
	line = string(data)
	ErrVersionLine := fmt.Errorf("malformed version line: %s", line)
	items := strings.Split(line, " ")
	if len(items) != 2 || path.Base(items[0]) != "openshift-install" {
		return ErrVersionLine
	}
	info.Installer.Version = strings.TrimSpace(items[1])

	data, _, err = rd.ReadLine()
	if err != nil {
		return err
	}
	line = string(data)
	if !strings.HasPrefix(line, "built from commit") {
		return fmt.Errorf("malformed commit line: %s", line)
	}
	info.Installer.BuildCommit = strings.TrimSpace(strings.TrimPrefix(line, "built from commit"))

	data, _, err = rd.ReadLine()
	if err != nil {
		return err
	}
	line = string(data)
	if !strings.HasPrefix(line, "release image") {
		return fmt.Errorf("malformed image line: %s", line)
	}
	info.ReleaseImageURL = strings.TrimSpace(strings.TrimPrefix(line, "release image"))

	return nil
}
