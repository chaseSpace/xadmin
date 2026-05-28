DROP TABLE IF EXISTS organization_position CASCADE;
CREATE TABLE IF NOT EXISTS organization_position (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(64) NOT NULL,
  code VARCHAR(64) NOT NULL,
  department_id BIGINT NOT NULL,
  level VARCHAR(32) NOT NULL,
  hc INTEGER NOT NULL DEFAULT 0,
  staffed INTEGER NOT NULL DEFAULT 0,
  status SMALLINT NOT NULL DEFAULT 1,
  sort INTEGER NOT NULL DEFAULT 0,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at BIGINT NOT NULL DEFAULT 0
);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_organization_position_code_deleted_at ON organization_position (code, deleted_at);
CREATE INDEX IF NOT EXISTS idx_organization_position_department_id ON organization_position (department_id);
CREATE INDEX IF NOT EXISTS idx_organization_position_status ON organization_position (status);
CREATE INDEX IF NOT EXISTS idx_organization_position_deleted_at ON organization_position (deleted_at);

INSERT INTO organization_position (name, code, department_id, level, hc, staffed, status, sort)
SELECT '总经理', 'POS-CEO', d.id, 'M3', 1, 1, 1, 10
FROM organization_department d
WHERE d.code = 'HQ' AND d.deleted_at = 0
ON CONFLICT (code, deleted_at) DO NOTHING;

INSERT INTO organization_position (name, code, department_id, level, hc, staffed, status, sort)
SELECT seed.name, seed.code, d.id, seed.level, seed.hc, seed.staffed, 1, seed.sort
FROM organization_department d
JOIN (
  SELECT '组织主管' AS name, 'POS-HR-MANAGER' AS code, 'M2' AS level, 2 AS hc, 1 AS staffed, 20 AS sort
  UNION ALL SELECT '运营专员', 'POS-OPS-SPECIALIST', 'P2', 8, 4, 30
  UNION ALL SELECT '审计专员', 'POS-AUDITOR', 'P2', 3, 1, 40
) seed ON TRUE
WHERE d.code = 'HQ' AND d.deleted_at = 0
ON CONFLICT (code, deleted_at) DO NOTHING;

INSERT INTO organization_position (name, code, department_id, level, hc, staffed, status, sort)
SELECT seed.name, seed.code, d.id, seed.level, seed.hc, seed.staffed, 1, seed.sort
FROM organization_department d
JOIN (
  SELECT 'TECH' AS department_code, '研发经理' AS name, 'POS-TECH-MANAGER' AS code, 'M2' AS level, 2 AS hc, 1 AS staffed, 110 AS sort
  UNION ALL SELECT 'TECH', '后端工程师', 'POS-BACKEND-ENGINEER', 'P3', 6, 3, 120
  UNION ALL SELECT 'PRODUCT_OPS', '产品运营经理', 'POS-PRODUCT-OPS-MANAGER', 'M2', 1, 1, 210
  UNION ALL SELECT 'PRODUCT_OPS', '用户运营', 'POS-USER-OPS', 'P2', 5, 2, 220
  UNION ALL SELECT 'RISK_AUDIT', '风控专员', 'POS-RISK-SPECIALIST', 'P2', 3, 1, 310
  UNION ALL SELECT 'RISK_AUDIT', '审计经理', 'POS-AUDIT-MANAGER', 'M2', 1, 1, 320
) seed ON seed.department_code = d.code
WHERE d.deleted_at = 0
ON CONFLICT (code, deleted_at) DO NOTHING;
