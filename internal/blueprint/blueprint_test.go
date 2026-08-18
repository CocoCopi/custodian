package blueprint

import (
	"strings"
	"testing"
)

const sample = `
apiVersion: custodian.dev/v1
kind: Blueprint
name: acme
services:
  - name: web
    build:
      type: dockerfile
      context: ./web
    runtime:
      replicas: 2
      port: 3000
      env:
        - name: NODE_ENV
          value: production
        - name: DATABASE_URL
          valueFrom: secret
      resources:
        cpu: "0.5"
        memory: "512Mi"
    healthCheck:
      path: /healthz
    autoscaling:
      enabled: true
      minReplicas: 2
      maxReplicas: 8
      targetCPU: 60
    domains:
      - acme.example.com
    persistence:
      - name: data
        size: 10Gi
        mountPath: /var/lib/data
  - name: worker
    build:
      type: buildpacks
`

func TestParseValid(t *testing.T) {
	bp, err := Parse([]byte(sample))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if bp.Name != "acme" {
		t.Errorf("Name = %q, want acme", bp.Name)
	}
	if len(bp.Services) != 2 {
		t.Fatalf("got %d services, want 2", len(bp.Services))
	}
	web, ok := bp.FindService("web")
	if !ok {
		t.Fatal("web service not found")
	}
	if web.Runtime.Replicas != 2 {
		t.Errorf("replicas = %d, want 2", web.Runtime.Replicas)
	}
	if web.Runtime.Port != 3000 {
		t.Errorf("port = %d, want 3000", web.Runtime.Port)
	}
	if web.Health == nil || web.Health.Path != "/healthz" {
		t.Errorf("health check not normalized: %+v", web.Health)
	}
	if web.Autoscale == nil || web.Autoscale.MaxReplicas != 8 {
		t.Errorf("autoscale not parsed: %+v", web.Autoscale)
	}
	if web.Build.Context != "./web" {
		t.Errorf("build context = %q, want ./web", web.Build.Context)
	}
	if web.Runtime.Env[1].ValueFrom != "secret" {
		t.Errorf("env secret ref not parsed")
	}
}

func TestParseDefaults(t *testing.T) {
	bp, err := Parse([]byte(sample))
	if err != nil {
		t.Fatal(err)
	}
	worker, _ := bp.FindService("worker")
	if worker.Build.Type != "buildpacks" {
		t.Errorf("build type = %q, want buildpacks", worker.Build.Type)
	}
	if worker.Runtime.Replicas != 1 {
		t.Errorf("default replicas = %d, want 1", worker.Runtime.Replicas)
	}
	if worker.Runtime.Port != 8080 {
		t.Errorf("default port = %d, want 8080", worker.Runtime.Port)
	}
}

func TestParseRejects(t *testing.T) {
	cases := []struct {
		name string
		yaml string
		want string
	}{
		{"missing name", "services:\n  - name: web\n", "name is required"},
		{"no services", "name: x\n", "at least one service"},
		{"empty service name", "name: x\nservices:\n  - build:\n      type: dockerfile\n", "has no name"},
		{"bad build type", "name: x\nservices:\n  - name: a\n    build:\n      type: heroku\n", "unsupported build type"},
		{"bad scale range", "name: x\nservices:\n  - name: a\n    autoscaling:\n      enabled: true\n      minReplicas: 5\n      maxReplicas: 2\n", "minReplicas > maxReplicas"},
		{"bad yaml", "name: [unclosed\n", "parse blueprint"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Parse([]byte(tc.yaml))
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tc.want)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error %q does not contain %q", err.Error(), tc.want)
			}
		})
	}
}
