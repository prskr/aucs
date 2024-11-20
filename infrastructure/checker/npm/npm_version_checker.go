package npm

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

// SupportedPackageType implements ports.UpdateChecker.
func (Checker) SupportedPackageType() string {
	return "npm"
}

// LatestVersionFor implements ports.UpdateChecker.
func (c Checker) LatestVersionFor(ctx context.Context, packageUrl packageurl.PackageURL) (*ports.PackageInfo, error) {
	var registryResult npmRegistryQueryResult

	err := requests.
		URL("https://registry.npmjs.org").
		Path(path.Join(packageUrl.Namespace, packageUrl.Name)).
		Client(c.Client).
		ToJSON(&registryResult).
		Fetch(ctx)

	if err != nil {
		return nil, err
	}

	return &ports.PackageInfo{
		Name:           registryResult.Name,
		LatestVersion:  registryResult.DistTags.Latest,
		CurrentVersion: packageUrl.Version,
		PackageManager: "npm",
	}, nil
}

type npmRegistryQueryResult struct {
	Name     string `json:"name"`
	DistTags struct {
		Latest string `json:"latest"`
	} `json:"dist-tags"`
}
