# Hello World

## 一个样例

```go
package main

import (
	"fmt"
	"net/http"
)

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "你好")
}

func world(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "世界")
}

func main() {
	server := http.Server{
		Addr: "127.0.0.1:8080",
	}
	http.HandleFunc("/hello", hello)
	http.HandleFunc("/world", world)

	server.ListenAndServe()
}
```

当我们通过在当前路径打开命令行输入一下代码， 来开启服务： 

```go
go run main.go
```

然后打开浏览器， 输入http://127.0.0.1:8080/hello可以获得“你好”的响应文本， 而在http://127.0.0.1:8080/world， 则可以获取“世界”。

main函数主要做了3件事情：

1. 通过http.Server生成一个类型为Server的结构体， 该结构体除了Addr(路径)之外， 还有诸如Handler之类的参数，由于我们后面直接通过HandleFunc来注册路由的Handler(处理函数) ， 所以可以不需要提供该参数。
2. 调用http.HandleFunc (路由， 处理函数)来注册处理函数。 如果查看一下http.HandleFunc的源代码就可以知道， HandleFunc会调用DefaultServeMux作为我们的ServeMux类型

```go
func HandleFunc(pattern string, handler func(ResponseWriter, *Request)) {
	DefaultServeMux.HandleFunc(pattern, handler)
}
```

接着DefaultServeMux.HandleFunc最终会调用Handle函数来完成路由函数的注册。

```go
func (mux *ServeMux) HandleFunc(pattern string, handler func(ResponseWriter, *Request)) {
	if handler == nil {
		panic("http: nil handler")
	}
	mux.Handle(pattern, HandlerFunc(handler))
}
```

好奇的你可能会问， 如果我不想使用DefaultServeMux， 那该怎么办？

下面就是通过自定义ServerMux的实现方式：

```go
func main(){
    mux := http.NewServeMux()
    files := http.FileServer(http.Dir("/public"))

    mux.Handle("/hello", hello)) // 这里我们直接使用Handle而不是HandleFunc
    mux.Handle("/world", world))

	  server := &http.Server{
	    Addr:     "0.0.0.0:8080",
	    Handler:  mux,   // 把我们创建的ServeMux作为Handler参数传入http.Server结构体
	  }

  server.ListenAndServe()

```

3. 最后通过无参数函数ListenAndServe开启我们的服务， ListenAndServe会默认使用tcp协议。

有一点需要指出来的是， 上面提到的ServeMux以及DefaultServeMux(ServeMux的一个实例)本质上也是Handler，这样就是为什么http.NewServeMux返回的ServeMux可以作为Server的Handler的值。 

## Handle Function

我们知道上面例子中的hello函数和world函数都是handle function， 那究竟什么是handle fucntion呢？handle function需要两个参数， 第一个是http.ResponseWriter 以及一个指向http.Request的指针。

通过ResponseWriter命名也能够知道， 它是个类Writer接口， 所以它一定具有Write方法！所以才可以作为fmt.FPrintf的第一个参数； 当然更常见的情况是直接在调用它的Write方法， 虽然本质上两者其实一样的（fmt.FPrintf也会调用它的write方法）

这里还需要指出来的是， handle function和Handler接口（注意Handler是一个接口！）的ServeHTTP函数的签名是一模一样的；前面提到的ServeMux就具有这种这种接口， 所以ServeMux类就是Handler，意味着DefaultServeMux也是Handler。因此， 我们也可以完全由自己去定义一个Handler作为http.Server的Handler的值, 如下：

```go
type Myhandler struct {}

func (handler *Myhandler) ServeHTTP（w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "你好")
} 

func main(){
    server := &http.Server{
	      Addr:     "0.0.0.0:8080",
	      Handler:  MyHandler,
	  }
    server.ListenAndServe()
}
```

同样的逻辑， 所有基于net/http包开发的第三方web开发应用也一定具有这个方法， 比如gin, 我们来看一下gin中ServeHTTP方法， 它是定义在Engined对象上的， 所以Engine对象可以作为Handler传入到http.Server中

```go
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
		c := engine.pool.Get().(*Context)
		c.writermem.reset(w)
		c.Request = req
		c.reset()
	
		engine.handleHTTPRequest(c)
	
		engine.pool.Put(c)
}
```