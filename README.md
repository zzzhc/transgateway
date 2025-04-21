# TransGateway

这是一个基于 Go 语言的翻译服务聚合项目，支持多种翻译服务提供商，提供统一的 REST API 接口。

## 功能特性

- 支持多种翻译服务提供商（mtranserver、Google 等）
- 提供统一的 REST API 接口
- 支持单文本翻译和批量翻译
- 支持自动语言检测
- 支持自定义配置

## 配置说明

通过 `-c` 参数指定配置文件路径，默认使用 `config.yml`。

配置文件示例：

```yaml
host: 127.0.0.1
port: 5678
providers:
    mtranserver:
        enable: true
        endpoints:
            - http://192.168.0.100:8989
            - http://192.168.0.101:8989
    google:
        enable: false
        proxy: http://127.0.0.1:7890
```

## API 文档

### 健康检查

```
GET /health
```

响应：

```json
{"status": "ok"}
```

### 单文本翻译

```
GET /translate?provider=mtranserver&from=auto&to=zh&text=hello
```

或

```
POST /translate
Content-Type: application/json

{
    "provider": "mtranserver",
    "from": "auto",
    "to": "zh",
    "text": "hello"
}
```

参数说明：

- provider: 翻译服务提供商，默认为 mtranserver
- from: 源语言，使用 ISO 639 语言代码，auto 表示自动检测
- to: 目标语言，使用 ISO 639 语言代码
- text: 需要翻译的文本

响应：

```json
{
    "detected_source_lang": "en",
    "result": "你好"
}
```

### 批量翻译

```
POST /batch_translate
Content-Type: application/json

{
    "provider": "mtranserver",
    "from": "auto",
    "to": "zh",
    "texts": ["hello", "world"]
}
```

响应：

```json
{
    "detected_source_lang": "en",
    "results": ["你好", "世界"]
}
```

## 运行项目

```bash
# 使用默认配置文件
go run main.go

# 指定配置文件
go run main.go -c config.yml
```

## 构建项目

```bash
go build
```

## 许可证

Apache 2.0
