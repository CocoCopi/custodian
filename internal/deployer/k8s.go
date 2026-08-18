package deployer

import (
	"bytes"
	"fmt"

	"github.com/CocoCopi/custodian/internal/blueprint"
	"gopkg.in/yaml.v3"
)

const (
	namespaceName = "custodian"
	appLabel      = "app.kubernetes.io/name"
)

// renderK3s produces a full set of Kubernetes manifests for a k3s cluster:
// namespace, deployments, services, HPAs, ingresses and persistent volumes.
func renderK3s(bp *blueprint.Blueprint, registry string) (*Result, error) {
	r := &Result{Manifests: map[string][]byte{}}

	r.Manifests["00-namespace.yaml"] = mustYAML(map[string]any{
		"apiVersion": "v1",
		"kind":       "Namespace",
		"metadata":   map[string]any{"name": namespaceName, "labels": map[string]string{"custodian.dev/project": bp.Name}},
	})

	for _, svc := range bp.Services {
		r.Manifests[fmt.Sprintf("10-%s-deployment.yaml", svc.Name)] = deployment(bp, svc, registry)
		r.Manifests[fmt.Sprintf("11-%s-service.yaml", svc.Name)] = service(svc)
		if svc.Autoscale != nil && svc.Autoscale.Enabled {
			r.Manifests[fmt.Sprintf("12-%s-hpa.yaml", svc.Name)] = hpa(svc)
		}
		if len(svc.Domains) > 0 {
			r.Manifests[fmt.Sprintf("13-%s-ingress.yaml", svc.Name)] = ingress(bp, svc)
		}
		for _, v := range svc.Persistence {
			r.Manifests[fmt.Sprintf("14-%s-pvc-%s.yaml", svc.Name, v.Name)] = pvc(bp, svc, v)
		}
	}

	return r, nil
}

func deployment(bp *blueprint.Blueprint, svc blueprint.Service, registry string) []byte {
	env := []any{}
	for _, e := range svc.Runtime.Env {
		item := map[string]any{"name": e.Name}
		if e.ValueFrom == "secret" {
			item["valueFrom"] = map[string]any{
				"secretKeyRef": map[string]any{
					"name": fmt.Sprintf("custodian-%s-%s", bp.Name, svc.Name),
					"key":  e.Name,
				},
			}
		} else {
			item["value"] = e.Value
		}
		env = append(env, item)
	}

	container := map[string]any{
		"name":  svc.Name,
		"image": imageRef(registry, bp.Name, svc.Name),
		"ports": []any{map[string]any{"containerPort": svc.Runtime.Port}},
		"env":   env,
	}
	if svc.Health != nil {
		container["readinessProbe"] = probe(svc.Health.Path, svc.Runtime.Port, svc.Health.Interval, svc.Health.Timeout)
		container["livenessProbe"] = probe(svc.Health.Path, svc.Runtime.Port, svc.Health.Interval, svc.Health.Timeout)
	}
	if svc.Runtime.Resources.CPU != "" || svc.Runtime.Resources.Memory != "" {
		limits := map[string]any{}
		if svc.Runtime.Resources.CPU != "" {
			limits["cpu"] = svc.Runtime.Resources.CPU
		}
		if svc.Runtime.Resources.Memory != "" {
			limits["memory"] = svc.Runtime.Resources.Memory
		}
		container["resources"] = map[string]any{"limits": limits}
	}

	volumes := []any{}
	volumeMounts := []any{}
	for _, v := range svc.Persistence {
		volumes = append(volumes, map[string]any{
			"name": v.Name,
			"persistentVolumeClaim": map[string]any{
				"claimName": fmt.Sprintf("%s-%s-%s", bp.Name, svc.Name, v.Name),
			},
		})
		volumeMounts = append(volumeMounts, map[string]any{"name": v.Name, "mountPath": v.MountPath})
	}
	if len(volumes) > 0 {
		container["volumeMounts"] = volumeMounts
	}

	return mustYAML(map[string]any{
		"apiVersion": "apps/v1",
		"kind":       "Deployment",
		"metadata": map[string]any{
			"name":      svc.Name,
			"namespace": namespaceName,
			"labels":    map[string]string{appLabel: svc.Name, "custodian.dev/project": bp.Name},
		},
		"spec": map[string]any{
			"replicas": svc.Runtime.Replicas,
			"selector": map[string]any{"matchLabels": map[string]string{appLabel: svc.Name}},
			"template": map[string]any{
				"metadata": map[string]any{"labels": map[string]string{appLabel: svc.Name}},
				"spec": map[string]any{
					"containers": []any{container},
					"volumes":    volumes,
				},
			},
		},
	})
}

func service(svc blueprint.Service) []byte {
	return mustYAML(map[string]any{
		"apiVersion": "v1",
		"kind":       "Service",
		"metadata": map[string]any{
			"name":      svc.Name,
			"namespace": namespaceName,
			"labels":    map[string]string{appLabel: svc.Name},
		},
		"spec": map[string]any{
			"selector": map[string]string{appLabel: svc.Name},
			"ports":    []any{map[string]any{"port": 80, "targetPort": svc.Runtime.Port}},
		},
	})
}

func hpa(svc blueprint.Service) []byte {
	metrics := []any{}
	if svc.Autoscale.TargetCPU > 0 {
		metrics = append(metrics, map[string]any{
			"type": "Resource",
			"resource": map[string]any{
				"name":   "cpu",
				"target": map[string]any{"type": "Utilization", "averageUtilization": svc.Autoscale.TargetCPU},
			},
		})
	}
	if svc.Autoscale.TargetMemory > 0 {
		metrics = append(metrics, map[string]any{
			"type": "Resource",
			"resource": map[string]any{
				"name":   "memory",
				"target": map[string]any{"type": "Utilization", "averageUtilization": svc.Autoscale.TargetMemory},
			},
		})
	}
	return mustYAML(map[string]any{
		"apiVersion": "autoscaling/v2",
		"kind":       "HorizontalPodAutoscaler",
		"metadata": map[string]any{
			"name":      svc.Name,
			"namespace": namespaceName,
			"labels":    map[string]string{appLabel: svc.Name},
		},
		"spec": map[string]any{
			"scaleTargetRef": map[string]any{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"name":       svc.Name,
			},
			"minReplicas": svc.Autoscale.MinReplicas,
			"maxReplicas": svc.Autoscale.MaxReplicas,
			"metrics":     metrics,
		},
	})
}

func ingress(bp *blueprint.Blueprint, svc blueprint.Service) []byte {
	rules := []any{}
	for _, domain := range svc.Domains {
		rules = append(rules, map[string]any{
			"host": domain,
			"http": map[string]any{
				"paths": []any{map[string]any{
					"path":     "/",
					"pathType": "Prefix",
					"backend": map[string]any{
						"service": map[string]any{"name": svc.Name, "port": map[string]any{"number": 80}},
					},
				}},
			},
		})
	}
	return mustYAML(map[string]any{
		"apiVersion": "networking.k8s.io/v1",
		"kind":       "Ingress",
		"metadata": map[string]any{
			"name":      svc.Name,
			"namespace": namespaceName,
			"labels":    map[string]string{appLabel: svc.Name, "custodian.dev/project": bp.Name},
			"annotations": map[string]string{
				"cert-manager.io/cluster-issuer":                   "letsencrypt-prod",
				"traefik.ingress.kubernetes.io/router.entrypoints": "websecure",
			},
		},
		"spec": map[string]any{
			"tls": []any{map[string]any{
				"hosts":      svc.Domains,
				"secretName": fmt.Sprintf("%s-%s-tls", bp.Name, svc.Name),
			}},
			"rules": rules,
		},
	})
}

func pvc(bp *blueprint.Blueprint, svc blueprint.Service, v blueprint.Volume) []byte {
	return mustYAML(map[string]any{
		"apiVersion": "v1",
		"kind":       "PersistentVolumeClaim",
		"metadata": map[string]any{
			"name":      fmt.Sprintf("%s-%s-%s", bp.Name, svc.Name, v.Name),
			"namespace": namespaceName,
		},
		"spec": map[string]any{
			"accessModes": []string{"ReadWriteOnce"},
			"resources": map[string]any{
				"requests": map[string]any{"storage": v.Size},
			},
		},
	})
}

func probe(path string, port int, interval, timeout string) map[string]any {
	return map[string]any{
		"httpGet":             map[string]any{"path": path, "port": port},
		"initialDelaySeconds": 5,
		"periodSeconds":       10,
		"timeoutSeconds":      3,
		"failureThreshold":    3,
	}
}

func mustYAML(v any) []byte {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(v); err != nil {
		panic(fmt.Sprintf("encode manifest: %v", err))
	}
	return buf.Bytes()
}
