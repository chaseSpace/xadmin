DROP TABLE IF EXISTS organization_department CASCADE;
CREATE TABLE IF NOT EXISTS organization_department (
  id BIGSERIAL PRIMARY KEY,
  parent_id BIGINT NOT NULL DEFAULT 0,
  name VARCHAR(64) NOT NULL,
  code VARCHAR(64) NOT NULL,
  status SMALLINT NOT NULL DEFAULT 1,
  sort INTEGER NOT NULL DEFAULT 0,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at BIGINT NOT NULL DEFAULT 0
);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_organization_department_code_deleted_at ON organization_department (code, deleted_at);
CREATE INDEX IF NOT EXISTS idx_organization_department_parent_id ON organization_department (parent_id);
CREATE INDEX IF NOT EXISTS idx_organization_department_deleted_at ON organization_department (deleted_at);

INSERT INTO organization_department (parent_id, name, code, status, sort)
VALUES (0, '集团总部', 'HQ', 1, 10)
ON CONFLICT (code, deleted_at) DO NOTHING;

INSERT INTO organization_department (parent_id, name, code, status, sort)
SELECT hq.id, seed.name, seed.code, 1, seed.sort
FROM organization_department hq
JOIN (
  SELECT '技术研发部' AS name, 'TECH' AS code, 20 AS sort
  UNION ALL SELECT '产品运营部', 'PRODUCT_OPS', 30
  UNION ALL SELECT '风控审计部', 'RISK_AUDIT', 40
  UNION ALL SELECT '人力资源部', 'HR', 50
  UNION ALL SELECT '财务管理部', 'FINANCE', 60
  UNION ALL SELECT '市场品牌部', 'MARKETING', 70
  UNION ALL SELECT '客户成功部', 'CUSTOMER_SUCCESS', 80
  UNION ALL SELECT '法务合规部', 'LEGAL', 90
  UNION ALL SELECT '数据智能部', 'DATA_AI', 100
  UNION ALL SELECT '行政采购部', 'ADMIN_PROCUREMENT', 110
) seed ON TRUE
WHERE hq.code = 'HQ' AND hq.deleted_at = 0
ON CONFLICT (code, deleted_at) DO NOTHING;
