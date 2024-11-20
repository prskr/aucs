package ports

import (
	"context"
	"errors"

	"github.com/package-url/packageurl-go"
)

var (
	ErrNoCheckerForPackageType         = errors.New("no checker for package type")
	ErrNoMatchingPackageFound          = errors.New("no matching package found")
	ErrAmbiguousPackageFound           = errors.New("ambiguous package found")
	ErrCurrentVersionGreaterThanLatest = errors.New("current version is greater than latest version")
)

type PackageInfo struct {
	Name           string
	LatestVersion  string
	CurrentVersion string
	PackageManager string
}

type UpdateChecker interface {
	LatestVersionFor(ctx context.Context, packageUrl packageurl.PackageURL) (*PackageInfo, error)
	SupportedPackageType() string
}
