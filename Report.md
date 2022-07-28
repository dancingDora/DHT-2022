# Distributed Hash Map
|Contributor|[夏天](https://github.com/Rainy-Memory)|[赵一龙](http://github.com/happierpig)|[陆逸凡](https://github.com/dancingDora/DHT-2022)|
|:---:|:---:|:---:|:---:|
## Contents
* chord Protocol
    * 文件结构
    * 协议细节
* Kademlia Protocol
    * 文件结构
* Application
    * 实现功能
    * 使用说明
* Reference
## [Chord Protocol](https://github.com/dancingDora/DHT-2022/tree/main/src/chord)
### 文件结构
#### directory : chord
```go
DHT.go 
    //type Node
func.go
    //tools
network.go
    //rpc相关
    //type network
    //包括RemoteCall客户端远程调用
NodeWrapper.go
    //满足rpc调用格式
```
#### directory : main
```go
main.go
userdef.go
......
```
### 协议细节
#### RPC
* `rpc.RegisterName()` 将对象方法注册为`RPC`函数
* `listener, err := net.Listen("tcp", ...)` 、`conn, err := listener.Accept()`建立唯一`TCP`链接
* `rpc.ServeConn()`在该`TCP`链接下提供`RPC`服务
* `client, err := rpc.Dial("tcp", ...)`用户拨号
* `err = client.Call()`用户调用方法
#### advance
`ForceQuit` : `station(*network). ShutDown()`
## Kademlia Protocol
### 文件结构
#### directory : Kademlia
```go
DHT.go
func.go
network.go
NodeWrapper.go
```
## Application
### 实现功能
* 上传：指定路径文件上传至DHT网络并产生BT种子文件
* 下载：通过BT种子文件将目标文件下载到指定路径下
### 使用说明
输入IP开始程序, 每一行输入四个string类型参数
第一个参数表达指令类型，共分为6种
*  `[JOIN]` `chord.Join(para2)`
*  `[CREATE]`
*  `[DOWNLOAD]` `download(para2, para3, myIP)`
*  `[QUIT]` `dhtNode.Quit()`
*  `[RUN]` `dhtNode.Run()`
## Reference
*   ### 使用包
    *   [github.com/sirupsen/logrus](https://github.com/sirupsen/logrus)
    *   [github.com/fatih/color](https://github.com/fatih/color)
    *   [github.com/jackpal/bencode-go](https://github.com/jackpal/bencode-go)
*   ### 参考资料
    * go语言学习
        * [go.dev/tour/welcome/1](https://go.dev/tour/welcome/1)
        * [runoob.com/go/go-tutorial.html](https://www.runoob.com/go/go-tutorial.html)
    * Chord学习
        * [pdos.csail.mit.edu/papers/chord:sigcomm01/chord_sigcomm.pdf](https://pdos.csail.mit.edu/papers/chord:sigcomm01/chord_sigcomm.pdf)
        * [luyuhuang.tech/2020/03/06/dht-and-p2p.html](https://luyuhuang.tech/2020/03/06/dht-and-p2p.html)
    * RPC学习
        * [chai2010.cn/advanced-go-programming-book/ch4-rpc/ch4-01-rpc-intro.html](https://chai2010.cn/advanced-go-programming-book/ch4-rpc/ch4-01-rpc-intro.html)
    * Kademlia学习
        *  [luyuhuang.tech/2020/03/06/dht-and-p2p.html](https://luyuhuang.tech/2020/03/06/dht-and-p2p.html)
        *  [pdos.csail.mit.edu/~petar/papers/maymounkov-kademlia-lncs.pdf](https://pdos.csail.mit.edu/~petar/papers/maymounkov-kademlia-lncs.pdf)
        *  [http://www.cnblogs/com/LittleHann/p/6180296.html](http://www.cnblogs/com/LittleHann/p/6180296.html)
    * Torrent & App 学习
        *  [blog.jse.li/posts/torrent/#putting-it-all-together](https://blog.jse.li/posts/torrent/#putting-it-all-together)
        *  [www.cnblogs.com/LittleHann/p/6180296.html](https://www.cnblogs.com/LittleHann/p/6180296.html)
        *  [blog.mynameisdhr.com/YongGOCongLingJianLiBitTorrentKeHuDuan/](https://blog.mynameisdhr.com/YongGOCongLingJianLiBitTorrentKeHuDuan/)
        *  [github.com/veggiedefender/torrent-client](https://github.com/veggiedefender/torrent-client)