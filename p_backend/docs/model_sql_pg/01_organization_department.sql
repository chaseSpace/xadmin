DROP TABLE IF EXISTS organization_department CASCADE;
CREATE TABLE IF NOT EXISTS organization_department (
  id BIGSERIAL PRIMARY KEY,
  parent_id BIGINT NOT NULL DEFAULT 0,
  name VARCHAR(64) NOT NULL,
  code VARCHAR(64) NOT NULL,
  status SMALLINT NOT NULL DEFAULT 1,
  member_count INTEGER NOT NULL DEFAULT 0,
  sort INTEGER NOT NULL DEFAULT 0,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at BIGINT NOT NULL DEFAULT 0
);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_organization_department_code_deleted_at ON organization_department (code, deleted_at);
CREATE INDEX IF NOT EXISTS idx_organization_department_parent_id ON organization_department (parent_id);
CREATE INDEX IF NOT EXISTS idx_organization_department_deleted_at ON organization_department (deleted_at);

INSERT INTO organization_department (parent_id, name, code, status, member_count, sort)
VALUES (0, '集团总部', 'HQ', 1, 0, 10)
ON CONFLICT (code, deleted_at) DO NOTHING;

INSERT INTO organization_department (parent_id, name, code, status, member_count, sort)
SELECT hq.id, seed.name, seed.code, 1, seed.member_count, seed.sort
FROM organization_department hq
JOIN (
  SELECT '技术研发部' AS name, 'TECH' AS code, 12 AS member_count, 20 AS sort
  UNION ALL SELECT '产品运营部', 'PRODUCT_OPS', 9, 30
  UNION ALL SELECT '风控审计部', 'RISK_AUDIT', 5, 40
  UNION ALL SELECT '人力资源部', 'HR', 8, 50
  UNION ALL SELECT '财务管理部', 'FINANCE', 6, 60
  UNION ALL SELECT '市场品牌部', 'MARKETING', 10, 70
  UNION ALL SELECT '客户成功部', 'CUSTOMER_SUCCESS', 11, 80
  UNION ALL SELECT '法务合规部', 'LEGAL', 4, 90
  UNION ALL SELECT '数据智能部', 'DATA_AI', 13, 100
  UNION ALL SELECT '行政采购部', 'ADMIN_PROCUREMENT', 7, 110
) seed ON TRUE
WHERE hq.code = 'HQ' AND hq.deleted_at = 0
ON CONFLICT (code, deleted_at) DO NOTHING;

INSERT INTO organization_department (parent_id, name, code, status, member_count, sort)
SELECT parent.id, seed.name, seed.code, 1, seed.member_count, seed.sort
FROM organization_department parent
JOIN (
  SELECT 'TECH' AS parent_code, '平台架构组' AS name, 'TECH_PLATFORM' AS code, 5 AS member_count, 101 AS sort
  UNION ALL SELECT 'TECH', '后端服务组', 'TECH_BACKEND', 8, 102
  UNION ALL SELECT 'TECH', '前端体验组', 'TECH_FRONTEND', 6, 103
  UNION ALL SELECT 'TECH', '测试保障组', 'TECH_QA', 4, 104
  UNION ALL SELECT 'TECH', '基础设施组', 'TECH_INFRA', 3, 105
  UNION ALL SELECT 'TECH', '数据平台组', 'TECH_DATA', 5, 106
  UNION ALL SELECT 'PRODUCT_OPS', '产品策划组', 'PRODUCT_OPS_PLANNING', 4, 201
  UNION ALL SELECT 'PRODUCT_OPS', '用户运营组', 'PRODUCT_OPS_USER', 3, 202
  UNION ALL SELECT 'PRODUCT_OPS', '增长运营组', 'PRODUCT_OPS_GROWTH', 2, 203
  UNION ALL SELECT 'RISK_AUDIT', '风险策略组', 'RISK_AUDIT_POLICY', 2, 301
  UNION ALL SELECT 'RISK_AUDIT', '内部审计组', 'RISK_AUDIT_INTERNAL', 3, 302
  UNION ALL SELECT 'HR', '招聘组', 'HR_RECRUITING', 3, 401
  UNION ALL SELECT 'HR', '组织发展组', 'HR_OD', 2, 402
  UNION ALL SELECT 'HR', '员工关系组', 'HR_ER', 2, 403
  UNION ALL SELECT 'HR', '薪酬绩效组', 'HR_CNB', 1, 404
  UNION ALL SELECT 'FINANCE', '财务核算组', 'FINANCE_ACCOUNTING', 2, 501
  UNION ALL SELECT 'FINANCE', '资金管理组', 'FINANCE_TREASURY', 2, 502
  UNION ALL SELECT 'MARKETING', '品牌传播组', 'MARKETING_BRAND', 3, 601
  UNION ALL SELECT 'MARKETING', '内容营销组', 'MARKETING_CONTENT', 3, 602
  UNION ALL SELECT 'MARKETING', '活动策划组', 'MARKETING_EVENT', 2, 603
  UNION ALL SELECT 'CUSTOMER_SUCCESS', '客户服务组', 'CS_SUPPORT', 4, 701
  UNION ALL SELECT 'CUSTOMER_SUCCESS', '实施交付组', 'CS_DELIVERY', 3, 702
  UNION ALL SELECT 'CUSTOMER_SUCCESS', '续费增长组', 'CS_RENEWAL', 2, 703
  UNION ALL SELECT 'LEGAL', '合同法务组', 'LEGAL_CONTRACT', 2, 801
  UNION ALL SELECT 'DATA_AI', '商业分析组', 'DATA_AI_BI', 4, 901
  UNION ALL SELECT 'DATA_AI', '算法应用组', 'DATA_AI_MODEL', 5, 902
  UNION ALL SELECT 'DATA_AI', '数据治理组', 'DATA_AI_GOV', 3, 903
  UNION ALL SELECT 'ADMIN_PROCUREMENT', '行政服务组', 'ADMIN_SERVICE', 2, 1001
  UNION ALL SELECT 'ADMIN_PROCUREMENT', '采购管理组', 'ADMIN_PROCUREMENT_BUY', 3, 1002
) seed ON seed.parent_code = parent.code
WHERE parent.deleted_at = 0
ON CONFLICT (code, deleted_at) DO NOTHING;
