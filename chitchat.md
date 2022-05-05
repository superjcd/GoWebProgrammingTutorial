# chitchat

## 目录结构

```go
├─data          数据模型及相关操作                  
├─public        静态资源
│  ├─css               
│  ├─fonts      
│  └─js
├─routers       路由     
├─templates     html模板
└─utils         帮助函数
```

chitchat代码的运行方式，详见chitchat下面的README文件， 这里会简要地介绍chitchat的代码组织和项目要点：

## 代码入口

```go
func main() {
		utils.P("ChitChat", utils.Version(), "started at", utils.Config.Address)
	
		mux := http.NewServeMux()
	    
		// 静态资源
		files := http.FileServer(http.Dir(utils.Config.Static))  // 指向public
		mux.Handle("/static/", http.StripPrefix("/static/", files))
	

		// index
		mux.HandleFunc("/", routers.Index)

		// defined in route_auth.go
		mux.HandleFunc("/login", routers.Login)

    ...
	
		// 准备server
		server := &http.Server{
			Addr:           utils.Config.Address,
			Handler:        mux,
			ReadTimeout:    time.Duration(utils.Config.ReadTimeout * int64(time.Second)),
			WriteTimeout:   time.Duration(utils.Config.WriteTimeout * int64(time.Second)),
			MaxHeaderBytes: 1 << 20,
		}
    // 启动server
		server.ListenAndServe()
}
```

这里我们直接把Handler(也就是mux)传入到Server中， 而不是使用[HandleFunction](./helloworld.md#handle-function)的方式。

我们把静态资源Handle到/static/"路由(对应utils.Config.Static中的值)， 这样后面的html文件可以直接通过/static就能引用public中的css、js等资源。

utils.Config也就是我们的各种网站配置项是通过下面的方式导入的(定义在utils.go）：

```go
type Configuration struct {
	Address      string
	ReadTimeout  int64
	WriteTimeout int64
	Static       string
}

var Config Configuration

func init() {
	loadConfig()
	file, err := os.OpenFile("chitchat.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Failed to open log file", err)
	}
	Logger = log.New(file, "INFO ", log.Ldate|log.Ltime|log.Lshortfile)
}

func loadConfig() {
	file, err := os.Open("config.json")
	if err != nil {
		log.Fatalln("Cannot open config file", err)
	}
	decoder := json.NewDecoder(file)
	Config = Configuration{}
	err = decoder.Decode(&Config)
	if err != nil {
		log.Fatalln("Cannot get configuration from file", err)
	}
}
```

init函数会在utills模块被引入的时候自动执行， init调用了loadConfig读取了"config.json"文件里的信息， 解码并赋值给了全局变量Config。init函数除了加载配置项之外， 还准备了一个log文件。

回到main函数， 接着就是对不同的路由配置路由函数的过程。

## 路由函数

我们以路由函数Index为例：

```go
func Index(writer http.ResponseWriter, request *http.Request) {
		threads, err := data.Threads()
		if err != nil {
			utils.ErrorMessage(writer, request, "Cannot get threads")
		} else {
			_, err := utils.Session(writer, request)
			if err != nil {
				utils.GenerateHTML(writer, threads, "layout", "public.navbar", "index")
			} else {
				utils.GenerateHTML(writer, threads, "layout", "private.navbar", "index")
			}
		}
}
```

Index路由以及其他所有的路由(除了登陆)都会调用一下utils.Session来检查用户是不是登陆的(其实这个是在不断地重复代码， 也就是为什么很多框架会引入中间件的概念， 类似于这种登陆状态的检查， 都可以在中间件在完成， 不需要重复编写)

我们来看一下utils.Session函数：

```go
func Session(writer http.ResponseWriter, request *http.Request) (sess data.Session, err error) {
		cookie, err := request.Cookie("_cookie")
		if err == nil {
			sess = data.Session{Uuid: cookie.Value}
			if ok, _ := sess.Check(); !ok {
				err = errors.New("invalid session")
			}
		}
		return
}
```

Session函数做的事情， 无非就是从请求头中[获取cookie](./cookie.md#获取cookie)(也就是这里的_cookie), 如果不存在这个cookie，说明没有登陆， 会直接返回这个错误； "_cookie"中保存了用户的Uuid， 进而可以通过Check来查询这个Uuid对应的记录， 如果ok， 就会返回这个Session。

由于Check每次都要访问数据库， 所以在真实的事件中会在Redis中检查， 来提升速度， 同时减少数据库的压力。

既然有从请求体中获取cookie的过程， 自然有服务器端向客户端[写cookie](./cookie.md#设置cookie)的过程：

```go
func Authenticate(writer http.ResponseWriter, request *http.Request) {
		err := request.ParseForm()
		user, err := data.UserByEmail(request.PostFormValue("email"))
		if err != nil {
			utils.Danger(err, "Cannot find user")
		}
		if user.Password == data.Encrypt(request.PostFormValue("password")) {
			session, err := user.CreateSession()
			if err != nil {
				utils.Danger(err, "Cannot create session")
			}
			cookie := http.Cookie{
				Name:     "_cookie",
				Value:    session.Uuid,
				HttpOnly: true,
			}
			http.SetCookie(writer, &cookie)
			http.Redirect(writer, request, "/", 302)
		} else {
			http.Redirect(writer, request, "/login", 302)
		}

}
```

首先我么你会去检查是不是有给定email的用户存在。

接着， 会创建一个session， 存在在数据库中。

然后， 我们构建了一个Cookie对象， 并写到响应中。

完成所有操作后，会跳转到首页。

## 渲染视图

接着回到Index路由函数， 下面的代码会把数据(下面的threads)传递给模板完成渲染：

```go
utils.GenerateHTML(writer, threads, "layout", "public.navbar", "index")
```

GenerateHTML是一个通用函数：

```go
func GenerateHTML(writer http.ResponseWriter, data interface{}, filenames ...string) {
		var files []string
		for _, file := range filenames {
			files = append(files, fmt.Sprintf("templates/%s.html", file))
		}
	
		templates := template.Must(template.ParseFiles(files...))
		templates.ExecuteTemplate(writer, "layout", data)
}
```

上诉代码收集了所有传入到GenerateHTML的模板（也就是 "layout", "public.navbar", "index"文件）， 最后会传入到[template.ParseFiles](./template.md)中。

注意最后执行的模板是layout模板。

我们来看一下layou模板， 是怎样定义的：

```html
{{ define "layout" }}

<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <meta http-equiv="X-UA-Compatible" content="IE=9">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>ChitChat</title>
    <link href="/static/css/bootstrap.min.css" rel="stylesheet">
    <link href="/static/css/font-awesome.min.css" rel="stylesheet">
  </head>
  <body>
    {{ template "navbar" . }}

    <div class="container">
      
      {{ template "content" . }}
      
    </div> <!-- /container -->
    
    <script src="/static/js/jquery-2.1.1.min.js"></script>
    <script src="/static/js/bootstrap.min.js"></script>
  </body>
</html>

{{ end }}
```

layout定义了所有视图的架构。 在head部分引用了所有需要的静态资源， 比如这里bootstrap相关的css和js文件。

注意下面的代码：

```html
    <div class="container">
      
      {{ template "content" . }}
      
    </div>
```

这段代码会引用名为content的模板。比如index.html:

```html
{{ define "content" }}
<p class="lead">
  <a href="/thread/new">Start a thread</a> or join one below!
</p>

{{ range . }}
  <div class="panel panel-default">
    <div class="panel-heading">
      <span class="lead"> <i class="fa fa-comment-o"></i> {{ .Topic }}</span>
    </div>
    <div class="panel-body">
      Started by {{ .User.Name }} - {{ .CreatedAtDate }} - {{ .NumReplies }} posts.
      <div class="pull-right">
        <a href="/thread/read?id={{.Uuid }}">Read more</a>
      </div>
    </div>
  </div>
{{ end }}

{{ end }}
```

当然在处理其他路由的时候， 只要在GenerateHTML中传入响应的content模板就可以了