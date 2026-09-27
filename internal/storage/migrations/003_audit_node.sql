ALTER TABLE audit ADD COLUMN node_id TEXT NOT NULL DEFAULT '';
UPDATE audit SET recorded_at = recorded_at * 1000 WHERE recorded_at < 100000000000;
UPDATE audit SET node_id = substr(resource_key, 6) WHERE node_id = '' AND resource_key LIKE 'node:%';
CREATE INDEX IF NOT EXISTS audit_recorded_at ON audit(recorded_at DESC, sequence DESC);
CREATE INDEX IF NOT EXISTS audit_actor_node ON audit(actor_id, node_id, recorded_at DESC);
