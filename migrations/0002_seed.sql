INSERT INTO projects(id,name,description,created_at,updated_at)
SELECT 'project-demo','demo-project','Seeded demonstration project',now(),now()
WHERE NOT EXISTS (SELECT 1 FROM projects WHERE id='project-demo');

INSERT INTO environments(id,project_id,name,kind,production,lock_key,created_at,updated_at)
SELECT 'env-demo','project-demo','development','development',false,'env:project-demo:development',now(),now()
WHERE NOT EXISTS (SELECT 1 FROM environments WHERE id='env-demo');
