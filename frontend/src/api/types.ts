export type ServiceStatus =
  | "provisioning"
  | "running"
  | "deploying"
  | "degraded"
  | "stopped"
  | "failed";

export interface Service {
  id: string;
  owner_id: string;
  name: string;
  repo_url?: string;
  branch: string;
  build_type: "dockerfile" | "buildpacks" | "static";
  image?: string;
  status: ServiceStatus;
  created_at: string;
  updated_at: string;
}

export interface Deployment {
  id: string;
  service_id: string;
  commit_sha?: string;
  status: ServiceStatus;
  image?: string;
  logs?: string;
  created_at: string;
  finished_at?: string | null;
}

export interface LogEntry {
  deployment_id: string;
  service_id: string;
  stream: "stdout" | "stderr";
  message: string;
  timestamp: string;
}
