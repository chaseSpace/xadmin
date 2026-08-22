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
JOIN (
  VALUES
    ('POS-CEO', 'super_admin'),
    ('POS-CEO', 'department_manager'),
    ('POS-HR-MANAGER', 'organization_admin'),
    ('POS-HR-MANAGER', 'department_manager'),
    ('POS-TECH-MANAGER', 'department_manager'),
    ('POS-BACKEND-ENGINEER', 'employee'),
    ('POS-PRODUCT-OPS-MANAGER', 'employee'),
    ('POS-PRODUCT-OPS-MANAGER', 'department_manager'),
    ('POS-USER-OPS', 'employee'),
    ('POS-RISK-SPECIALIST', 'auditor'),
    ('POS-AUDIT-MANAGER', 'auditor'),
    ('POS-AUDIT-MANAGER', 'department_manager')
) AS seed(position_code, role_code) ON seed.position_code = p.code
JOIN permission_role r ON r.role_code = seed.role_code AND r.deleted_at = 0
WHERE p.deleted_at = 0
ON CONFLICT (position_id, role_id) DO NOTHING;
