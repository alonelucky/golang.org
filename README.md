# Golang.org/x project mirror
个人梳理创建该目录，同步自 github.com/golang
```
#!/bin/bash

url="https://github.com/golang/"
urls=(
    "arch"
	"build"
	"crypto"
	"debug"
	"dl"
	"exp"
	"image"
	"lint"
	"mobile"
	"net"
	"perf"
	"playground"
	"review"
	"scratch"
	"sync"
	"sys"
	"time"
	"tools"
	"tour"
	"vgo"
)
str="/archive/master.zip"

for u in ${urls[@]};do
  echo $url$u$str;
  wget $url$u$str
  unzip master.zip
  mv $u-master $u
  rm master.zip
done;
```

## 若觉得麻烦，或1.11以后的go mod 模式，可以配置go proxy 代理服务
[Go module 私服](https://qiubo.ink/2019/03/19/go-module-e7-a7-81-e6-9c-8d/)
