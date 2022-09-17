# Hiseek
一个快速从web.archive.org找到目标的工具


#### 各版本打包命令
```bash
 # 默认打包
 go build 

 # 指定打包
 # mac
 CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o Hiseek main.go 
 
 # windows
 CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o Hiseek.exe main.go

 # linux
 CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o Hiseek main.go 

```