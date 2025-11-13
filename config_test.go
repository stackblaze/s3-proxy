package main

import (
	"os"
	"testing"
)

func TestValidateSite(t *testing.T) {
	tests := []struct {
		name    string
		site    Site
		wantErr bool
	}{
		{
			name: "valid site",
			site: Site{
				AWSKey:    "test-key",
				AWSSecret: "test-secret",
				AWSRegion: "us-east-1",
				AWSBucket: "test-bucket",
			},
			wantErr: false,
		},
		{
			name: "missing key",
			site: Site{
				AWSSecret: "test-secret",
				AWSRegion: "us-east-1",
				AWSBucket: "test-bucket",
			},
			wantErr: true,
		},
		{
			name: "missing secret",
			site: Site{
				AWSKey:    "test-key",
				AWSRegion: "us-east-1",
				AWSBucket: "test-bucket",
			},
			wantErr: true,
		},
		{
			name: "missing region",
			site: Site{
				AWSKey:    "test-key",
				AWSSecret: "test-secret",
				AWSBucket: "test-bucket",
			},
			wantErr: true,
		},
		{
			name: "missing bucket",
			site: Site{
				AWSKey:    "test-key",
				AWSSecret: "test-secret",
				AWSRegion: "us-east-1",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.site.validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateSiteWithHost(t *testing.T) {
	tests := []struct {
		name    string
		site    Site
		wantErr bool
	}{
		{
			name: "valid with host",
			site: Site{
				Host:      "example.com",
				AWSKey:    "test-key",
				AWSSecret: "test-secret",
				AWSRegion: "us-east-1",
				AWSBucket: "test-bucket",
			},
			wantErr: false,
		},
		{
			name: "missing host",
			site: Site{
				AWSKey:    "test-key",
				AWSSecret: "test-secret",
				AWSRegion: "us-east-1",
				AWSBucket: "test-bucket",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.site.validateWithHost()
			if (err != nil) != tt.wantErr {
				t.Errorf("validateWithHost() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseUsers(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{
			name:    "single user",
			input:   "user1:pass1",
			want:    1,
			wantErr: false,
		},
		{
			name:    "multiple users",
			input:   "user1:pass1,user2:pass2",
			want:    2,
			wantErr: false,
		},
		{
			name:    "empty string",
			input:   "",
			want:    0,
			wantErr: false,
		},
		{
			name:    "invalid format",
			input:   "user1",
			want:    0,
			wantErr: true,
		},
		{
			name:    "invalid format no colon",
			input:   "user1pass1",
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			users, err := parseUsers(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseUsers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(users) != tt.want {
				t.Errorf("parseUsers() len = %v, want %v", len(users), tt.want)
			}
		})
	}
}

func TestConfiguredProxyHandler_SingleMode(t *testing.T) {
	// Set up single mode environment
	os.Setenv("S3PROXY_AWS_KEY", "test-key")
	os.Setenv("S3PROXY_AWS_SECRET", "test-secret")
	os.Setenv("S3PROXY_AWS_REGION", "us-east-1")
	os.Setenv("S3PROXY_AWS_BUCKET", "test-bucket")
	defer func() {
		os.Unsetenv("S3PROXY_AWS_KEY")
		os.Unsetenv("S3PROXY_AWS_SECRET")
		os.Unsetenv("S3PROXY_AWS_REGION")
		os.Unsetenv("S3PROXY_AWS_BUCKET")
	}()

	handler, err := ConfiguredProxyHandler()
	if err != nil {
		t.Fatalf("ConfiguredProxyHandler() error = %v", err)
	}
	if handler == nil {
		t.Fatal("ConfiguredProxyHandler() returned nil handler")
	}
}

