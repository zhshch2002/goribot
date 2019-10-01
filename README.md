# Goribot
A golang spider framework.

[![Codecov](https://img.shields.io/codecov/c/gh/zhshch2002/goribot)](https://codecov.io/gh/zhshch2002/goribot)
![GitHub](https://img.shields.io/github/license/zhshch2002/goribot)

# Example
```go
package main

import (
	"fmt"
	"github.com/zhshch2002/goribot"
)

func main() {
	s := goribot.NewSpider()
	s.NewTask(
		goribot.MustNewGetReq("https://httpbin.org/get?Goribot%20test=hello%20world"),
		func(ctx *goribot.Context) {
			fmt.Println("got resp data", ctx.Text)
		})
	s.Run()
}

```

# Start to use
## install
```shell
go get -u github.com/zhshch2002/goribot
```

# TODO
* 制作更多的内建插件`extensions`
    * [x] 随机UA
    * [ ] 自动添加Referer
    * [ ] Robots.txt解析
    * [ ] 最大请求数限制
    * [ ] Host和URL过滤
    * [ ] 随机代理
* `Spider`主体功能
    * [ ] 随机延时