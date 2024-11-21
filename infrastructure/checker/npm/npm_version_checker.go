package npm

import (
	"context"
	"net/http"
	"path"
	"strings"

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

	info := ports.PackageInfo{
		Name:           registryResult.Name,
		LatestVersion:  registryResult.DistTags.Latest,
		CurrentVersion: packageUrl.Version,
		PackageManager: "npm",
	}

	if idx := strings.Index(registryResult.Name, "/"); idx >= 0 {
		info.Namespace = registryResult.Name[:idx]
		info.Name = registryResult.Name[idx:]
	}

	return &info, nil
}

type npmRegistryQueryResult struct {
	Name     string `json:"name"`
	DistTags struct {
		Latest string `json:"latest"`
	} `json:"dist-tags"`
}
