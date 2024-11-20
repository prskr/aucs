package pypi

import (
	"context"
	"net/http"
	"path"

	"github.com/carlmjohnson/requests"
	"github.com/package-url/packageurl-go"
	"github.com/prskr/aucs/core/ports"
)

var _ ports.UpdateChecker = (*Checker)(nil)

func NewChecker(client *http.Client) Checker {
	return Checker{Client: client}
}

type Checker struct {
	Client *http.Client
}

// LatestVersionFor implements ports.UpdateChecker.
func (c Checker) LatestVersionFor(ctx context.Context, packageUrl packageurl.PackageURL) (*ports.PackageInfo, error) {
	var pypiResult pypiQueryResult

	err := requests.
		URL("https://pypi.python.org").
		Path(path.Join("pypi", packageUrl.Name, "json")).
		Client(c.Client).
		ToJSON(&pypiResult).
		Fetch(ctx)
	if err != nil {
		return nil, err
	}

	return &ports.PackageInfo{
		Name:           pypiResult.Info.Name,
		CurrentVersion: packageUrl.Version,
		LatestVersion:  pypiResult.Info.Version,
		PackageManager: "pypi",
	}, nil
}

// SupportedPackageType implements ports.UpdateChecker.
func (Checker) SupportedPackageType() string {
	return "pypi"
}

type pypiQueryResult struct {
	Info pypiPackageInfo `json:"info"`
}

type pypiPackageInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}
