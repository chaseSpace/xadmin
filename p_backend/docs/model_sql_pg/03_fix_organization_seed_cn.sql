UPDATE organization_department
SET name = '集团总部'
WHERE code = 'HQ' AND deleted_at = 0;

UPDATE organization_position
SET name = '总经理'
WHERE code = 'POS-CEO' AND deleted_at = 0;
