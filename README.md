# FlowerLottery Backend

花愿奇遇活动后端。使用 Go 1.24+、Gin、GORM、MySQL 8、Viper、Zap、JWT 和 bcrypt；用户端与管理后台统一位于同级 `FlowerLotteryFrontend` 项目。

## 目录

- `cmd/server`：API 服务入口
- `cmd/seed`：活动、奖池、概率、用户和管理员初始化
- `controller/service/repository/model`：企业分层业务代码
- `docs/schema.sql`：MySQL 8 建表脚本

## 数据库启动

1. 安装 MySQL 8，并创建可访问账号。
2. 导入建表脚本：

```bash
mysql -uroot -p < docs/schema.sql
```

3. 修改 `config.yaml` 中的数据库 DSN 和 JWT secret。
4. 初始化活动数据：

```bash
go run ./cmd/seed
```

开发账号：

- H5 用户：`demo / 123456`
- 管理员：`admin / 123456`

## Backend 启动

```bash
go mod tidy
go run ./cmd/server
```

默认地址：`http://127.0.0.1:8080`，健康检查：`GET /api/v1/health`。

## Frontend 启动

```bash
cd ../FlowerLotteryFrontend
npm install
npm run dev
```

Vite 开发代理将 `/api` 转发至 Backend `8080` 端口。活动页位于 `/`，管理后台登录页位于 `/admin/login`。

生产构建：

```bash
npm run build
```

## 验证

```bash
go test ./...
cd ../FlowerLotteryFrontend && npm run build
```

核心规则以需求文档为准：六档金币兑换花瓣、白昼/星夜 1/10/30 抽、概率与金币保底点亮、18 朵提前停止退款、实际花瓣消耗进入榜单。
## 用户头像存储

- 上传接口：注册使用 `POST /api/v1/auth/register` 的 multipart 表单可选字段 `avatar`；登录用户使用 `POST /api/v1/me/avatar`，文件字段同为 `avatar`。
- 仅接受 JPEG/PNG，上传文件最大 8MB，解码后最大 4000 万像素。
- Backend 居中裁剪并生成 512×512、质量 85 的 JPEG，不保留用户原图和原始文件名。
- 默认存储目录为 `storage/uploads/avatars`，数据库保存 `/uploads/avatars/{随机名}.jpg`。
- `storage/uploads` 必须具备写权限；生产部署必须挂载持久化磁盘，避免容器重启后头像丢失。
- Backend 直接提供 `/uploads` 静态文件服务；生产网关需同时转发 `/api` 和 `/uploads`。
- 更换头像成功后会删除旧的本地头像文件；历史外链头像不会被删除。
