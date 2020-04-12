# Goribot 扩展
::: tip 提示
扩展在`Spider`中注册的顺序会影响到实际上扩展 Hook 函数的执行顺序，请关注扩展执行的优先级。
:::

## SaveItemsAsJSON | 保存 Item 到 JSON 文件
```Go
f, err := os.Create("./test.json")
if err != nil {
	panic(err)
}
s := goribot.NewSpider(
	goribot.SaveItemsAsJSON(f),
)
```
详细用法请参考 [_examples/saver_extensions.go](https://github.com/zhshch2002/goribot/blob/master/_examples/saver_extensions.go)。

## SaveItemsAsCSV | 保存 Item 到 CSV 文件
```Go
f, err := os.Create("./test.cvs")
if err != nil {
	panic(err)
}
s := goribot.NewSpider(
	goribot.SaveItemsAsCSV(f),
)
```
详细用法请参考 [_examples/saver_extensions.go](https://github.com/zhshch2002/goribot/blob/master/_examples/saver_extensions.go)。

## Retry | 失败重试
```Go
s := goribot.NewSpider(
	goribot.Retry(3, http.StatusOK),
)
```
激活后会在蜘蛛会自动重试 **出现错误** 或 **不是指定响应码** 的请求，直到达到重试上限次数。

## RobotsTxt | Robots.txt 支持
```Go
s := goribot.NewSpider(
	goribot.RobotsTxt("https://github.com/", "Goribot"),
)
```
激活后会在蜘蛛会自动抛弃 robots.txt 所限制的请求。

## SpiderLogError | 记录意外和错误
```Go
f, _ := os.Create("./test.log")
s := goribot.NewSpider(
	goribot.SpiderLogError(f),
)
```
在蜘蛛遇到错误和意外（比如突然出现的验证码等）将日志记录下来。详细用法请参考 [_examples/logerror.go](https://github.com/zhshch2002/goribot/blob/master/_examples/logerror.go)。

## SpiderLogPrint | 打印蜘蛛运行状态
```Go
s := goribot.NewSpider(
	goribot.SpiderLogPrint(),
)
```
激活后会在蜘蛛开始和结束运行时打印日志，并每隔 5sec 打印蜘蛛执行了多少 Task 和收集了多少 Item。

## RefererFiller | 填充 Referer
```Go
s := goribot.NewSpider(
	goribot.RefererFiller(),
)
```
启用此插件后，使用`ctx`创建的新任务会自动携带创建该任务时的地址作为`Referer`。

## SetDepthFirst | 设置为深度优先策略
```Go
s := goribot.NewSpider(
	goribot.SetDepthFirst(true | false),
)
```
此扩展可以配置蜘蛛的爬取策略。
::: warning 警告
此扩展只支持使用`goribot.BaseScheduler`调度器。否则将触发`panic`。
:::

## ReqDeduplicate | 请求去重
```Go
s := goribot.NewSpider(
	goribot.ReqDeduplicate(),
)
```
此扩展会在`OnAdd`中判断当前`Req`的 Hash 是否出现过，若是将会抛弃该任务。

## RandomProxy | 随机代理
```Go
s := goribot.NewSpider(
	goribot.RandomProxy("proxy1","proxy2"),
)
```
此扩展会随机填充一个 UA 给 UA 为空的请求。

## RandomUserAgent | 随机 UA
```Go
s := goribot.NewSpider(
	goribot.RandomUserAgent(),
)
```
此扩展会随机填充一个 UA 给 UA 为空的请求。