package pypi_test

import (
	"net/http"
	"testing"

	packageurl "github.com/package-url/packageurl-go"
	"github.com/prskr/aucs/infrastructure/checker/pypi"
	"github.com/prskr/aucs/internal/testx"
	"github.com/stretchr/testify/assert"
)

func TestChecker_LatestVersionFor(t *testing.T) {
	t.Parallel()

	type args struct {
		packageUrl string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Outdated existing dependency",
			args: args{
				packageUrl: "pkg:pypi/pytest-httpbin@1.0.2",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := pypi.NewChecker(http.DefaultClient)
			purl, err := packageurl.FromString(tt.args.packageUrl)
			if !assert.NoError(t, err) {
				return
			}

			got, err := c.LatestVersionFor(testx.Context(t), purl)
			if (err != nil) != tt.wantErr {
				t.Errorf("Checker.LatestVersionFor() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("%s, current: %s, latest: %s", got.Name, got.CurrentVersion, got.LatestVersion)
			assert.NotEmpty(t, got.LatestVersion)
		})
	}
}
