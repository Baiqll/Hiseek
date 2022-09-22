# Hiseek
一个快速从web.archive.org找到目标的工具

#### 使用
```bash

  # 
  cat domains.txt | Hiseek 
  
  cat example.com | Hiseek

  Hiseek -d example.com 


  # 文件后缀查询
  cat example.com | Hiseek -s \.js

  # 多个关键字查询
  cat example.com | Hiseek -s proxy,url=,=http 

  # 关键字排除查询
  cat example.com | Hiseek -e www

  # 使用子域名 查询（example.com 存档记录太多时，有些记录查不到，可以使用 example.com 的子域名进行搜索）
  cat example.com | Hiseek -w dict.txt 

  # 查询结果导出 子域名、子域名字典
  # 生成 sub_domain_example.txt 、dict_example.txt 两个文件
  cat example.com | Hiseek -od example.txt
  

  # 联合xray、nuclei 等进行被动扫描
  cat example.com | Hiseek -scan http://127.0.0.1:7777 

  cat example.com | Hiseek -silent | httpx -proxy http://127.0.0.1:7777 

  # 设置代理 proxy （ 对于国内无法访问直接 web.archive.org 的情况下 需要设置代理）
  cat example.com | Hiseek -proxy http://127.0.0.1:1089

```


#### 安装

```bash

  go install github.com/arews-cn/Hiseek@latest

```


#### 自己打包
```bash
 # 默认打包
 go build 

 # mac
 CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o Hiseek main.go 
 
 # windows
 CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o Hiseek.exe main.go

 # linux
 CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o Hiseek main.go 

```