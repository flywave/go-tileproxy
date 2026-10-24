# 将 GCJ-02 XYZ 瓦片重投影为标准 Web Mercator

此示例依赖 go-geo 新增的 `GCJ02:3857` 自定义投影。需要包含该功能的 go-geo 版本，不能将名称替换为旧的 `EPSG:GCJ02MC`：两者投影公式不同。

处理链为：原始 GCJ-02 XYZ → 原始瓦片缓存 → 重投影缓存（EPSG:3857）→ XYZ 客户端。地点数据仍保留 WGS84，无需按图源修改地点坐标。

在 `examples/amap-wgs84.json` 中填写有权使用的瓦片源地址，之后运行：

```sh
go run ./cmd serve --config examples/amap-wgs84.json --host 127.0.0.1 --port 4387 --watch=false
```

客户端连接 `http://127.0.0.1:4387/tiles/amap/{z}/{x}/{y}.png`，瓦片大小 256。网格配置使用 `ul` 原点，服务对外使用 `nw` 的 XYZ 请求方向。

示例不包含任何账号、密钥或私人地点。图源的授权、覆盖范围与可用缩放级别由使用者配置；代理不授予额外图源使用权限。首次请求会获取并重采样多张来源瓦片，之后使用本地缓存。重采样会带来少量模糊。
