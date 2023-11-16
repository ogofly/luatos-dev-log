# golang 版本的合宙 luatos err dump 日志接收服务

原版构建部署麻烦一些，现使用 golang 简化开发部署，并支持了 sqlite3

### 依赖以下标准定义
- 设备采用 udp 端口上报错误日志
- 错误日志格式
  如：pencpu-slc_lodverxxx,0.9.0,866250060829193,91937594125402,long error mesage
  解析为以下字段：
  ```
  dev: 866250060829193
  proj: opencpu-slc
  lodver: lodverxxx
  selfver: 0.9.0
  devsn: 91937594125402
  errlog: long error mesage
  ```
### 运行

go run main.go -h 查看配置参数

### 测试
```bash
nc -u localhost 9072
```
输入： `pencpu-slc_lodverxxx,0.9.0,866250060829193,91937594125402,long error mesage` 测试数据  
查看数据库是否成功保存

