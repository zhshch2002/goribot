# 分布式爬虫
要使用 Goribot 的分布式功能，你需要一个 Redis 服务器来传递爬虫间的信息。推荐使用 Docker 来快速开启一个 Redis 服务器。

```sh
docker run --name some-redis -d -p 6379:6379 redis
```

## 分布式结构
你需要阅读完之前的文档，尤其是 [《开始使用》](./get-start.html) 以确保你更好的理解以下内容。

在非分布式模式下部署 Goribot 爬虫大致是这样的：
1. 新建一个爬虫
2. 配置扩展等
3. 添加种子任务及其回调函数
4. 必要的话更多的任务……
5. Run

分布式支持将作为一类扩展，像其他扩展一样注册到爬虫上。分布式模式下，除了 Redis 服务器，我们还需要两个东西：
1. 爬虫，也就是安装了分布式扩展的 Goribot 爬虫。
2. 管理器，用于发布爬虫种子任务并回收爬虫结果（爬虫结果也可以由爬虫本身处理，此部分为可选）。

## 管理器 Manager
管理器有连个任务，其使用和 Goribot 的 Spider 十分类似。
```Go
package main

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/zhshch2002/goribot"
)

func main() {
	ro := &redis.Options{ // Redis 配置
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
    }
    
    sName := "DistributedTest" // 爬虫识别名，用于在 Redis 服务器上区分爬虫。保证一个爬虫用一个名字就行，内容无所谓。

	m := goribot.NewManager(redis.NewClient(ro), sName) // 创建 Manager
	m.OnItem(func(i interface{}) interface{} { // 处理蜘蛛回传的结果，其中使用了 gob 包来转换结构体
		fmt.Println(i)
		return i
    })
    
    // 发布爬取种子任务。
    // ❗ 注意这里只有请求 Request，请求对应的回调函数 Handler 需要在蜘蛛侧配置
    m.SendReq(goribot.GetReq("https://httpbin.org/get").SetHeader("goribot", "hello world"))
    m.SendReq(goribot.GetReq("https://httpbin.org/get").SetHeader("goribot", "hello second"))
    // ……

	m.Run() // 开始运行爬虫结果回收线程，次调用将阻塞线程❗永不退出
}
```

## 爬虫 Spider
为了支持分布式操作，对一般的单机爬虫做了如下扩展：
1. 替换了原先的调度器`Scheduler`，从 Redis 同步种子任务，同时维护本地任务队列
2. 替换了原先的调度器`Scheduler`，截获爬虫存储的`Item`，上报给`Manager`，同时保留本机处理`Item`的原本功能
3. 提供了`RedisReqDeduplicate`替换原有的`ReqDeduplicate`扩展，用于在多个爬虫节点间去重任务。

```Go
package main

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/zhshch2002/goribot"
)

func main() {
	ro := &redis.Options{ // Redis 配置
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
    }
    
    sName := "DistributedTest" // 爬虫识别名，用于在 Redis 服务器上区分爬虫。保证一个爬虫用一个名字就行，内容无所谓。
    
	s := goribot.NewSpider( // 创建蜘蛛
		goribot.RedisDistributed( // 加载分布式扩展
			ro, // Redis 配置
			sName, // 爬虫识别名
			true, // 是否激活`RedisReqDeduplicate`去重
			func(ctx *goribot.Context) { // 种子任务的回调函数 ❗注意❗是在这里设置的，不是发布任务时
				goribot.Log.Info("got seed resp")
				ctx.AddItem(ctx.Resp.Text)
                ctx.AddTask(
                    goribot.GetReq("https://httpbin.org/get").SetHeader("goribot", "hi!"), 
                    func(ctx *goribot.Context) {
                        goribot.Log.Info("got resp")
                        ctx.AddItem(ctx.Resp.Text)
                    },
                )
			},
		),
	)

	s.Run() // 被添加分布式扩展的蜘蛛❗不会自动停止，将会一直阻塞线程
}
```

## 完成
🎉分别在不同的机器上运行不同的程序就行了！