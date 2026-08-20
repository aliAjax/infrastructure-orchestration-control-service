package infrastructure

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/infra-orchestration/controlplane/internal/resource/domain"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) CreateProject(ctx context.Context, input domain.ProjectInput) (domain.Project, error) {
	id := uuid.NewString()
	now := time.Now().UTC()
	var p domain.Project
	err := r.db.QueryRow(ctx, `
		INSERT INTO projects(id, name, description, created_at, updated_at)
		VALUES($1,$2,$3,$4,$5)
		RETURNING id, name, description, created_at, updated_at`,
		id, input.Name, input.Description, now, now).Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return domain.Project{}, fmt.Errorf("insert project: %w", err)
	}
	return p, nil
}

func (r *PostgresRepository) GetProject(ctx context.Context, id string) (domain.Project, error) {
	var p domain.Project
	err := r.db.QueryRow(ctx, `SELECT id,name,description,created_at,updated_at FROM projects WHERE id=$1`, id).
		Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return domain.Project{}, fmt.Errorf("get project %s: %w", id, err)
	}
	return p, nil
}

func (r *PostgresRepository) ListProjects(ctx context.Context) ([]domain.Project, error) {
	rows, err := r.db.Query(ctx, `SELECT id,name,description,created_at,updated_at FROM projects ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	defer rows.Close()
	var out []domain.Project
	for rows.Next() {
		var p domain.Project
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan project: %w", err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *PostgresRepository) UpdateProject(ctx context.Context, id string, input domain.ProjectInput) (domain.Project, error) {
	var p domain.Project
	err := r.db.QueryRow(ctx, `
		UPDATE projects SET name=$1,description=$2,updated_at=$3 WHERE id=$4
		RETURNING id,name,description,created_at,updated_at`,
		input.Name, input.Description, time.Now().UTC(), id).Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return domain.Project{}, fmt.Errorf("update project: %w", err)
	}
	return p, nil
}

func (r *PostgresRepository) CreateEnvironment(ctx context.Context, input domain.EnvironmentInput) (domain.Environment, error) {
	id := uuid.NewString()
	now := time.Now().UTC()
	var e domain.Environment
	err := r.db.QueryRow(ctx, `
		INSERT INTO environments(id, project_id, name, kind, production, lock_key, created_at, updated_at)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id,project_id,name,kind,production,lock_key,created_at,updated_at`,
		id, input.ProjectID, input.Name, input.Kind, input.Production, input.LockKey, now, now).
		Scan(&e.ID, &e.ProjectID, &e.Name, &e.Kind, &e.Production, &e.LockKey, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return domain.Environment{}, fmt.Errorf("insert environment: %w", err)
	}
	return e, nil
}

func (r *PostgresRepository) GetEnvironment(ctx context.Context, id string) (domain.Environment, error) {
	var e domain.Environment
	err := r.db.QueryRow(ctx, `SELECT id,project_id,name,kind,production,lock_key,created_at,updated_at FROM environments WHERE id=$1`, id).
		Scan(&e.ID, &e.ProjectID, &e.Name, &e.Kind, &e.Production, &e.LockKey, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return domain.Environment{}, fmt.Errorf("get environment %s: %w", id, err)
	}
	return e, nil
}

func (r *PostgresRepository) ListEnvironments(ctx context.Context, projectID string) ([]domain.Environment, error) {
	query := `SELECT id,project_id,name,kind,production,lock_key,created_at,updated_at FROM environments`
	args := []any{}
	if projectID != "" {
		query += ` WHERE project_id=$1`
		args = append(args, projectID)
	}
	query += ` ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list environments: %w", err)
	}
	defer rows.Close()
	var out []domain.Environment
	for rows.Next() {
		var e domain.Environment
		if err := rows.Scan(&e.ID, &e.ProjectID, &e.Name, &e.Kind, &e.Production, &e.LockKey, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan environment: %w", err)
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *PostgresRepository) UpsertResource(ctx context.Context, input domain.ResourceInput) (domain.Resource, error) {
	if input.DependsOn == nil {
		input.DependsOn = []string{}
	}
	var id string
	var res domain.Resource
	err := r.db.QueryRow(ctx, `SELECT id FROM resources WHERE environment_id=$1 AND name=$2`, input.EnvironmentID, input.Name).Scan(&id)
	if err != nil && err != pgx.ErrNoRows {
		return domain.Resource{}, fmt.Errorf("find resource: %w", err)
	}
	if err == pgx.ErrNoRows {
		id = uuid.NewString()
		now := time.Now().UTC()
		err = r.db.QueryRow(ctx, `
			INSERT INTO resources(id, environment_id, name, type, provider, desired_state, depends_on, version, sensitive, approval_required, lock_key, created_at, updated_at)
			VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
			RETURNING id,environment_id,name,type,provider,desired_state,depends_on,version,sensitive,approval_required,lock_key,created_at,updated_at`,
			id, input.EnvironmentID, input.Name, input.Type, input.Provider, []byte(input.DesiredState), input.DependsOn, 1, input.Sensitive, input.ApprovalRequired, input.LockKey, now, now).
			Scan(&res.ID, &res.EnvironmentID, &res.Name, &res.Type, &res.Provider, &res.DesiredState, &res.DependsOn, &res.Version, &res.Sensitive, &res.ApprovalRequired, &res.LockKey, &res.CreatedAt, &res.UpdatedAt)
		if err != nil {
			return domain.Resource{}, fmt.Errorf("insert resource: %w", err)
		}
	} else {
		err = r.db.QueryRow(ctx, `
			UPDATE resources SET type=$1,provider=$2,desired_state=$3,depends_on=$4,version=version+1,sensitive=$5,approval_required=$6,lock_key=$7,updated_at=$8
			WHERE id=$9
			RETURNING id,environment_id,name,type,provider,desired_state,depends_on,version,sensitive,approval_required,lock_key,created_at,updated_at`,
			input.Type, input.Provider, []byte(input.DesiredState), input.DependsOn, input.Sensitive, input.ApprovalRequired, input.LockKey, time.Now().UTC(), id).
			Scan(&res.ID, &res.EnvironmentID, &res.Name, &res.Type, &res.Provider, &res.DesiredState, &res.DependsOn, &res.Version, &res.Sensitive, &res.ApprovalRequired, &res.LockKey, &res.CreatedAt, &res.UpdatedAt)
		if err != nil {
			return domain.Resource{}, fmt.Errorf("update resource: %w", err)
		}
	}
	return res, nil
}

func (r *PostgresRepository) GetResource(ctx context.Context, id string) (domain.Resource, error) {
	var res domain.Resource
	err := r.db.QueryRow(ctx, `SELECT id,environment_id,name,type,provider,desired_state,depends_on,version,sensitive,approval_required,lock_key,created_at,updated_at FROM resources WHERE id=$1`, id).
		Scan(&res.ID, &res.EnvironmentID, &res.Name, &res.Type, &res.Provider, &res.DesiredState, &res.DependsOn, &res.Version, &res.Sensitive, &res.ApprovalRequired, &res.LockKey, &res.CreatedAt, &res.UpdatedAt)
	if err != nil {
		return domain.Resource{}, fmt.Errorf("get resource %s: %w", id, err)
	}
	return res, nil
}

func (r *PostgresRepository) ListResources(ctx context.Context, environmentID string) ([]domain.Resource, error) {
	query := `SELECT id,environment_id,name,type,provider,desired_state,depends_on,version,sensitive,approval_required,lock_key,created_at,updated_at FROM resources`
	args := []any{}
	if environmentID != "" {
		query += ` WHERE environment_id=$1`
		args = append(args, environmentID)
	}
	query += ` ORDER BY name`
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list resources: %w", err)
	}
	defer rows.Close()
	var out []domain.Resource
	for rows.Next() {
		var res domain.Resource
		if err := rows.Scan(&res.ID, &res.EnvironmentID, &res.Name, &res.Type, &res.Provider, &res.DesiredState, &res.DependsOn, &res.Version, &res.Sensitive, &res.ApprovalRequired, &res.LockKey, &res.CreatedAt, &res.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan resource: %w", err)
		}
		out = append(out, res)
	}
	return out, rows.Err()
}

func (r *PostgresRepository) UpsertVariable(ctx context.Context, input domain.VariableInput) (domain.Variable, error) {
	id := uuid.NewString()
	now := time.Now().UTC()
	var v domain.Variable
	err := r.db.QueryRow(ctx, `
		INSERT INTO variables(id, project_id, key, value, sensitive, created_at, updated_at)
		VALUES($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT(project_id,key) DO UPDATE SET value=EXCLUDED.value,sensitive=EXCLUDED.sensitive,updated_at=EXCLUDED.updated_at
		RETURNING id,project_id,key,value,sensitive,created_at,updated_at`,
		id, input.ProjectID, input.Key, input.Value, input.Sensitive, now, now).
		Scan(&v.ID, &v.ProjectID, &v.Key, &v.Value, &v.Sensitive, &v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		return domain.Variable{}, fmt.Errorf("upsert variable: %w", err)
	}
	return v, nil
}

func (r *PostgresRepository) GetVariable(ctx context.Context, projectID, key string) (domain.Variable, error) {
	var v domain.Variable
	err := r.db.QueryRow(ctx, `SELECT id,project_id,key,value,sensitive,created_at,updated_at FROM variables WHERE project_id=$1 AND key=$2`, projectID, key).
		Scan(&v.ID, &v.ProjectID, &v.Key, &v.Value, &v.Sensitive, &v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		return domain.Variable{}, fmt.Errorf("get variable: %w", err)
	}
	return v, nil
}

func (r *PostgresRepository) ListVariables(ctx context.Context, projectID string) ([]domain.Variable, error) {
	rows, err := r.db.Query(ctx, `SELECT id,project_id,key,value,sensitive,created_at,updated_at FROM variables WHERE project_id=$1 ORDER BY key`, projectID)
	if err != nil {
		return nil, fmt.Errorf("list variables: %w", err)
	}
	defer rows.Close()
	var out []domain.Variable
	for rows.Next() {
		var v domain.Variable
		if err := rows.Scan(&v.ID, &v.ProjectID, &v.Key, &v.Value, &v.Sensitive, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan variable: %w", err)
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
