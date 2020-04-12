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
* 丰富的扩展支持
  * 请求去重（👈支持分布式）
  * Json，csv 存储结果
  * Robots.txt支持
  * 记录请求异常
  * 随机 UA、随机代理
  * 失败重试
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

## 🙏感谢

* [ants](https://github.com/panjf2000/ants)
* [chardet](https://github.com/saintfish/chardet)
* [colly](https://github.com/gocolly/colly)
* [gjson](https://github.com/tidwall/gjson)
* [goquery](https://github.com/PuerkitoBio/goquery)
* [go-logging](https://github.com/op/go-logging)
* [go-redis](https://github.com/go-redis/redis)
* [robots](https://github.com/slyrz/robots)

万分感谢以上项目的帮助🙏。

## 📃TODO

* ~~分布式支持~~
* 扩展
  * ~~Json、CVS 数据收集~~
  * site Limiter
  * ~~随机代理~~
  * ~~错误重试~~
  * ~~过滤响应码~~
* English Document