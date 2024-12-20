package npm_test

import (
	_ "embed"
	"testing"

	"github.com/package-url/packageurl-go"
	"github.com/stretchr/testify/assert"

	"github.com/prskr/aucs/core/ports"
	"github.com/prskr/aucs/infrastructure/checker/npm"
	"github.com/prskr/aucs/internal/testx"
)

var (
	//go:embed testdata/is_even_ai.json
	isEvenAIResponse []byte
	//go:embed testdata/ampproject_remapping.json
	ampProjectRemappingResponse []byte
)

func TestChecker_LatestVersionFor(t *testing.T) {
	t.Parallel()

	type args struct {
		packageUrl string
	}
	type fields struct {
		clientConfig map[string][]byte
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    *ports.PackageInfo
		wantErr bool
	}{
		{
			name: "Outdated, existing dependency - no namespace",
			args: args{
				packageUrl: "pkg:npm/is-even-ai@1.0.1",
			},
			fields: fields{
				clientConfig: map[string][]byte{
					"https://registry.npmjs.org/is-even-ai": isEvenAIResponse,
				},
			},
			want: &ports.PackageInfo{
				Name:           "is-even-ai",
				CurrentVersion: "1.0.1",
				LatestVersion:  "1.0.5",
			},
			wantErr: false,
		},
		{
			name: "Outdated, existing dependency",
			args: args{
				packageUrl: "pkg:npm/%40ampproject/remapping@2.2.1",
			},
			fields: fields{
				clientConfig: map[string][]byte{
					"https://registry.npmjs.org/%40ampproject/remapping": ampProjectRemappingResponse,
				},
			},
			want: &ports.PackageInfo{
				Namespace:      "@ampproject",
				Name:           "remapping",
				CurrentVersion: "2.2.1",
				LatestVersion:  "2.3.0",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			responseRules := make([]testx.ResponseRule, 0, len(tt.fields.clientConfig))
			for rawUrl, resp := range tt.fields.clientConfig {
				respRule, err := testx.NewSimpleUrlRule(rawUrl, resp)
				if !assert.NoError(t, err) {
					return
				}
				responseRules = append(responseRules, respRule)
			}

			c := npm.NewChecker(testx.MockHTTPClient(responseRules...))
			purl, err := packageurl.FromString(tt.args.packageUrl)
			if !assert.NoError(t, err) {
				return
			}

			got, err := c.LatestVersionFor(testx.Context(t), purl)
			if (err != nil) != tt.wantErr {
				t.Errorf("Checker.LatestVersionFor() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Log(got.Name)
			assert.NotEmpty(t, got.LatestVersion)
		})
	}
}
