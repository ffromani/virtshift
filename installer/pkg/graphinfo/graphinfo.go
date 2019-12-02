package graphinfo

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/fromanirh/virtshift/installer/pkg/debug"
)

type BuildInfo struct {
	Version string `json:"version"`
	Payload string `json:"payload"`
}

type Edge [2]int

type Graph struct {
	Edges []Edge      `json:"edges"`
	Nodes []BuildInfo `json:"nodes"`
}

func NewFromURL(location, version string) ([]BuildInfo, error) {
	u, err := url.Parse(location)
	if err != nil {
		return nil, err
	}
	var allInfos []BuildInfo
	if u.Scheme == "" {
		// assume local path
		allInfos, err = newFromFile(u.Path)
	} else {
		allInfos, err = newFromHTTP(location)
	}
	if err != nil {
		return nil, err
	}
	versionInfos := takeByVersion(allInfos, version)
	// TODO: sort
	return versionInfos, nil
}

func newFromFile(path string) ([]BuildInfo, error) {
	debug.Printf("reading buildinfos from file: %s", path)
	fd, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fd.Close()
	return extractBuildInfoFromGraphJSON(fd)
}

func newFromHTTP(location string) ([]BuildInfo, error) {
	debug.Printf("reading buildinfos from http: %s", location)
	resp, err := http.Get(location)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return extractBuildInfoFromGraphJSON(resp.Body)
}

func extractBuildInfoFromGraphJSON(r io.Reader) ([]BuildInfo, error) {
	debug.Printf("parsing graphinfos")
	g := Graph{}
	dec := json.NewDecoder(r)
	err := dec.Decode(&g)
	if err != nil {
		return nil, err
	}
	return g.Nodes, nil
}

func takeByVersion(bi []BuildInfo, version string) []BuildInfo {
	res := []BuildInfo{}
	for _, item := range bi {
		if strings.HasPrefix(item.Version, version) {
			res = append(res, item)
		}
	}
	return res
}
