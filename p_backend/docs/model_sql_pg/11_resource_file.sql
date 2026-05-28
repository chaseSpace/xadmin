DROP TABLE IF EXISTS resource_file CASCADE;
CREATE TABLE IF NOT EXISTS resource_file (
  id BIGSERIAL PRIMARY KEY,
  file_type VARCHAR(32) NOT NULL,
  name VARCHAR(160) NOT NULL,
  file_url VARCHAR(500) NOT NULL,
  mime_type VARCHAR(120) NOT NULL DEFAULT '',
  extension VARCHAR(16) NOT NULL DEFAULT '',
  size_bytes BIGINT NOT NULL DEFAULT 0,
  remark VARCHAR(50) NOT NULL DEFAULT '',
  require_auth BOOLEAN NOT NULL DEFAULT FALSE,
  access_mode VARCHAR(16) NOT NULL DEFAULT 'preview',
  "exists" BOOLEAN NOT NULL DEFAULT TRUE,
  exists_checked_at TIMESTAMP NULL,
  last_access_at TIMESTAMP NULL,
  access_count INTEGER NOT NULL DEFAULT 0,
  creator_uid INTEGER NOT NULL DEFAULT 0,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at BIGINT NOT NULL DEFAULT 0
);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_resource_file_url_deleted_at ON resource_file (file_url, deleted_at);
CREATE INDEX IF NOT EXISTS idx_resource_file_type_deleted_at ON resource_file (file_type, deleted_at);
CREATE INDEX IF NOT EXISTS idx_resource_file_exists_checked_at ON resource_file ("exists", exists_checked_at, deleted_at);
CREATE INDEX IF NOT EXISTS idx_resource_file_created_at ON resource_file (created_at);
CREATE INDEX IF NOT EXISTS idx_resource_file_deleted_at ON resource_file (deleted_at);

INSERT INTO resource_file (
  file_type, name, file_url, mime_type, extension, size_bytes, remark,
  require_auth, access_mode, "exists", exists_checked_at, access_count, creator_uid, deleted_at
)
SELECT seed.file_type, seed.name, seed.file_url, seed.mime_type, seed.extension, seed.size_bytes, seed.remark,
       seed.require_auth, seed.access_mode, seed."exists", seed.exists_checked_at, seed.access_count, seed.creator_uid, 0
FROM (
  SELECT 'image' AS file_type, '示范图片-产品封面.jpg' AS name, '/demo/resource/images/product-cover.jpg' AS file_url, 'image/jpeg' AS mime_type, '.jpg' AS extension, 245760 AS size_bytes, '示范图片' AS remark, FALSE AS require_auth, 'preview' AS access_mode, TRUE AS "exists", NOW() AS exists_checked_at, 12 AS access_count, 1 AS creator_uid
  UNION ALL SELECT 'image', '示范图片-活动海报.png', '/demo/resource/images/event-poster.png', 'image/png', '.png', 518144, '示范图片', FALSE, 'preview', TRUE, NOW(), 8, 1
  UNION ALL SELECT 'image', '示范图片-头像素材.webp', '/demo/resource/images/avatar-sample.webp', 'image/webp', '.webp', 98304, '示范图片', TRUE, 'download', TRUE, NOW(), 21, 1
  UNION ALL SELECT 'audio', '示范音频-欢迎语.mp3', '/demo/resource/audio/welcome.mp3', 'audio/mpeg', '.mp3', 1048576, '示范音频', FALSE, 'preview', TRUE, NOW(), 5, 1
  UNION ALL SELECT 'audio', '示范音频-系统提示.wav', '/demo/resource/audio/system-prompt.wav', 'audio/wav', '.wav', 2097152, '示范音频', TRUE, 'download', FALSE, NOW(), 2, 1
  UNION ALL SELECT 'audio', '示范音频-课程片段.m4a', '/demo/resource/audio/course-preview.m4a', 'audio/mp4', '.m4a', 1572864, '示范音频', FALSE, 'preview', TRUE, NOW() - INTERVAL '4 day', 7, 1
  UNION ALL SELECT 'video', '示范视频-功能介绍.mp4', '/demo/resource/video/feature-intro.mp4', 'video/mp4', '.mp4', 12582912, '示范视频', FALSE, 'preview', TRUE, NOW(), 16, 1
  UNION ALL SELECT 'video', '示范视频-操作演示.webm', '/demo/resource/video/operation-demo.webm', 'video/webm', '.webm', 9437184, '示范视频', TRUE, 'download', TRUE, NOW(), 4, 1
  UNION ALL SELECT 'video', '示范视频-历史素材.mov', '/demo/resource/video/archive.mov', 'video/quicktime', '.mov', 18874368, '示范视频', FALSE, 'download', FALSE, NOW() - INTERVAL '5 day', 1, 1
  UNION ALL SELECT 'document', '示范文档-产品说明.pdf', '/demo/resource/docs/product-guide.pdf', 'application/pdf', '.pdf', 786432, '示范文档', FALSE, 'preview', TRUE, NOW(), 9, 1
  UNION ALL SELECT 'document', '示范文档-运营日报.docx', '/demo/resource/docs/operation-report.docx', 'application/zip', '.docx', 393216, '示范文档', TRUE, 'download', TRUE, NOW(), 3, 1
  UNION ALL SELECT 'document', '示范文档-导入模板.xlsx', '/demo/resource/docs/import-template.xlsx', 'application/zip', '.xlsx', 262144, '示范文档', FALSE, 'download', TRUE, NOW() - INTERVAL '4 day', 6, 1
) seed
ON CONFLICT (file_url, deleted_at) DO NOTHING;
