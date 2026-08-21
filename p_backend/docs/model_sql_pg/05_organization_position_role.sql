DROP TABLE IF EXISTS organization_position_role CASCADE;
CREATE TABLE IF NOT EXISTS organization_position_role (
  id BIGSERIAL PRIMARY KEY,
  position_id BIGINT NOT NULL,
  role_id BIGINT NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_position_role ON organization_position_role (position_id, role_id);
CREATE INDEX IF NOT EXISTS idx_organization_position_role_position_id ON organization_position_role (position_id);
CREATE INDEX IF NOT EXISTS idx_organization_position_role_role_id ON organization_position_role (role_id);

INSERT INTO organization_position_role (position_id, role_id)
SELECT p.id, r.id
FROM organization_position p
JOIN permission_role r ON r.role_code = 'super_admin' AND r.deleted_at = 0
WHERE p.code = 'POS-CEO' AND p.deleted_at = 0
LIMIT 1
ON CONFLICT (position_id, role_id) DO NOTHING;

INSERT INTO organization_position_role (position_id, role_id)
SELECT p.id, r.id
FROM organization_position p
JOIN (
  SELECT 'POS-HR-MANAGER' AS position_code, 'organization_admin' AS role_code
  UNION ALL SELECT 'POS-OPS-SPECIALIST', 'organization_admin'
  UNION ALL SELECT 'POS-AUDITOR', 'auditor'
  UNION ALL SELECT 'POS-TECH-MANAGER', 'organization_admin'
  UNION ALL SELECT 'POS-BACKEND-ENGINEER', 'organization_admin'
  UNION ALL SELECT 'POS-PRODUCT-OPS-MANAGER', 'organization_admin'
  UNION ALL SELECT 'POS-USER-OPS', 'organization_admin'
  UNION ALL SELECT 'POS-RISK-SPECIALIST', 'auditor'
  UNION ALL SELECT 'POS-AUDIT-MANAGER', 'auditor'
) seed ON seed.position_code = p.code
JOIN permission_role r ON r.role_code = seed.role_code AND r.deleted_at = 0
WHERE p.deleted_at = 0
ON CONFLICT (position_id, role_id) DO NOTHING;
