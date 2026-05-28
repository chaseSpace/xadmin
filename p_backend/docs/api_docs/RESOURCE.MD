# RESOURCE API

## 文件管理 API 列表
- `GET /v1/resource/files`
- `POST /v1/resource/files`
- `PUT /v1/resource/files/:id`
- `DELETE /v1/resource/files/:id`
- `POST /v1/resource/files/:id/access`
- `POST /v1/resource/files/check`

## GET /v1/resource/files
- 描述：文件管理列表，支持分页、类型筛选、有效性筛选与关键字搜索。
- Query：
  - `page_no`、`page_size`
  - `file_type`：`image` / `audio` / `video` / `document` / `archive`
  - `keyword`
  - `exists`：`true` / `false`，按 1 天内最近一次文件有效性检测结果筛选
- Response item：
  - `exists`：1 天内最近一次显式检测得到的文件存在状态；未检测或检测时间超过 1 天时返回 `null` 表示未知。
  - `exists_checked_at`：最近一次检测时间。
  - `require_auth`：文件访问是否需要登录鉴权。
  - `access_mode`：访问方式，`preview` 为预览模式，`download` 为直接下载。
- Response：
  - `total_size_bytes`：当前查询条件下所有文件的总大小，分页不影响该统计。

## POST /v1/resource/files
- 描述：上传文件并创建资源文件记录，底层复用 assets/assert 存储能力，并按二进制文件头识别真实文件类型。
- Content-Type：`multipart/form-data`
- Form：
  - `file`：必填
  - `file_type`：必填，`image` / `audio` / `video` / `document` / `archive`，必须与二进制识别结果一致
  - `name`：可选，默认使用上传文件名
  - `remark`：可选，最多 50 字
  - `require_auth`：可选，`true/false` 或 `1/0`，默认 `false`
  - `access_mode`：可选，`preview` / `download`，默认 `preview`
- 限制：
  - 图片：5M，支持 JPEG/PNG/GIF/WebP/AVIF
  - 语音：5M，支持 MP3/WAV/OGG/FLAC/M4A
  - 视频：100M，支持 MP4/MOV/WebM/AVI
  - 文档：20M，支持 PDF/Office/TXT/MD
  - 压缩包：100M，支持 ZIP/RAR/7Z/GZ

## PUT /v1/resource/files/:id
- 描述：更新文件类型、名称、备注、访问鉴权开关和访问方式。
- Body：
  - `file_type`：必填，资源文件类型枚举
  - `name`：必填
  - `remark`：可选，最多 50 字
  - `require_auth`：必填，布尔值
  - `access_mode`：必填，`1` 表示预览模式，`2` 表示直接下载

## DELETE /v1/resource/files/:id
- 描述：软删除文件记录。

## POST /v1/resource/files/:id/access
- 描述：返回文件地址、访问次数和上次访问时间；访问次数更新在后端异步执行。

## POST /v1/resource/files/check
- 描述：批量检测指定类型或全部文件是否仍存在，并更新列表中的 `exists` 与 `exists_checked_at`；后台同一时刻只允许一个检测任务运行。
- Query：
  - `file_type`：可选，`image` / `audio` / `video` / `document` / `archive`；不传时检测全部。
- Response：
  - `checked_count`：检测总数
  - `exists_count`：存在数量
  - `missing_count`：不存在数量

## GET /v1/assets/GetFile
- 描述：读取文件内容。普通资源公开访问；`require_auth=true` 的资源必须携带有效登录 Token。
