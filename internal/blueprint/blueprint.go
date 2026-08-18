// Package blueprint parses and validates custodian.yaml — the declarative,
// Render-style blueprint format used to describe services, autoscaling,
// domains and persistence for a project.
package blueprint

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// Default values applied when fields are omitted.
const (
	DefaultReplicas    = 1
	DefaultMinReplicas = 1
	DefaultMaxReplicas = 10
	DefaultTargetCPU   = 70
	DefaultBranch      = "main"
)

// Blueprint is the root document of a custodian.yaml file.
type Blueprint struct {
	APIVersion string    `yaml:"apiVersion" json:"apiVersion"`
	Kind       string    `yaml:"kind" json:"kind"`
	Name       string    `yaml:"name" json:"name"`
	Services   []Service `yaml:"services" json:"services"`
}

// Service describes a single deployable unit.
type Service struct {
	Name        string     `yaml:"name" json:"name"`
	Build       Build      `yaml:"build" json:"build"`
	Runtime     Runtime    `yaml:"runtime" json:"runtime"`
	Health      *Health    `yaml:"healthCheck,omitempty" json:"healthCheck,omitempty"`
	Autoscale   *Autoscale `yaml:"autoscaling,omitempty" json:"autoscaling,omitempty"`
	Domains     []string   `yaml:"domains,omitempty" json:"domains,omitempty"`
	Persistence []Volume   `yaml:"persistence,omitempty" json:"persistence,omitempty"`
}

// Build describes how the service image is produced.
type Build struct {
	Type       string            `yaml:"type" json:"type"` // dockerfile | buildpacks | static
	Dockerfile string            `yaml:"dockerfile,omitempty" json:"dockerfile,omitempty"`
	Context    string            `yaml:"context,omitempty" json:"context,omitempty"`
	Command    string            `yaml:"command,omitempty" json:"command,omitempty"` // buildpacks start command / docker CMD
	BuildArgs  map[string]string `yaml:"buildArgs,omitempty" json:"buildArgs,omitempty"`
}

// Runtime holds environment, resources and replica configuration.
type Runtime struct {
	Replicas  int       `yaml:"replicas,omitempty" json:"replicas,omitempty"`
	Env       []EnvVar  `yaml:"env,omitempty" json:"env,omitempty"`
	Resources Resources `yaml:"resources,omitempty" json:"resources,omitempty"`
	Port      int       `yaml:"port,omitempty" json:"port,omitempty"`
}

// EnvVar is either a literal value or a reference to a stored secret.
type EnvVar struct {
	Name      string `yaml:"name" json:"name"`
	Value     string `yaml:"value,omitempty" json:"value,omitempty"`
	ValueFrom string `yaml:"valueFrom,omitempty" json:"valueFrom,omitempty"` // "secret"
}

// Resources describes CPU/memory requests and limits.
type Resources struct {
	CPU           string `yaml:"cpu,omitempty" json:"cpu,omitempty"`
	Memory        string `yaml:"memory,omitempty" json:"memory,omitempty"`
	CPURequest    string `yaml:"cpuRequest,omitempty" json:"cpuRequest,omitempty"`
	MemoryRequest string `yaml:"memoryRequest,omitempty" json:"memoryRequest,omitempty"`
}

// Health configures the readiness/liveness probe.
type Health struct {
	Path     string `yaml:"path" json:"path"`
	Interval string `yaml:"interval,omitempty" json:"interval,omitempty"`
	Timeout  string `yaml:"timeout,omitempty" json:"timeout,omitempty"`
}

// Autoscale enables HPA/KEDA-driven horizontal scaling.
type Autoscale struct {
	Enabled      bool `yaml:"enabled" json:"enabled"`
	MinReplicas  int  `yaml:"minReplicas,omitempty" json:"minReplicas,omitempty"`
	MaxReplicas  int  `yaml:"maxReplicas,omitempty" json:"maxReplicas,omitempty"`
	TargetCPU    int  `yaml:"targetCPU,omitempty" json:"targetCPU,omitempty"`
	TargetMemory int  `yaml:"targetMemory,omitempty" json:"targetMemory,omitempty"`
	ScaleToZero  bool `yaml:"scaleToZero,omitempty" json:"scaleToZero,omitempty"`
}

// Volume requests persistent storage.
type Volume struct {
	Name      string `yaml:"name" json:"name"`
	Size      string `yaml:"size" json:"size"`
	MountPath string `yaml:"mountPath" json:"mountPath"`
}

// Parse decodes and validates a custodian.yaml document.
func Parse(data []byte) (*Blueprint, error) {
	var bp Blueprint
	if err := yaml.Unmarshal(data, &bp); err != nil {
		return nil, fmt.Errorf("parse blueprint: %w", err)
	}
	if err := bp.Validate(); err != nil {
		return nil, err
	}
	return &bp, nil
}

// Validate applies normalization defaults and structural checks.
func (bp *Blueprint) Validate() error {
	if bp.APIVersion == "" {
		bp.APIVersion = "custodian.dev/v1"
	}
	if bp.Kind == "" {
		bp.Kind = "Blueprint"
	}
	if bp.Name == "" {
		return fmt.Errorf("blueprint: name is required")
	}
	if len(bp.Services) == 0 {
		return fmt.Errorf("blueprint %q: at least one service is required", bp.Name)
	}
	seen := map[string]bool{}
	for i := range bp.Services {
		s := &bp.Services[i]
		if s.Name == "" {
			return fmt.Errorf("blueprint %q: service %d has no name", bp.Name, i+1)
		}
		if seen[s.Name] {
			return fmt.Errorf("blueprint %q: duplicate service name %q", bp.Name, s.Name)
		}
		seen[s.Name] = true

		switch s.Build.Type {
		case "", "dockerfile":
			s.Build.Type = "dockerfile"
		case "buildpacks", "static":
		default:
			return fmt.Errorf("service %q: unsupported build type %q (want dockerfile, buildpacks or static)", s.Name, s.Build.Type)
		}
		if s.Build.Context == "" {
			s.Build.Context = "."
		}
		if s.Build.Dockerfile == "" {
			s.Build.Dockerfile = "Dockerfile"
		}

		if s.Runtime.Replicas == 0 {
			s.Runtime.Replicas = DefaultReplicas
		}
		if s.Runtime.Port == 0 {
			s.Runtime.Port = 8080
		}
		if s.Health != nil {
			if !strings.HasPrefix(s.Health.Path, "/") {
				s.Health.Path = "/" + s.Health.Path
			}
			if s.Health.Interval == "" {
				s.Health.Interval = "10s"
			}
			if s.Health.Timeout == "" {
				s.Health.Timeout = "3s"
			}
		}
		if s.Autoscale != nil {
			if s.Autoscale.MinReplicas == 0 {
				s.Autoscale.MinReplicas = DefaultMinReplicas
			}
			if s.Autoscale.MaxReplicas == 0 {
				s.Autoscale.MaxReplicas = DefaultMaxReplicas
			}
			if s.Autoscale.MinReplicas > s.Autoscale.MaxReplicas {
				return fmt.Errorf("service %q: autoscaling minReplicas > maxReplicas", s.Name)
			}
			if s.Autoscale.TargetCPU == 0 {
				s.Autoscale.TargetCPU = DefaultTargetCPU
			}
		}
		for _, v := range s.Persistence {
			if v.Name == "" || v.MountPath == "" {
				return fmt.Errorf("service %q: persistence entries require name and mountPath", s.Name)
			}
		}
	}
	return nil
}

// FindService returns a service by name.
func (bp *Blueprint) FindService(name string) (*Service, bool) {
	for i := range bp.Services {
		if bp.Services[i].Name == name {
			return &bp.Services[i], true
		}
	}
	return nil, false
}
