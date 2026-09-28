# 缓存策略方案（Cache Policy）

本文回答一个问题：**配置了 tile cache 之后，什么时候用缓存、什么时候回源？**
并提供可直接使用的配置与行为契约。

## 一、背景与现状

瓦片读取链路为 `TileManager.LoadTileCoord`：

```
请求 ──▶ cache.LoadTiles（读本地缓存文件）
     ──▶ IsCached(coord) 判断是否可用
     ──▶ 未可用 ──▶ TileCreator ──▶ 源（WMS/tile/mapbox/…）拉取 ──▶ 写回缓存
```

改造前的缺口：

- `IsCached` 里已经存在一个全局阈值 `ExpireTimestamp`（"该时刻之前写入的
  瓦片视为过期"），但**没有任何配置能够设置它**，形同虚设；
- 没有 per-tile 的 TTL 语义；
- 没有"命中即用、永不回源"或"绕过缓存"的显式模式；
- 缓存过期后只能同步回源，无法利用旧值兜底。

## 二、策略模式

每个 cache（`caches.<name>.cache_policy`）可配置以下四种模式之一：

| 模式 | 行为 | 典型场景 |
|---|---|---|
| `cache_first`（默认） | 命中即用，**永不回源**；只有缓存未命中才拉源。除非 clean（`RemoveTileCoords`/清理目录）或设置全局重刷阈值，缓存内容永不过期 | 底图数据基本不变（历史影像、行政区划底图）；离线打包场景 |
| `ttl` | 命中且未过期（文件 mtime + `ttl_hours`）即用；**过期后下一次请求同步回源刷新** | 数据按天/小时更新（每日卫星影像、定期更新的矢量底图） |
| `revalidate` | 同 `ttl` 判定过期，但**立即返回旧值，同时后台刷新缓存**；下一次请求拿到新值 | 对延迟敏感的在线地图；上游偶尔抖动也不能拖慢响应 |
| `bypass` | **不读也不写缓存**，每次请求直连源 | 调试上游、源数据实时变化且不值得缓存 |

补充机制（与策略正交，对所有模式生效）：

- **全局重刷阈值**：`TileManager.SetExpireTimestamp(t)` —— 写入时间早于 `t`
  的缓存瓦片一律视为过期（运维"全量重刷"入口）。
- **刷新互斥**：过期回源/后台刷新都经过 tile locker（内存或文件锁），
  并发请求同一瓦片只会触发一次源拉取（防缓存击穿）。
- **后台刷新兜底**：`revalidate` 的后台刷新失败时静默保留旧值，60s 超时，
  不会泄漏 goroutine。

## 三、配置格式

```json
{
  "caches": {
    "base_cache": {
      "grid": "g1",
      "sources": ["s1"],
      "format": "png",
      "cache": {"directory_layout": "tms", "directory": "/data/cache/base"},
      "tile_options": {"type": "image", "format": "png"},
      "cache_policy": {
        "mode": "ttl",
        "ttl_hours": 24
      }
    }
  }
}
```

- `mode`：`cache_first` | `ttl` | `revalidate` | `bypass`；缺省 `cache_first`
  （与历史行为完全兼容，存量配置无需改动）。
- `ttl_hours`：`ttl` / `revalidate` 必填且 > 0，其他模式忽略。
- 配置校验：`POST /services/{id}/proxy/validate` 会对 `cache_policy` 的模式
  与 ttl 组合做合法性检查（未知模式、ttl 缺失都会报字段级错误）。

## 四、行为矩阵

| 场景 | cache_first | ttl | revalidate | bypass |
|---|---|---|---|---|
| 缓存未命中 | 回源+写入 | 回源+写入 | 回源+写入 | 回源，不写 |
| 缓存命中、未过期 | 用缓存 | 用缓存 | 用缓存 | 回源 |
| 缓存命中、已过期（mtime+ttl） | 用缓存（永不过期） | 同步回源刷新 | 返回旧值+后台刷新 | 回源 |
| 早于全局重刷阈值 | 视为过期 | 视为过期 | 视为过期（旧值兜底） | 回源 |
| Clean / RemoveTileCoords | 之后重新回源 | 同左 | 同左 | 无影响 |

## 五、实现位置

| 组件 | 位置 |
|---|---|
| 策略模型与解析 | `cache/policy.go`（`CachePolicy`、`ParseCachePolicy`） |
| 决策点 | `cache/tile_manager.go` `IsCached`（策略判定）、`loadTileCoords`（revalidate 后台刷新、bypass 跳过读）、`ShouldStoreToCache`（bypass 不写） |
| 写缓存拦截 | `cache/creator.go`（`shouldStoreFor`） |
| 配置字段 | `setting/spec.go` `CacheSource.CachePolicy` |
| 装配 | `setting/factory.go` `PreLoadCacheManager` |
| 配置校验 | `setting/validation.go` `validateCachesField` |
| 测试 | `cache/policy_test.go`（四种模式的回源计数 + revalidate 后台刷新验证） |

## 六、选型建议

- 底图稳定、追求极致响应：`cache_first`（配合离线 seed/clean 工具链）。
- 数据周期更新、可容忍单次慢请求：`ttl`（如 `ttl_hours: 24`）。
- 在线产品、延迟优先、数据最终一致即可：`revalidate`。
- 上游排查/演示实时源：`bypass`（注意上游压力）。
