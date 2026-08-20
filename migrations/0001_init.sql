CREATE TABLE IF NOT EXISTS projects (
  id text PRIMARY KEY,
  name text NOT NULL,
  description text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS environments (
  id text PRIMARY KEY,
  project_id text NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  name text NOT NULL,
  kind text NOT NULL DEFAULT 'development',
  production boolean NOT NULL DEFAULT false,
  lock_key text NOT NULL,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  UNIQUE(project_id, name)
);

CREATE TABLE IF NOT EXISTS resources (
  id text PRIMARY KEY,
  environment_id text NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
  name text NOT NULL,
  type text NOT NULL,
  provider text NOT NULL DEFAULT 'mock',
  desired_state jsonb NOT NULL,
  depends_on text[] NOT NULL DEFAULT '{}',
  version integer NOT NULL DEFAULT 1,
  sensitive boolean NOT NULL DEFAULT false,
  approval_required boolean NOT NULL DEFAULT false,
  lock_key text NOT NULL,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  UNIQUE(environment_id, name)
);

CREATE TABLE IF NOT EXISTS variables (
  id text PRIMARY KEY,
  project_id text NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  key text NOT NULL,
  value text NOT NULL,
  sensitive boolean NOT NULL DEFAULT false,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  UNIQUE(project_id, key)
);

CREATE TABLE IF NOT EXISTS resource_states (
  id text PRIMARY KEY,
  resource_id text NOT NULL REFERENCES resources(id) ON DELETE CASCADE,
  environment_id text NOT NULL,
  desired_state jsonb NOT NULL DEFAULT '{}'::jsonb,
  actual_state jsonb NOT NULL DEFAULT '{}'::jsonb,
  last_execution_status text NOT NULL DEFAULT 'pending',
  version integer NOT NULL,
  lock_version integer NOT NULL DEFAULT 0,
  updated_at timestamptz NOT NULL,
  UNIQUE(resource_id, version)
);

CREATE TABLE IF NOT EXISTS state_snapshots (
  id text PRIMARY KEY,
  resource_id text NOT NULL REFERENCES resources(id) ON DELETE CASCADE,
  desired_state jsonb NOT NULL,
  actual_state jsonb NOT NULL,
  status text NOT NULL,
  version integer NOT NULL,
  captured_at timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS plans (
  id text PRIMARY KEY,
  environment_id text NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
  status text NOT NULL,
  diff jsonb NOT NULL DEFAULT '[]'::jsonb,
  created_by text NOT NULL DEFAULT 'system',
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS plan_diff_items (
  id text PRIMARY KEY,
  plan_id text NOT NULL REFERENCES plans(id) ON DELETE CASCADE,
  position integer NOT NULL,
  resource_id text NOT NULL,
  name text NOT NULL,
  type text NOT NULL,
  operation text NOT NULL,
  before jsonb NOT NULL,
  after jsonb NOT NULL
);

CREATE TABLE IF NOT EXISTS approval_requests (
  id text PRIMARY KEY,
  plan_id text NOT NULL REFERENCES plans(id) ON DELETE CASCADE,
  environment_id text NOT NULL,
  requested_by text NOT NULL,
  reason text NOT NULL DEFAULT '',
  status text NOT NULL,
  approved_by text NOT NULL DEFAULT '',
  decision_note text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS execution_tasks (
  id text PRIMARY KEY,
  plan_id text NOT NULL REFERENCES plans(id) ON DELETE CASCADE,
  resource_id text NOT NULL,
  environment_id text NOT NULL,
  status text NOT NULL,
  attempt integer NOT NULL DEFAULT 0,
  max_attempts integer NOT NULL DEFAULT 3,
  input jsonb NOT NULL,
  output jsonb NOT NULL DEFAULT '{}'::jsonb,
  error text NOT NULL DEFAULT '',
  lock_key text NOT NULL,
  timeout integer NOT NULL DEFAULT 30000,
  lease_owner text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL,
  started_at timestamptz,
  completed_at timestamptz,
  next_run_at timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS drift_records (
  id text PRIMARY KEY,
  environment_id text NOT NULL,
  resource_id text NOT NULL,
  desired_state jsonb NOT NULL,
  actual_state jsonb NOT NULL,
  severity text NOT NULL,
  detected_at timestamptz NOT NULL,
  resolved boolean NOT NULL DEFAULT false,
  remedy_plan_id text NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS audit_events (
  id text PRIMARY KEY,
  type text NOT NULL,
  actor text NOT NULL,
  entity_type text NOT NULL,
  entity_id text NOT NULL,
  content jsonb NOT NULL DEFAULT '{}'::jsonb,
  result text NOT NULL DEFAULT '',
  error text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_resources_env ON resources(environment_id);
CREATE INDEX IF NOT EXISTS idx_states_resource ON resource_states(resource_id, version DESC);
CREATE INDEX IF NOT EXISTS idx_plans_env ON plans(environment_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_tasks_claim ON execution_tasks(status, next_run_at, lease_owner);
CREATE INDEX IF NOT EXISTS idx_drift_env_resolved ON drift_records(environment_id, resolved);
CREATE INDEX IF NOT EXISTS idx_audit_entity ON audit_events(entity_type, entity_id, created_at DESC);
