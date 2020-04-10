# Goribot 扩展
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

## RandomUserAgent | 随机 UA
```Go
s := goribot.NewSpider(
	goribot.RandomUserAgent(),
)
```
此扩展会随机填充一个 UA 给 UA 为空的请求。