// Package deployer renders deployable artifacts (docker-compose files or
// Kubernetes manifests) from a validated blueprint.
package deployer

import (
	"fmt"

	"github.com/CocoCopi/custodian/internal/blueprint"
)

// Engine selects the target runtime for rendered artifacts.
type Engine string

const (
	// EngineCompose renders docker-compose.yml for single-server deployments.
	EngineCompose Engine = "compose"
	// EngineK3s renders Kubernetes manifests for k3s clusters.
	EngineK3s Engine = "k3s"
)

// Result is the set of files produced for a project.
type Result struct {
	// Compose is the docker-compose.yml content for EngineCompose.
	Compose []byte
	// Manifests maps a relative filename to Kubernetes YAML for EngineK3s.
	Manifests map[string][]byte
}

// Render produces deployment artifacts for the given engine.
func Render(engine Engine, bp *blueprint.Blueprint, registry string) (*Result, error) {
	switch engine {
	case EngineCompose:
		return renderCompose(bp, registry)
	case EngineK3s:
		return renderK3s(bp, registry)
	default:
		return nil, fmt.Errorf("unknown deploy engine %q", engine)
	}
}
