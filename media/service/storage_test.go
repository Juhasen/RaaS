package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestLoadConfig_S3Env(t *testing.T) {
	t.Setenv("S3_ENDPOINT", "http://minio:9000")
	t.Setenv("S3_ACCESS_KEY", "ak")
	t.Setenv("S3_SECRET_KEY", "sk")
	t.Setenv("S3_BUCKET", "b")
	t.Setenv("S3_PUBLIC_URL", "http://localhost:9000/b")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if cfg.S3Endpoint != "http://minio:9000" || cfg.S3AccessKey != "ak" || cfg.S3SecretKey != "sk" ||
		cfg.S3Bucket != "b" || cfg.S3PublicURL != "http://localhost:9000/b" {
		t.Errorf("unexpected S3 config: %+v", cfg)
	}
}

func TestNewS3Client_RequiresEndpointAndCredentials(t *testing.T) {
	if c := newS3Client(&Config{}); c != nil {
		t.Error("expected nil client without S3 settings")
	}
	if c := newS3Client(&Config{S3Endpoint: "http://x:9000", S3AccessKey: "a"}); c != nil {
		t.Error("expected nil client without secret key")
	}
	if c := newS3Client(&Config{S3Endpoint: "http://x:9000", S3AccessKey: "a", S3SecretKey: "s"}); c == nil {
		t.Error("expected client when fully configured")
	}
}

// fakeS3 records requests and answers 404 to HEAD until the bucket is created.
func fakeS3(t *testing.T) (*httptest.Server, *[]string) {
	t.Helper()
	var mu sync.Mutex
	var calls []string
	created := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		calls = append(calls, r.Method+" "+r.URL.Path+"?"+r.URL.RawQuery)
		switch {
		case r.Method == http.MethodHead && !created:
			w.WriteHeader(http.StatusNotFound)
		case r.Method == http.MethodPut && r.URL.RawQuery == "":
			created = true
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusOK)
		}
	}))
	t.Cleanup(srv.Close)
	return srv, &calls
}

func TestEnsureBucket_CreatesWithPathStyleAndPublicPolicy(t *testing.T) {
	srv, calls := fakeS3(t)
	client := newS3Client(&Config{S3Endpoint: srv.URL, S3AccessKey: "a", S3SecretKey: "s"})

	if err := ensureBucket(context.Background(), client, "media"); err != nil {
		t.Fatalf("ensureBucket: %v", err)
	}

	want := []string{"HEAD /media?", "PUT /media?", "PUT /media?policy="}
	if len(*calls) != len(want) {
		t.Fatalf("calls = %v, want %v", *calls, want)
	}
	for i, w := range want {
		if (*calls)[i] != w {
			t.Errorf("call %d = %q, want %q (path-style bucket in path)", i, (*calls)[i], w)
		}
	}
}
