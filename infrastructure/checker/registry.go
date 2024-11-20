package checker

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/package-url/packageurl-go"
	"github.com/prskr/aucs/core/ports"
)

func NewRegistry(kv ports.KeyValueStore) *Registry {
	return &Registry{
		KV:             kv,
		CheckersByType: make(map[string]ports.UpdateChecker),
	}
}

type Registry struct {
	KV             ports.KeyValueStore
	CheckersByType map[string]ports.UpdateChecker
}

func (r *Registry) Register(checkers ...ports.UpdateChecker) {
	for _, c := range checkers {
		r.CheckersByType[c.SupportedPackageType()] = c
	}
}

func (r Registry) LatestVersionFor(ctx context.Context, packageUrl string) (*ports.PackageInfo, error) {
	purl, err := packageurl.FromString(packageUrl)
	if err != nil {
		return nil, err
	}

	var (
		cacheKey = cacheKeyFor(purl)
		info     = new(ports.PackageInfo)
	)

	rawInfo, err := r.KV.Get(ctx, cacheKey)
	if err != nil {
		if !errors.Is(err, ports.ErrNoKVEntryForKey) {
			return nil, err
		}
	}

	if rawInfo != nil {
		return info, json.Unmarshal(rawInfo, info)
	}

	// Get the checker for the package type
	checker, ok := r.CheckersByType[purl.Type]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ports.ErrNoCheckerForPackageType, purl.Type)
	}

	// Delegate the call to the checker
	info, err = checker.LatestVersionFor(ctx, purl)
	if err != nil {
		return nil, err
	}

	rawInfo, err = json.Marshal(info)
	if err != nil {
		return nil, err
	}

	return info, r.KV.Put(ctx, cacheKey, rawInfo)
}

func cacheKeyFor(purl packageurl.PackageURL) []byte {
	return bytes.Join([][]byte{[]byte(purl.Type), []byte(purl.Namespace), []byte(purl.Name)}, []byte("/"))
}
