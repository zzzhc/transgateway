# 基本需求

## 目标

聚合多种翻译服务，对外提供统一接口。

## 配置

通过`-c config_path`指定配置文件路径

config.example.yml:

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

## API服务

### GET /health

response:

```json
{"status": "ok"}
```

### GET /translate

request parameters:

* provider: 使用哪个翻译服务，默认为mtranserver
* from: source language, ISO 639 language codes Set1, 如en, 值为auto时自动检测语言
* to: target language, ISO 639 language codes Set1, 如zh
* text: 需要翻译的文本

response:

```json
{
    "detected_source_lang": "ja",
    "result": "..."
}
```

### POST /translate

request body:

```json
{
    "provider": "mtranserver",
    "from": "auto",
    "to": "zh",
    "text": "hello, world"
}
```

说明:

* provider: 使用哪个翻译服务，默认为mtranserver
* from: source language, ISO 639 language codes Set1, 如en, 值为auto时自动检测语言
* to: target language, ISO 639 language codes Set1, 如zh
* text: 需要翻译的文本

response:

```json
{
    "detected_source_lang": "ja",
    "result": "..."
}
```

### POST /batch_translate

request body:

```json
{
    "provider": "mtranserver",
    "from": "auto",
    "to": "zh",
    "texts": [
        "hello, world",
        "welcome"
    ]
}
```

说明:

* provider: 使用哪个翻译服务，默认为mtranserver
* from: source language, ISO 639 language codes Set1, 如en, 值为auto时自动检测语言
* to: target language, ISO 639 language codes Set1, 如zh
* text: 需要翻译的文本

response:

```json
{
    "detected_source_lang": "ja",
    "results": [
        "你好,世界",
        "欢迎"
    ]
}
```
