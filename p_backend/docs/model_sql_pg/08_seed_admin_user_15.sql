-- 目的：插入 15 条随机后台用户测试数据（可重复执行）
-- 说明：默认密码哈希与 admin 保持一致，便于本地联调
-- 密码明文（联调用）：123456

INSERT INTO admin_user (
  uid, username, password_hash, display_name, avatar, email, phone, status, last_login_at, last_login_ip
) VALUES
  (21001, 'ceo_user_21001', '$2a$10$E7izPmKoa8FVP4fkae4QjOZ5SEs4Tg4LBf2FVbFb8GqTYZC0lEGim', '周锐', '', 'zhou.rui21001@example.com', '13800021001', 1, '2026-04-21 08:31:10', '10.22.1.11'),
  (21002, 'hr_manager_21002', '$2a$10$E7izPmKoa8FVP4fkae4QjOZ5SEs4Tg4LBf2FVbFb8GqTYZC0lEGim', '林嘉', '', 'lin.jia21002@example.com', '13800021002', 0, '2026-04-20 19:22:47', '10.22.1.12'),
  (21003, 'ops_specialist_21003', '$2a$10$E7izPmKoa8FVP4fkae4QjOZ5SEs4Tg4LBf2FVbFb8GqTYZC0lEGim', '谢宁', '', 'xie.ning21003@example.com', '13800021003', 1, '2026-04-20 10:05:23', '10.22.1.13'),
  (21004, 'auditor_21004', '$2a$10$E7izPmKoa8FVP4fkae4QjOZ5SEs4Tg4LBf2FVbFb8GqTYZC0lEGim', '高晨', '', 'gao.chen21004@example.com', '13800021004', 2, '2026-04-19 14:41:56', '10.22.1.14'),
  (21005, 'hr_manager_21005', '$2a$10$E7izPmKoa8FVP4fkae4QjOZ5SEs4Tg4LBf2FVbFb8GqTYZC0lEGim', '宋扬', '', 'song.yang21005@example.com', '13800021005', 1, '2026-04-19 09:16:03', '10.22.1.15'),
  (21006, 'ops_specialist_21006', '$2a$10$E7izPmKoa8FVP4fkae4QjOZ5SEs4Tg4LBf2FVbFb8GqTYZC0lEGim', '唐可', '', 'tang.ke21006@example.com', '13800021006', 0, '2026-04-18 21:10:41', '10.22.1.16'),
  (21007, 'auditor_21007', '$2a$10$E7izPmKoa8FVP4fkae4QjOZ5SEs4Tg4LBf2FVbFb8GqTYZC0lEGim', '温航', '', 'wen.hang21007@example.com', '13800021007', 1, '2026-04-18 11:52:18', '10.22.1.17'),
  (21008, 'hr_manager_21008', '$2a$10$E7izPmKoa8FVP4fkae4QjOZ5SEs4Tg4LBf2FVbFb8GqTYZC0lEGim', '彭越', '', 'peng.yue21008@example.com', '13800021008', 1, '2026-04-17 20:47:30', '10.22.1.18'),
  (21009, 'ops_specialist_21009', '$2a$10$E7izPmKoa8FVP4fkae4QjOZ5SEs4Tg4LBf2FVbFb8GqTYZC0lEGim', '蒋尧', '', 'jiang.yao21009@example.com', '13800021009', 2, '2026-04-17 09:03:22', '10.22.1.19'),
  (21010, 'auditor_21010', '$2a$10$E7izPmKoa8FVP4fkae4QjOZ5SEs4Tg4LBf2FVbFb8GqTYZC0lEGim', '邵凡', '', 'shao.fan21010@example.com', '13800021010', 1, '2026-04-16 16:33:14', '10.22.1.20'),
  (21011, 'hr_manager_21011', '$2a$10$E7izPmKoa8FVP4fkae4QjOZ5SEs4Tg4LBf2FVbFb8GqTYZC0lEGim', '顾真', '', 'gu.zhen21011@example.com', '13800021011', 0, '2026-04-16 08:45:50', '10.22.1.21'),
  (21012, 'ops_specialist_21012', '$2a$10$E7izPmKoa8FVP4fkae4QjOZ5SEs4Tg4LBf2FVbFb8GqTYZC0lEGim', '易澄', '', 'yi.cheng21012@example.com', '13800021012', 1, '2026-04-15 22:14:08', '10.22.1.22'),
  (21013, 'auditor_21013', '$2a$10$E7izPmKoa8FVP4fkae4QjOZ5SEs4Tg4LBf2FVbFb8GqTYZC0lEGim', '贺晨', '', 'he.chen21013@example.com', '13800021013', 1, '2026-04-15 11:27:35', '10.22.1.23'),
  (21014, 'hr_manager_21014', '$2a$10$E7izPmKoa8FVP4fkae4QjOZ5SEs4Tg4LBf2FVbFb8GqTYZC0lEGim', '许然', '', 'xu.ran21014@example.com', '13800021014', 2, '2026-04-14 18:59:57', '10.22.1.24'),
  (21015, 'ops_specialist_21015', '$2a$10$E7izPmKoa8FVP4fkae4QjOZ5SEs4Tg4LBf2FVbFb8GqTYZC0lEGim', '戴星', '', 'dai.xing21015@example.com', '13800021015', 1, '2026-04-14 07:12:06', '10.22.1.25')
ON CONFLICT (uid) DO UPDATE SET
  password_hash = EXCLUDED.password_hash,
  username = EXCLUDED.username,
  display_name = EXCLUDED.display_name,
  avatar = EXCLUDED.avatar,
  email = EXCLUDED.email,
  phone = EXCLUDED.phone,
  status = EXCLUDED.status,
  last_login_at = EXCLUDED.last_login_at,
  last_login_ip = EXCLUDED.last_login_ip,
  updated_at = CURRENT_TIMESTAMP;

UPDATE admin_user u
SET department_id = d.id,
    position_id = p.id,
    updated_at = CURRENT_TIMESTAMP
FROM (
  SELECT 21001 AS uid, 'HQ' AS department_code, 'POS-CEO' AS position_code
  UNION ALL SELECT 21002, 'HQ', 'POS-HR-MANAGER'
  UNION ALL SELECT 21003, 'HQ', 'POS-OPS-SPECIALIST'
  UNION ALL SELECT 21004, 'HQ', 'POS-AUDITOR'
  UNION ALL SELECT 21005, 'TECH', 'POS-TECH-MANAGER'
  UNION ALL SELECT 21006, 'TECH', 'POS-BACKEND-ENGINEER'
  UNION ALL SELECT 21007, 'PRODUCT_OPS', 'POS-PRODUCT-OPS-MANAGER'
  UNION ALL SELECT 21008, 'PRODUCT_OPS', 'POS-USER-OPS'
  UNION ALL SELECT 21009, 'RISK_AUDIT', 'POS-RISK-SPECIALIST'
  UNION ALL SELECT 21010, 'RISK_AUDIT', 'POS-AUDIT-MANAGER'
  UNION ALL SELECT 21011, 'HQ', 'POS-HR-MANAGER'
  UNION ALL SELECT 21012, 'TECH', 'POS-BACKEND-ENGINEER'
  UNION ALL SELECT 21013, 'PRODUCT_OPS', 'POS-USER-OPS'
  UNION ALL SELECT 21014, 'RISK_AUDIT', 'POS-RISK-SPECIALIST'
  UNION ALL SELECT 21015, 'TECH', 'POS-TECH-MANAGER'
) mapping
JOIN organization_department d ON d.code = mapping.department_code AND d.deleted_at = 0
JOIN organization_position p ON p.department_id = d.id AND p.code = mapping.position_code AND p.deleted_at = 0
WHERE u.uid = mapping.uid
  AND u.uid BETWEEN 21001 AND 21015
  AND u.deleted_at = 0;
