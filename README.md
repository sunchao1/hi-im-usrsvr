# hi-im-usrsvr

hi-im 生态 **用户注册 + 会话 ONLINE** 服务（Gin HTTP + hubclient BACKEND）。契约定义见 [hi-im-api](https://github.com/sunchao1/hi-im-api)。

**作者**：sunchao1 · **许可证**：Apache License 2.0

## 依赖

- **hi-im-api** v0.1.0 — header、proto、rediskey、errno
- **hi-im-hubclient** v0.1.0 — Hub BACKEND 平面
- **hi-im-seqsvr** v0.1.0 — gRPC `AllocSid` / `AllocSeq`
- **Redis** — 会话在线态

## 快速开始

```bash
export HIIM_BACKEND_ADDR="127.0.0.1:28889"
export HIIM_REDIS_ADDR="127.0.0.1:6379"
export HIIM_SEQSVR_ADDR="127.0.0.1:50051"
export HIIM_NID="30001"
export HIIM_AUTH_USER="proxy"
export HIIM_AUTH_PASS="proxy"
export HIIM_SUB_CMDS="0x0101,0x0103,0x0105"

make build && ./bin/usrsvr
```

注册示例：

```bash
curl 'http://127.0.0.1:8081/im/register?uid=1&nation=86'
```

## 环境变量

| 变量 | 默认 | 说明 |
|------|------|------|
| `HIIM_HTTP_LISTEN` | `:8081` | Gin HTTP |
| `HIIM_BACKEND_ADDR` | — | Hub BACKEND（必填） |
| `HIIM_NID` | `30001` | 本进程 NID |
| `HIIM_AUTH_USER` / `HIIM_AUTH_PASS` | `proxy` | Hub 认证 |
| `HIIM_SUB_CMDS` | `0x0101,0x0103,0x0105` | SUB 命令集 |
| `HIIM_REDIS_ADDR` | `127.0.0.1:6379` | Redis |
| `HIIM_SEQSVR_ADDR` | `127.0.0.1:50051` | seqsvr gRPC |
| `HIIM_M4_SKIP_TOKEN` | `true` | M4 跳过 token 校验 |
| `HIIM_LOG_LEVEL` | `info` | slog 级别 |

## 健康检查

- `GET /healthz` — 进程存活
- `GET /readyz` — Redis Ping + hubclient Ready

## 测试

```bash
make test
make test-integration   # 可选，含 //go:build integration
```

## Docker

```bash
make docker
docker run --rm -p 8081:8081 \
  -e HIIM_BACKEND_ADDR=hub:28889 \
  -e HIIM_REDIS_ADDR=redis:6379 \
  -e HIIM_SEQSVR_ADDR=hi-im-seqsvr:50051 \
  hi-im-usrsvr:latest
```

## Compose 对接（hi-im 主仓 M4+）

主仓 `profile biz` 将加入 `hi-im-usrsvr` service：

```yaml
hi-im-usrsvr:
  image: hi-im-usrsvr:latest
  depends_on:
    hub:
      condition: service_healthy
    redis:
      condition: service_healthy
    hi-im-seqsvr:
      condition: service_started
  environment:
    HIIM_BACKEND_ADDR: "hub:28889"
    HIIM_REDIS_ADDR: "redis:6379"
    HIIM_SEQSVR_ADDR: "hi-im-seqsvr:50051"
    HIIM_NID: "30001"
    HIIM_SUB_CMDS: "0x0101,0x0103,0x0105"
  ports:
    - "8081:8081"
```

## 文档

- [技术设计文档](doc/技术设计文档.md)
- [M1 实施清单（M4）](doc/M1-实施清单.md)

## License

Copyright 2026 sunchao1 · Apache License 2.0（见 [LICENSE](LICENSE)）
