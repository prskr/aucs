package java_test

import (
	_ "embed"
	"testing"

	"github.com/package-url/packageurl-go"
	"github.com/stretchr/testify/assert"

	"github.com/prskr/aucs/core/ports"
	"github.com/prskr/aucs/infrastructure/checker/java"
	"github.com/prskr/aucs/internal/testx"
)

var (
	//go:embed testdata/jackson-databind-metadata.xml
	jacksonDatabindResponse []byte
	//go:embed testdata/spring-boot-starter-web-metadata.xml
	springBootStarterWebResponse []byte
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
			name: "Outdated, existing dependency",
			args: args{
				packageUrl: "pkg:maven/com.fasterxml.jackson.core/jackson-databind@2.17.2",
			},
			fields: fields{
				clientConfig: map[string][]byte{
					"https://repo.maven.apache.org/maven2/com/fasterxml/jackson/core/jackson-databind/maven-metadata.xml": jacksonDatabindResponse,
				},
			},
			want: &ports.PackageInfo{
				Namespace:      "com.fasterxml.jackson.core",
				Name:           "jackson-databind",
				CurrentVersion: "2.17.2",
				LatestVersion:  "2.18.1",
				PackageManager: "maven",
			},
			wantErr: false,
		},
		{
			name: "Dependency without version",
			args: args{
				packageUrl: "pkg:maven/org.springframework.boot/spring-boot-starter-web",
			},
			fields: fields{
				clientConfig: map[string][]byte{
					"https://repo.maven.apache.org/maven2/org/springframework/boot/spring-boot-starter-web/maven-metadata.xml": springBootStarterWebResponse,
				},
			},
			want: &ports.PackageInfo{
				Namespace:      "org.springframework.boot",
				Name:           "spring-boot-starter-web",
				CurrentVersion: "",
				LatestVersion:  "3.3.6",
				PackageManager: "maven",
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

			c := java.NewChecker(testx.MockHTTPClient(responseRules...))
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
