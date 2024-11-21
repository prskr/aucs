package nuget_test

import (
	_ "embed"
	"testing"

	"github.com/package-url/packageurl-go"
	"github.com/stretchr/testify/assert"

	"github.com/prskr/aucs/core/ports"
	"github.com/prskr/aucs/infrastructure/checker/nuget"
	"github.com/prskr/aucs/internal/testx"
)

//go:embed testdata/BouncyCastle.Cryptography.json
var bouncyCastleCryptographyResponse []byte

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
			name: "Deprecated, existing dependency",
			args: args{
				packageUrl: "pkg:nuget/BouncyCastle.Cryptography@2.2.1",
			},
			fields: fields{
				clientConfig: map[string][]byte{
					"https://azuresearch-usnc.nuget.org/query?q=packageid:BouncyCastle.Cryptography": bouncyCastleCryptographyResponse,
				},
			},
			want: &ports.PackageInfo{
				Name:           "BouncyCastle.Cryptography",
				CurrentVersion: "2.2.1",
				LatestVersion:  "2.4.0",
				PackageManager: "nuget",
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

			c := nuget.NewChecker(testx.MockHTTPClient(responseRules...))

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
