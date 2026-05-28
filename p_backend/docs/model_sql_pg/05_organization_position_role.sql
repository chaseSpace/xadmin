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
JOIN permission_role r ON r.role_name = '超级管理员' AND r.deleted_at = 0
WHERE p.code = 'POS-CEO' AND p.deleted_at = 0
LIMIT 1
ON CONFLICT (position_id, role_id) DO NOTHING;

INSERT INTO organization_position_role (position_id, role_id)
SELECT p.id, r.id
FROM organization_position p
JOIN (
  SELECT 'POS-HR-MANAGER' AS position_code, '组织管理员' AS role_name
  UNION ALL SELECT 'POS-OPS-SPECIALIST', '组织管理员'
  UNION ALL SELECT 'POS-AUDITOR', '审计员'
  UNION ALL SELECT 'POS-TECH-MANAGER', '组织管理员'
  UNION ALL SELECT 'POS-BACKEND-ENGINEER', '组织管理员'
  UNION ALL SELECT 'POS-PRODUCT-OPS-MANAGER', '组织管理员'
  UNION ALL SELECT 'POS-USER-OPS', '组织管理员'
  UNION ALL SELECT 'POS-RISK-SPECIALIST', '审计员'
  UNION ALL SELECT 'POS-AUDIT-MANAGER', '审计员'
) seed ON seed.position_code = p.code
JOIN permission_role r ON r.role_name = seed.role_name AND r.deleted_at = 0
WHERE p.deleted_at = 0
ON CONFLICT (position_id, role_id) DO NOTHING;
