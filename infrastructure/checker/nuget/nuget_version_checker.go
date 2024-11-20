package nuget

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/carlmjohnson/requests"
	"github.com/package-url/packageurl-go"
	"github.com/prskr/aucs/core/ports"
)

var _ ports.UpdateChecker = (*Checker)(nil)

func NewNugetChecker(client *http.Client) *Checker {
	return &Checker{Client: client}
}

type Checker struct {
	Client *http.Client
}

// SupportedPackageType implements ports.UpdateChecker.
func (c Checker) SupportedPackageType() string {
	return "nuget"
}

// LatestVersionFor implements ports.UpdateChecker.
func (c Checker) LatestVersionFor(ctx context.Context, packageUrl packageurl.PackageURL) (*ports.PackageInfo, error) {
	var nugetResult NugetQueryResult

	err := requests.
		URL("https://azuresearch-usnc.nuget.org/query").
		Param("q", fmt.Sprintf("packageid:%s", packageUrl.Name)).
		Client(c.Client).
		ToJSON(&nugetResult).
		Fetch(ctx)

	if err != nil {
		return nil, err
	}

	if nugetResult.TotalHits == 0 {
		return nil, fmt.Errorf("%w: %s", ports.ErrNoMatchingPackageFound, packageUrl.Name)
	}

	if nugetResult.TotalHits > 1 {
		return nil, fmt.Errorf("%w: %s", ports.ErrAmbiguousPackageFound, packageUrl.Name)
	}

	currentVersion, err := parseNugetVersion(packageUrl.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to parse current version - %s: %w", packageUrl.Version, err)
	}
	latestVersion, err := parseNugetVersion(nugetResult.Packages[0].Version)
	if err != nil {
		return nil, fmt.Errorf("failed to parse latest version - %s: %w", nugetResult.Packages[0].Version, err)
	}

	if currentVersion.GreaterThan(latestVersion) {
		return nil, fmt.Errorf("%w: %s", ports.ErrCurrentVersionGreaterThanLatest, nugetResult.Packages[0].Version)
	}

	return &ports.PackageInfo{
		Name:           nugetResult.Packages[0].Id,
		LatestVersion:  nugetResult.Packages[0].Version,
		CurrentVersion: packageUrl.Version,
		PackageManager: "nuget",
	}, nil
}

type NugetQueryResult struct {
	TotalHits int                `json:"totalHits"`
	Packages  []NugetPackageInfo `json:"data"`
}

type NugetPackageInfo struct {
	Id      string `json:"id"`
	Version string `json:"version"`
	Title   string `json:"title"`
}

func parseNugetVersion(version string) (*semver.Version, error) {
	if strings.Count(version, ".") > 2 {
		version = strings.Join(strings.Split(version, ".")[:3], ".")
	}

	return semver.NewVersion(version)
}
