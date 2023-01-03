### 设计思路

框架层总体包含前端、后端、聚合应用三部分

- 前端front

  封装UI界面 

- 后端back

  提炼出客户端角度的玩家通用接口

  用户可继承，可重写

- 聚合应用app

  作为连接前端后端的桥梁

  封装一些框架接口

##### 前端

ui界面其实就是一个网页，内容就是index.html

开源库lorca负责拉起浏览器并访问index.html 

并且提供了GO与JS相互调用的接口

- GO调用JS

  假如js提供了函数add

  ```js
  function add(x, y){
      return x + y
  }
  ```

  在GO代码中调用add

  ```go
  ui.Eval(`add(2, 3)`).Int() // 
  ```

  调用操作已封装在calljs.go文件中

- JS调用GO

  go实现的函数

  ```go
  func Add(x, y int) int{
  	return x + y
  }
  // 调用 后相当于js环境有了Add这个变量
  ui.Bind("Add", Add)
  ```

  在JS代码中 使用

  ```js
  // 调用Add返回的是个promise 是异步的
  Add(1, 2).then((res)=>{
      console.log(res) // res = 3
  })
  ```

  实际调用已封装在fromjs.go中

除了实现GO JS互相调用外，还需实现一个静态文件服务器

让浏览器可以获取js、css等静态文件

具体实现在fileserver.go

##### 后端

主要是IGamer接口

```go
type IGamer interface {
	GetUid() string // 玩家唯一标识，登录前就要确定
	Close()         // 主动调用 关闭玩家
	OnExit()        // 退出回调

	MsgChan() <-chan interface{} // 消息channel
	ExitChan() <-chan struct{}   // exit channel
	ProcessMsg(interface{})      // 网络消息处理函数

	GetTickMs() int64 // 玩家间隔多少毫秒跑一遍行为树
    Stop() 			 // 让玩家停止跑行为树，但是不退出 此时不发消息只收消息 提升消息指标的精度
	IsStopped() bool // 
}
```

##### 聚合应用

主要就是IApp接口，用户直接继承即可

```go
type IApp interface {
	Init(*Options) // 初始化应用
	Run(context.Context) // 程序主循环，已有默认实现， 可按需重写
	Close()    // 关闭应用
	OnExit(reason string) // 应用退出回调

	// 根据配置文件创建一个玩家
	CreateGamer(confJson string, seq int32) (back.IGamer, error)
	RunGamer(g back.IGamer, tree *btree.Tree) // 玩家主循环 ，已有默认实现， 可按需重写

	// 解析pb结构体
	ParsePbInfo(back.IPbInfo)

	StressStart(start, count int32, treeID, confJs string) // 压测开始
	PrintStressStatus()     // 打印压测状态
}
```

