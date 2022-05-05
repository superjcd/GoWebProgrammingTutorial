package main

import (
	"chitchat/routers"
	"chitchat/utils"
	"net/http"
	"time"
)

func main() {
	utils.P("ChitChat", utils.Version(), "started at", utils.Config.Address)

	mux := http.NewServeMux()
    
	// 静态资源
	files := http.FileServer(http.Dir(utils.Config.Static))  // 指向public
	mux.Handle("/static/", http.StripPrefix("/static/", files))

	//
	// all route patterns matched here
	// route handler functions defined in other files
	//

	// index
	mux.HandleFunc("/", routers.Index)
	// error
	mux.HandleFunc("/err", routers.Err)

	// defined in route_auth.go
	mux.HandleFunc("/login", routers.Login)
	mux.HandleFunc("/logout", routers.Logout)
	mux.HandleFunc("/signup", routers.Signup)
	mux.HandleFunc("/signup_account", routers.SignupAccount)
	mux.HandleFunc("/authenticate", routers.Authenticate)

	// defined in route_thread.go
	mux.HandleFunc("/thread/new", routers.NewThread)
	mux.HandleFunc("/thread/create", routers.CreateThread)
	mux.HandleFunc("/thread/post", routers.PostThread)
	mux.HandleFunc("/thread/read", routers.ReadThread)

	// starting up the server
	server := &http.Server{
		Addr:           utils.Config.Address,
		Handler:        mux,
		ReadTimeout:    time.Duration(utils.Config.ReadTimeout * int64(time.Second)),
		WriteTimeout:   time.Duration(utils.Config.WriteTimeout * int64(time.Second)),
		MaxHeaderBytes: 1 << 20,
	}
	server.ListenAndServe()
}
