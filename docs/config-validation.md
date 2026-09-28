# 配置校验与热更新对接指南

本文说明 flywave-server 如何接入 go-tileproxy 的配置校验与热更新能力，实现
“前端修改配置 → 后端校验 → 反馈给前端 → 校验通过后热生效” 的完整闭环。

## 一、总体流程

```
前端表单/编辑器
      │  ① POST 配置 JSON
      ▼
flywave-server 校验端点  ──② 调 setting.ValidateConfigJSON──▶ go-tileproxy
      │  ③ 返回结构化校验报告 (JSON)                                │
      ▼                                                            │
前端渲染错误/警告（字段级定位）                                     │
      │  ④ 校验通过后保存并调用热更新                                │
      ▼                                                            ▼
flywave-server 保存配置文件 ──▶ proxy.ReloadFromFile / ReloadServices
```

关键原则：**校验未通过时禁止保存/应用**；热更新 API 自身也会再校验一次，
即使有配置绕过前端直接写入文件，坏配置也不会影响正在运行的服务。

## 二、校验 API（go-tileproxy 提供）

### 1. 整份配置校验（推荐给 HTTP 校验端点使用）

```go
import "github.com/flywave/go-tileproxy/setting"

func validateHandler(w http.ResponseWriter, r *http.Request) {
    body, err := io.ReadAll(io.LimitReader(r.Body, setting.MaxJSONSize+1))
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    report := setting.ValidateConfigJSON(body) // 永不 panic，总是返回报告
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(report)
}
```

### 2. 单服务校验（编辑单个服务时增量校验）

```go
ps := &setting.ProxyService{}
json.Unmarshal(payload, ps)          // 前端提交的原始 JSON
rep := setting.ValidateProxyService(ps) // 转换 + 全量检查，返回所有问题
```

校验的副作用是把 raw payload 转换为具体类型结构，因此校验通过的对象可直接
交给 `tileproxy.ReloadServices` 使用。

### 3. CLI（运维 / CI 使用）

```bash
go run ./cmd validate -c config.json   # 输出报告 JSON；无效时退出码 1
```

## 三、校验报告结构（响应契约）

```json
{
  "valid": false,
  "errors": [
    {"field": "", "message": "invalid JSON document: ..."}
  ],
  "services": [
    {
      "id": "svc1",
      "index": 0,
      "valid": false,
      "errors": [
        {"field": "service.type", "message": "unknown or missing service type \"tms2\" (expected wms, tms, wmts, mapbox or cesium)"},
        {"field": "sources.s1.url_template", "message": "url_template is required"},
        {"field": "sources.s1.grid", "message": "references undefined grid \"g9\""},
        {"field": "caches.c1.grid", "message": "references undefined grid \"g9\""},
        {"field": "service.layers.0.source", "message": "references undefined cache or source \"nope\""}
      ],
      "warnings": [
        {"field": "service.image_formats", "message": "no image formats defined; clients cannot request maps"}
      ]
    }
  ]
}
```

字段说明：

| 字段 | 含义 |
|---|---|
| `valid` | 整份配置（或单个服务）是否可应用 |
| `errors[].field` | 配置对象内的字段路径，如 `sources.s1.url_template`；文档级错误为空串 |
| `errors[].message` | 人类可读的错误说明（缺失、非法、引用不存在等） |
| `warnings[]` | 非致命提示（如 WMS 未配 image_formats），配置可应用 |
| `services[].index` | 服务在配置文件中的序号，便于前端定位表单项 |

### 校验覆盖范围

- **缺失必填项**：service 类型、grid 的 srs/bbox、source 的 url/url_template、
  cache 的 grid/directory/tile_options、layer 的 source 引用等。
- **取值合法性**：SRS 标识、bbox 顺序（min < max）、directory_layout 枚举、
  tile_options 类型、service id 格式（URL 安全，且不能占用保留字 `health`）。
- **交叉引用**：service layers → caches/sources、caches → grids、
  sources → grids、cache → sources（warning 级）。
- **重复与未知**：重复的服务 id、未知的服务/数据源类型（会被明确报错，
  而不是静默忽略）。

## 四、热更新联动（校验通过后）

```go
proxy := tileproxy.NewTileProxy(globals, services, nil)

// 方式 A：flywave-server 已拿到配置内容（推荐——保存前先校验）
services, err := setting.LoadProxyServicesFromFile(path)
if err == nil {
    err = proxy.ReloadServices(services, nil)
}

// 方式 B：直接从文件热加载（内部会再校验一次）
err = proxy.ReloadFromFile(path, nil)
```

两种方式在配置非法时都返回错误且**不影响正在运行的服务**；被替换的旧服务
会等飞行中的请求完成后在后台释放（最多 30s）。

可选：文件监听（不依赖调用方触发）

```go
watcher := tileproxy.NewConfigWatcher(proxy, path, nil, 2*time.Second, time.Second)
watcher.Start()
defer watcher.Stop()
```

## 五、端到端示例（curl 模拟前端）

```bash
# ① 前端提交配置到校验端点
curl -s -X POST --data @config.json http://flywave-server/api/tileproxy/validate

# ② 响应 valid=true 后，保存配置文件

# ③ 触发热更新（或等待文件监听自动生效）
curl -s -X POST http://flywave-server/api/tileproxy/reload
```
