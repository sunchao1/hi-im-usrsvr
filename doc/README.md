# hi-im-usrsvr 文档

> **hi-im-usrsvr** 是 hi-im 生态 **L3 用户与会话服务**（Gin HTTP + hubclient BACKEND）；对应必嗨 usrsvr。  
> **状态**：**M4 可联调**（register + ONLINE/OFFLINE/PING + Redis）  
> **作者**：sunchao1 · **许可证**：Apache License 2.0（见仓库根目录 `LICENSE`）

---

## 阅读顺序

| 顺序 | 文档 | 内容 |
|------|------|------|
| 1 | [技术设计文档.md](技术设计文档.md) | 定位、HTTP/Hub 双平面、Redis、M4/M5/M6 边界 |
| 2 | [M1-实施清单.md](M1-实施清单.md) | 生态 **M4** 任务拆解（与 hi-im-seqsvr 并行） |

---

## 生态对照

| 文档 | 说明 |
|------|------|
| [hi-im/doc/hi-im-档C技术方案设计.md](https://github.com/sunchao1/hi-im/blob/main/doc/hi-im-档C技术方案设计.md) | 生态总方案 §6 gRPC、§7 Gin、§11 M4～M6 |
| [hi-im-api/doc/技术设计文档.md](https://github.com/sunchao1/hi-im-api/blob/main/doc/技术设计文档.md) | IM 头、proto、rediskey、errno |
| [hi-im-seqsvr/doc/技术设计文档.md](https://github.com/sunchao1/hi-im-seqsvr/blob/main/doc/技术设计文档.md) | gRPC 发号 |
| [hi-im-hubclient/doc/技术设计文档.md](https://github.com/sunchao1/hi-im-hubclient/blob/main/doc/技术设计文档.md) | BACKEND AsyncSend / RegisterHandler |
| beehive-im `src/golang/exec/usrsvr/` | 必嗨实现对照 |

---

## 角色对照

```text
档 C 总方案     →  hi-im/doc/hi-im-档C技术方案设计.md
契约            →  hi-im-api（header / proto / rediskey / httpx）
发号            →  hi-im-seqsvr（gRPC client）
Hub 传输        →  hi-im-hubclient（BACKEND 平面）
用户/会话/群元数据 →  hi-im-usrsvr（本仓库）
群聊 fan-out    →  hi-im-msgsvr（M6，非 usrsvr）
WS 接入         →  hi-im-gateway（M5）
```
