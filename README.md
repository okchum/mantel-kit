# mantel-kit

[Mantel](https://github.com/okchum/mantel)（自托管常亮屏）和 [Mantel Cast](https://github.com/okchum/mantel-cast)（让 Nest Hub 等投屏设备一直亮着 Mantel 页面）共用的小工具，原来两边各复制一份、靠测试比对，现在都依赖这里。

- `conf`：配置里的基本类型和写法——`"30s"` 这样的时长（`Duration`）、`"$NAME"` / `"$HOME/ca.pem"` / `"Bearer $TOKEN"` 读环境变量（`ExpandEnv`、`EmptyEnv`）、`"23:30-07:00"` 这样的每日时段（`ParseWindow`、`InWindow`、`FormatWindow`；起点含、终点不含，终点可写 `24:00`，全天写 `00:00-24:00`）。

纯 Go 标准库，零第三方依赖。改了这里，打新 tag，再在两个仓库里 `go get github.com/okchum/mantel-kit@<tag>`。

本地同时改几个仓库时，在它们的上一级目录放一个 `go.work`：

```
go 1.24

use (
	./mantel
	./mantel-cast
	./mantel-kit
)
```

## 许可证

[MIT](LICENSE)
