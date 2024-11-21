package pypi_test

import (
	_ "embed"
	"testing"

	"github.com/package-url/packageurl-go"
	"github.com/stretchr/testify/assert"

	"github.com/prskr/aucs/core/ports"
	"github.com/prskr/aucs/infrastructure/checker/pypi"
	"github.com/prskr/aucs/internal/testx"
)

//go:embed testdata/requests.json
var requestsResponse []byte

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
			name: "Outdated existing dependency",
			args: args{
				packageUrl: "pkg:pypi/requests@2.29.0",
			},
			fields: fields{
				clientConfig: map[string][]byte{
					"https://pypi.org/pypi/requests/json": requestsResponse,
				},
			},
			want: &ports.PackageInfo{
				Name:           "requests",
				CurrentVersion: "2.29.0",
				LatestVersion:  "2.32.3",
				PackageManager: "pypi",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
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

			c := pypi.NewChecker(testx.MockHTTPClient(responseRules...))
			purl, err := packageurl.FromString(tt.args.packageUrl)
			if !assert.NoError(t, err) {
				return
			}

			got, err := c.LatestVersionFor(testx.Context(t), purl)
			if (err != nil) != tt.wantErr {
				t.Errorf("Checker.LatestVersionFor() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equal(t, tt.want, got)
		})
	}
}
