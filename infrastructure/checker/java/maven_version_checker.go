package java

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"path"
	"slices"
	"strings"

	"github.com/Masterminds/semver/v3"
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
	var metadataResult mavenMetadata

	requestPath := path.Join(
		"maven2",
		strings.ReplaceAll(packageUrl.Namespace, ".", "/"),
		packageUrl.Name,
		"maven-metadata.xml",
	)

	err := requests.
		URL("https://repo.maven.apache.org/").
		Path(requestPath).
		Client(c.Client).
		ToDeserializer(xml.Unmarshal, &metadataResult).
		Fetch(ctx)
	if err != nil {
		return nil, err
	}

	latestVersion, err := metadataResult.latestVersion()
	if err != nil {
		return nil, err
	}

	return &ports.PackageInfo{
		Name:           fmt.Sprintf("%s:%s", metadataResult.GroupID, metadataResult.ArtifactID),
		CurrentVersion: packageUrl.Version,
		LatestVersion:  latestVersion,
	}, nil
}

// SupportedPackageType implements ports.UpdateChecker.
func (Checker) SupportedPackageType() string {
	return "maven"
}

type mavenMetadata struct {
	XMLName    xml.Name `xml:"metadata"`
	GroupID    string   `xml:"groupId"`
	ArtifactID string   `xml:"artifactId"`
	Versions   []string `xml:"versioning>versions>version"`
}

func (m mavenMetadata) latestVersion() (string, error) {
	versions := make([]*semver.Version, 0, len(m.Versions))
	for _, v := range m.Versions {
		if strings.Count(v, ".") > 2 {
			v = strings.Join(strings.Split(v, ".")[:3], ".")
		}

		parsed, err := semver.NewVersion(v)
		if err != nil {
			return "nil", fmt.Errorf("failed to parse version %s - %w", v, err)
		}

		versions = append(versions, parsed)
	}

	latestVersion := slices.MaxFunc(versions, func(a, b *semver.Version) int {
		return a.Compare(b)
	})

	return latestVersion.String(), nil
}
