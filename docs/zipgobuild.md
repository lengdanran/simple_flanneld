# 减小 Go 代码编译后的二进制体积

golang compiler optimization

```shell
#upx压缩 - 会有解压的操作，耗时可以忽略不计
go build -ldflags="-s -w" -o flanneld main.go && upx -9 flanneld



[root@zsh simple_flannel]# go build -ldflags="-s -w" -o flanneld main.go && upx -9 flanneld
                       Ultimate Packer for eXecutables
                          Copyright (C) 1996 - 2020
UPX 3.96        Markus Oberhumer, Laszlo Molnar & John Reiser   Jan 23rd 2020

        File size         Ratio      Format      Name
   --------------------   ------   -----------   -----------
  12169216 ->   4433684   36.43%   linux/amd64   flanneld                        

Packed 1 file.
```


## 1 基线用例
减小编译后的二进制的体积，能够加快程序的发布和安装过程。接下来呢，我们分别从编译选项和第三方压缩工具两方面来介绍如何有效地减小 Go 语言编译后的体积。

我们采用同一个测试工程来测试不同方式的效果。

使用的测试工程如下，该程序启动了一个 RPC 服务，引用了 log、net/http 和 net/rpc 三个 package。
```go
package main

import (
"log"
"net/http"
"net/rpc"
)

type Result struct {
Num, Ans int
}

type Calc int

// Square calculates the square of num
func (calc *Calc) Square(num int, result *Result) error {
result.Num = num
result.Ans = num * num
return nil
}

func main() {
rpc.Register(new(Calc))
rpc.HandleHTTP()

	log.Printf("Serving RPC flanneld on port %d", 1234)
	if err := http.ListenAndServe(":1234", nil); err != nil {
		log.Fatal("Error serving: ", err)
	}
}
```
使用默认选项编译该程序，编译后的程序大小约为 9.8M。
```shell
$ go build -o flanneld main.go
$ ls -lh flanneld
-rwxr-xr-x  1 dj  staff   9.8M Dec  7 23:57 flanneld
```
## 2 编译选项
Go 编译器默认编译出来的程序会带有符号表和调试信息，一般来说 release 版本可以去除调试信息以减小二进制体积。
```shell
$ go build -ldflags="-s -w" -o flanneld main.go
$ ls -lh flanneld
-rwxr-xr-x  1 dj  staff   7.8M Dec  8 00:29 flanneld
```
-s：忽略符号表和调试信息。
-w：忽略DWARFv3调试信息，使用该选项后将无法使用gdb进行调试。
体积从 9.8M 下降到 7.8M，下降约 20%。

## 3 使用 upx 减小体积
### 3.1 upx 安装
UPX is an advanced executable file compressor. UPX will typically reduce the file size of programs and DLLs by around 50%-70%, thus reducing disk space, network load times, download times and other distribution and storage costs.

upx 是一个常用的压缩动态库和可执行文件的工具，通常可减少 50-70% 的体积。

upx 的安装方式非常简单，我们可以直接从 github 下载最新的 release 版本，支持 Windows 和 Linux，在 Ubuntu 或 Mac 可以直接使用包管理工具安装。例如 Mac 下可以直接使用 brew：
```shell
$ brew install upx
$ upx
Ultimate Packer for eXecutables
Copyright (C) 1996 - 2020
UPX 3.96        Markus Oberhumer, Laszlo Molnar & John Reiser   Jan 23rd 2020

Usage: upx [-123456789dlthVL] [-qvfk] [-o file] file..
...
Type 'upx --help' for more detailed help.
```
### 3.2 仅使用 upx
upx 有很多参数，最重要的则是压缩率，1-9，1 代表最低压缩率，9 代表最高压缩率。

接下来，我们看一下，如果只使用 upx 压缩，二进制的体积可以减小多少呢。
```shell
$ go build -o flanneld main.go && upx -9 flanneld
File size         Ratio      Format      Name
   --------------------   ------   -----------   -----------
10253684 ->   5210128   50.81%   macho/amd64   flanneld
$ ls -lh flanneld
-rwxr-xr-x  1 dj  staff   5.0M Dec  8 00:45 flanneld
```
可以看到，使用 upx 后，可执行文件的体积从 9.8M 缩小到了 5M，缩小了 50%。

### 3.3 upx 和编译选项组合
```shell
$ go build -ldflags="-s -w" -o flanneld main.go && upx -9 flanneld
File size         Ratio      Format      Name
   --------------------   ------   -----------   -----------
8213876 ->   3170320   38.60%   macho/amd64   flanneld
$ ls -lh flanneld
-rwxr-xr-x  1 dj  staff   3.0M Dec  8 00:47 flanneld
```
使用编译选项后，体积从原来的 9.8M 下降了 20% 到 7.8M，使用 upx 压缩后，体积进一步下降 60% 到 3M。累进下降约 70%。

### 3.4 upx 的原理
upx 压缩后的程序和压缩前的程序一样，无需解压仍然能够正常地运行，这种压缩方法称之为带壳压缩，压缩包含两个部分：

在程序开头或其他合适的地方插入解压代码；
将程序的其他部分压缩。
执行时，也包含两个部分：

首先执行的是程序开头的插入的解压代码，将原来的程序在内存中解压出来；
再执行解压后的程序。
也就是说，upx 在程序执行时，会有额外的解压动作，不过这个耗时几乎可以忽略。

如果对编译后的体积没什么要求的情况下，可以不使用 upx 来压缩。一般在服务器端独立运行的后台服务，无需压缩体积。