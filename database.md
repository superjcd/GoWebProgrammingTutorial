## 数据库链接

```go
import (
 "fmt"
 "database/sql"
 _ "github.com/lib/pq"
)

var Db *sql.DB

func init() {
    var err error
		 Db, err = sql.Open("postgres", "user=gwp dbname=gwp password=gwp
		 sslmode=disable")
		 if err != nil {
		 panic(err)
 }
}
```

使用database/sql库来链接我们的PG数据库； 通过sql.Open来建立一个数据库连接池， 第一个字段是数据库的类型， 第二个字段则是一串验证参数， 返回的Db将会被全局使用。

## CRUD

增删改查(CRUD)可以说是最为常见的数据库操作， 接着我们一起来看一下使用database/sql是如何实现CRUD的方式， 基本上会是原生sql + 变量（占位符）的形式。 

在开始前， 我们首先需要创建一张表， 这里以Post表为例：

 

```sql
create table posts (
  id      serial primary key,
  content text,
  author  varchar(255)
);
```

表对应的结构体是：

```go
type Post struct {
	Id      int
	Content string
	Author  string
}
```

### Create

```go
func (post *Post) Create() (err error) {
	 statement := "insert into posts (content, author) values ($1, $2) returning id "
	 stmt, err := Db.Prepare(statement)
	 if err != nil {
	 return
	 }
	 defer stmt.Close()
	 err = stmt.QueryRow(post.Content, post.Author).Scan(&post.Id)
	 if err != nil {
	 return
	 }
	 return
}
```

上面的Create函数有三个值得注意的事情：

- 我们在函数定义过程中， 把post结构体作为我们的接收者；
- 然后， 插入数据的时候， 我们先准备了一条Insert插入语句， 语句中有两个变量(占位符)$1, $2， 分别对应post的content和author变量
- 使用QueryRow将结果扫描到了post的Id字段

在使用的时候， 可以是下面这样:

```go
post := Post{Content: "Hello World!", Author: "Sau Sheong"}

post.Create()
```

### Read

```go
func GetPost(id int) (post Post, err error) {
	 post = Post{}
	 err = Db.QueryRow("select id, content, author from posts where id =$1", id).Scan(&post.Id, &post.Content, &post.Author)
	 return
}
```

这里我们直接使用了Db.QueryRow， 而没有去Prepare一个sql； 如果sql语句里没有returing语句， 那么Scan会扫描sql语句中的的三个变量的返回值， 也就是这里的id, content和author

### Update

```go
func (post *Post) Update() (err error) {
 _, err = Db.Exec("update posts set content = $2, author = $3 where id = $1", post.Id, post.Content, post.Author)
 return
}
```

这里我们直接使用Db.Exec来运行sql命令,  注意变量的顺序（$1对应第一个变量）

### Delete

```go
func (post *Post) Delete() (err error) {
 _, err = Db.Exec("delete from posts where id = $1", post.Id)
 return
}
```

### 获取所有的post

上面我们的读取只是通过QueryRow读取了单行的数据， 接着我们试着读取多个post（多行数据）

```go
func Posts(limit int) (posts []Post, err error) {
		 rows, err := Db.Query("select id, content, author from posts limit $1", limit)
		 if err != nil {
		 return
		 }
		 for rows.Next() {
			 post := Post{}
			 err = rows.Scan(&post.Id, &post.Content, &post.Author)
			 if err != nil {
			 return
		 }
		 posts = append(posts, post)
 }
 rows.Close()
```

首先我们通过Db.Query来获取查询结果， 然后遍历这个结果：

```go
for rows.Next() {
    ...
    rows.Scan(&post.Id, &post.Content, &post.Author)  
    ...
}
```

## 构建关系

常见的数据库关系有：

- 1 对 1
- 1 对 多
- 多 对 1
- 多对多

### 准备数据

我们在数据库建立一对多的数据表格：

```sql
create table posts (
   id serial primary key,
   content text,
   author varchar(255)
);

create table comments (
   id serial primary key,
   content text,
   author varchar(255),
   post_id integer references posts(id)
);
```

外键（post_id）通过定义在了多的一侧。

### Go中的结构体

```go
type Post struct {
	 Id int
	 Content string
	 Author string
	 Comments []Comment
}

type Comment struct {
	 Id int
	 Content string
	 Author string
	 Post *Post
}

```

### 创建一个关系

我们对评论对象创建一个Create方法

```go
func (comment *Comment) Create() (err error) {
 if comment.Post == nil {
     err = errors.New("Post not found")
      return
 }
 err = Db.QueryRow("insert into comments (content, author, post_id) values ($1, $2, $3) returning id", comment.Content, comment.Author, comment.Post.Id).Scan(&comment.Id)
      return
}
```

首先我们需要确定comment的Post属性是非空的， 空的话直接返回错

### 获取关系

```go
func GetPost(id int) (post Post, err error) {
		 post = Post{}
		 post.Comments = []Comment{}
		 err = Db.QueryRow("select id, content, author from posts where id =
		  $1", id).Scan(&post.Id, &post.Content, &post.Author)
		
		 rows, err := Db.Query("select id, content, author from comments where
		  post_id = $1", id)
		 if err != nil {
		 return
		 }
		 for rows.Next() {
		     comment := Comment{Post: &post}
		     err = rows.Scan(&comment.Id, &comment.Content, &comment.Author)
		     if err != nil {
		         return
		 }
		     post.Comments = append(post.Comments, comment)
		 }
		 rows.Close()
		 return
}
```


很显然， 使用原生的database/sql库去操作数据库， 需要通过编写原生sql的形式进行操作， 所以并不是十分方便。所以在CRUD的场景下， 使用ORM库(比如gorm)应该会是一个更方便的选项。

但是， 在处理非常复杂的查询的时候， 还是直接使用sql， 使用ORM的方式反而会惨不忍睹。好在几乎所有的第三方库都会支持用户编写原生的sql代码。