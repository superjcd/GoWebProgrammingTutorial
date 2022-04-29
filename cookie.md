# Cookie

我们知道http请求时一种无状态的请求， 每一次的都是独立的； 但是很多时候我么需要记住用户的状态(比如用户的地区、 使用语言等信息)， 所以我们需要在无状态的请求中加一些可以持续存在的信息， 那么cookie就是一种比较古老且常见的方式。

和Cookie相关的操作有两个， 一个是设置cookie的， 由服务端写在响应头中;另外一个是获取cookie, cookie会由用户通过请求头传送给服务器。

### 设置cookie

```go
func setCookie(w http.ResponseWriter, r *http.Request) {
		 c1 := http.Cookie{
		     Name: "cookie1",
		     Value: "123",
		     HttpOnly: true,
		 }
		 c2 := http.Cookie{
		     Name: "cookie2",
	       Value: "abc",
		     HttpOnly: true,
		 }
		 w.Header().Set("Set-Cookie", c1.String())
		 w.Header().Add("Set-Cookie", c2.String())
}

```

这里我们在ResponseWriter里写入了两个Cookie对象, 注意第一次写入的时候用的是Set方法， 第二次用的是Add(不然会覆盖掉前面设置的Cookie)

<aside>
💡 注意cookie的结构， Cookie结构中最为重要的是name和value属性， 用来存储服务端需要客户端存储的信息；其他比较重要的属性， 包括Expires和MaxAge来设置过期日期和cookie的有效时长;如果我们设置Expires和MaxAge的话， 那么我们的cookie的会在会话结束后被清空

</aside>

处理类似上面的设置cookie方式， http库还提供了下面的快捷方式：

```go
http.SetCookie(w, &c1)
http.SetCookie(w, &c2)
```

### 获取cookie

类似地， 获取cookie也有两种方式：

第一种:

```go
func getCookie(w http.ResponseWriter, r *http.Request){
    cookie := r.Header["Cookie"]
    ...
}
```

上面的之中方式会将所有的cookie以一个字符串的形式返回；用户需要自己去解析具体的某个键值，

当然http也提供了便捷的方式：

```go
func getCookie(w http.ResponseWriter, r *http.Request){
    cookie1， err := r.Cookie("cookie1")
    ...
}
```