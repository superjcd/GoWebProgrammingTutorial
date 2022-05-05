# ChitChat

## 运行方式
应用链接的是Postgresql数据库， 如果没有的话就需要安装一个。  

使用data文件夹下main的setup.sql， 建立数据库chitchat， 以及后面需要用到的若干表。

在data目录下的data.go文件中更新一下数据库的链接信息

```
psqlInfo := fmt.Sprintf("user=%s "+
    "password=%s dbname=%s sslmode=disable",
    "postgres", "root", "chitchat")

```

最后通过运行`go run main.go`来开启服务。