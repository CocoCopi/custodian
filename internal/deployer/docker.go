package deployer

import (
	"bytes"
	"fmt"

	"github.com/CocoCopi/custodian/internal/blueprint"
	"gopkg.in/yaml.v3"
)

// renderCompose turns a blueprint into a docker-compose.yml targeting
// single-server deployments fronted by Traefik.
func renderCompose(bp *blueprint.Blueprint, registry string) (*Result, error) {
	services := map[string]any{}

	for _, svc := range bp.Services {
		env := map[string]any{}
		for _, e := range svc.Runtime.Env {
			if e.ValueFrom == "secret" {
				// Secrets are injected by the control plane at deploy time.
				env[e.Name] = "${CUSTODIAN_SECRET_" + e.Name + ":-}"
			} else {
				env[e.Name] = e.Value
			}
		}

		labels := map[string]string{
			"traefik.enable": "true",
		}
		for i, domain := range svc.Domains {
			if i == 0 {
				labels["traefik.http.routers."+svc.Name+".rule"] = "Host(`" + domain + "`)"
				labels["traefik.http.routers."+svc.Name+".entrypoints"] = "websecure"
				labels["traefik.http.routers."+svc.Name+".tls.certresolver"] = "letsencrypt"
				labels["traefik.http.services."+svc.Name+".loadbalancer.server.port"] = fmt.Sprintf("%d", svc.Runtime.Port)
			} else {
				labels[fmt.Sprintf("traefik.http.routers.%s_%d.rule", svc.Name, i)] = "Host(`" + domain + "`)"
			}
		}

		serviceDef := map[string]any{
			"build": map[string]any{
				"context":    svc.Build.Context,
				"dockerfile": svc.Build.Dockerfile,
				"args":       svc.Build.BuildArgs,
			},
			"image":       imageRef(registry, bp.Name, svc.Name),
			"restart":     "unless-stopped",
			"environment": env,
			"labels":      labels,
			"expose":      []int{svc.Runtime.Port},
			"healthcheck": healthCheck(svc),
		}

		if svc.Runtime.Replicas > 1 || (svc.Autoscale != nil && svc.Autoscale.Enabled) {
			serviceDef["deploy"] = map[string]any{
				"replicas": svc.Runtime.Replicas,
				"resources": map[string]any{
					"limits": resourceMap(svc.Runtime.Resources),
				},
				"restart_policy": map[string]any{"condition": "any"},
			}
		}

		if len(svc.Persistence) > 0 {
			volumes := []any{}
			for _, v := range svc.Persistence {
				volumes = append(volumes, fmt.Sprintf("%s_%s:%s", bp.Name, v.Name, v.MountPath))
			}
			serviceDef["volumes"] = volumes
		}

		services[svc.Name] = serviceDef
	}

	compose := map[string]any{
		"version":  "3.8",
		"services": services,
	}

	// Named volumes for persistence.
	volumes := map[string]any{}
	for _, svc := range bp.Services {
		for _, v := range svc.Persistence {
			volumes[bp.Name+"_"+v.Name] = map[string]any{}
		}
	}
	if len(volumes) > 0 {
		compose["volumes"] = volumes
	}

	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(compose); err != nil {
		return nil, fmt.Errorf("encode compose: %w", err)
	}
	if err := enc.Close(); err != nil {
		return nil, err
	}
	return &Result{Compose: buf.Bytes()}, nil
}

func healthCheck(svc blueprint.Service) map[string]any {
	path := "/"
	interval := "10s"
	timeout := "3s"
	if svc.Health != nil {
		path = svc.Health.Path
		interval = svc.Health.Interval
		timeout = svc.Health.Timeout
	}
	return map[string]any{
		"test":     []string{"CMD-SHELL", fmt.Sprintf("wget -qO- http://localhost:%d%s || exit 1", svc.Runtime.Port, path)},
		"interval": interval,
		"timeout":  timeout,
		"retries":  3,
	}
}

func resourceMap(r blueprint.Resources) map[string]any {
	out := map[string]any{}
	if r.CPU != "" {
		out["cpus"] = r.CPU
	}
	if r.Memory != "" {
		out["memory"] = r.Memory
	}
	return out
}

func imageRef(registry, project, service string) string {
	if registry == "" {
		registry = "localhost:5000"
	}
	return fmt.Sprintf("%s/custodian/%s/%s:latest", registry, project, service)
}
