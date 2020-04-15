---
title: 欢迎 Welcome！
---

# Goribot
一个轻量的分布式友好的 Golang 爬虫框架。

## 🚀Feature
* 优雅的 API
* 整洁的文档
* 高速（单核处理 >1K task/sec）
* 友善的分布式支持
* 便捷的细节
  * 相对链接自动转换
  * 字符编码自动解码
  * HTML,JSON 自动解析
* 丰富的扩展支持
  * [请求去重](./extensions.html#reqdeduplicate-%e8%af%b7%e6%b1%82%e5%8e%bb%e9%87%8d)（👈支持分布式）
  * [限制请求、速率、并发](./extensions.html#limiter-%e9%99%90%e5%88%b6%e8%af%b7%e6%b1%82%e3%80%81%e9%80%9f%e7%8e%87%e3%80%81%e5%b9%b6%e5%8f%91)
  * [Json](./extensions.html#saveitemsasjson-%e4%bf%9d%e5%ad%98-item-%e5%88%b0-json-%e6%96%87%e4%bb%b6)，[CSV](./extensions.html#saveitemsascsv-%e4%bf%9d%e5%ad%98-item-%e5%88%b0-csv-%e6%96%87%e4%bb%b6) 存储结果
  * [Robots.txt 支持](./extensions.html#robotstxt-robots-txt-%e6%94%af%e6%8c%81)
  * [记录请求异常](./extensions.html#spiderlogerror-%e8%ae%b0%e5%bd%95%e6%84%8f%e5%a4%96%e5%92%8c%e9%94%99%e8%af%af)
  * [随机 UA ](./extensions.html#randomuseragent-%e9%9a%8f%e6%9c%ba-ua)、[随机代理](./extensions.html#randomproxy-%e9%9a%8f%e6%9c%ba%e4%bb%a3%e7%90%86)
  * [失败重试](./extensions.html#retry-%e5%a4%b1%e8%b4%a5%e9%87%8d%e8%af%95)
* 轻量，适于学习或快速开箱搭建

::: warning 版本警告
Goribot 仅支持 Go1.13 及以上版本。
:::

## 👜获取 Goribot
```sh
go get -u github.com/zhshch2002/goribot
```
::: tip
Goribot 包含一个历史开发版本，如果您需要使用过那个版本，请拉取 Tag 为 v0.0.1 版本。
:::

## ⚡建立你的第一个项目
```Go
package main

import (
	"fmt"
	"github.com/zhshch2002/goribot"
)

func main() {
	s := goribot.NewSpider()

	s.AddTask(
		goribot.GetReq("https://httpbin.org/get"),
		func(ctx *goribot.Context) {
			fmt.Println(ctx.Resp.Text)
			fmt.Println(ctx.Resp.Json("headers.User-Agent"))
		},
	)

	s.Run()
}
```

## 🎉完成
至此你已经可以使用 Goribot 了。更多内容请从 [开始使用](./get-start) 了解。
![](https://cdn.jsdelivr.net/gh/zhshch2002/pic/20200414171115.png)


## 🙏感谢

* [ants](https://github.com/panjf2000/ants)
* [chardet](https://github.com/saintfish/chardet)
* [colly](https://github.com/gocolly/colly)
* [gjson](https://github.com/tidwall/gjson)
* [goquery](https://github.com/PuerkitoBio/goquery)
* [go-logging](https://github.com/op/go-logging)
* [go-redis](https://github.com/go-redis/redis)
* [robots](https://github.com/slyrz/robots)
* [glob](https://github.com/gobwas/glob)

万分感谢以上项目的帮助🙏。

## 📃TODO

* ~~分布式支持~~
* 扩展
  * ~~Json、CVS 数据收集~~
  * ~~Limiter~~
  * ~~随机代理~~
  * ~~错误重试~~
  * ~~过滤响应码~~
* English Document