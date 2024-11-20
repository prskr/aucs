package npm_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/package-url/packageurl-go"
	"github.com/stretchr/testify/assert"

	"github.com/prskr/aucs/infrastructure/checker/npm"
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
			name: "Outdated, existing dependency",
			args: args{
				packageUrl: "pkg:npm/%40ampproject/remapping@2.2.1",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := npm.NewChecker(http.DefaultClient)
			purl, err := packageurl.FromString(tt.args.packageUrl)
			if !assert.NoError(t, err) {
				return
			}
			got, err := c.LatestVersionFor(context.TODO(), purl)
			if (err != nil) != tt.wantErr {
				t.Errorf("Checker.LatestVersionFor() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Log(got.Name)
			assert.NotEmpty(t, got.LatestVersion)
		})
	}
}
