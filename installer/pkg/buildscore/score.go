package buildscore

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/fromanirh/virtshift/installer/pkg/graphinfo"
)

const (
	releaseStreamBaseURL = "https://openshift-release.svc.ci.openshift.org/releasestream/"
)

type BuildScore map[string]int

func NewFromBuildInfo(bi []graphinfo.BuildInfo) (BuildScore, error) {
	var err error
	scores := make(map[string]int)
	for _, item := range bi {
		page, err := getReleasePage(item.Version)
		if err != nil {
			continue // TODO
		}
		scores[item.Version] = strings.Count(page, "Succeeded")
	}
	return scores, err
}

func getReleasePage(ver string) (string, error) {
	kind := kindOf(ver)
	if kind == "" {
		return "", fmt.Errorf("unable to extract version kind from %s", ver)
	}
	return getPage(fmt.Sprintf("%s/%s/release/%s", releaseStreamBaseURL, kind, ver))
}

func getPage(location string) (string, error) {
	resp, err := http.Get(location)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	return string(data), err
}

// kindOf("4.2.0-0.nightly-2019-11-28-230858") -> "4.2.0-0.nightly"
func kindOf(ver string) string {
	items := strings.Split(ver, "-")
	if len(items) >= 2 && hasKind(items[1]) {
		return fmt.Sprintf("%s-%s", items[0], items[1])
	}
	return ""
}

func hasKind(ver string) bool {
	return strings.Contains(ver, "ci") || strings.Contains(ver, "nightly")
}
