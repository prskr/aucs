package java_test

import (
	"net/http"
	"testing"

	"github.com/package-url/packageurl-go"
	"github.com/stretchr/testify/assert"

	"github.com/prskr/aucs/infrastructure/checker/java"
	"github.com/prskr/aucs/internal/testx"
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
				packageUrl: "pkg:maven/org.springframework.boot/spring-boot-starter-web",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := java.NewChecker(http.DefaultClient)
			purl, err := packageurl.FromString(tt.args.packageUrl)
			if !assert.NoError(t, err) {
				return
			}

			got, err := c.LatestVersionFor(testx.Context(t), purl)
			if (err != nil) != tt.wantErr {
				t.Errorf("Checker.LatestVersionFor() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("found latest version: %s", got.LatestVersion)
			assert.NotEmpty(t, got.LatestVersion)
		})
	}
}
