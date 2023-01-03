package front

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/rakyll/statik/fs"
)

// 静态文件服务器
type FileServer struct {
	http.Handler
	addr *net.TCPAddr
	*http.Server
}

func (svr *FileServer) Addr() string {
	return fmt.Sprintf("http://localhost:%v", svr.addr.Port)
}

func NewFileServer(fileDir string) *FileServer {
	fs := ensureFileSystem(fileDir)
	server := &FileServer{
		Handler: http.FileServer(fs),
	}
	return server
}

func ensureFileSystem(fileDir string) http.FileSystem {
	_, er := os.Stat(fileDir)
	if er == nil {
		return http.Dir(fileDir)
	}

	statikFS, err := fs.New()
	if err != nil {
		panic(err)
	}

	return statikFS
}

func (svr *FileServer) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Access-Control-Allow-Origin", "*")             //允许访问所有域
	rw.Header().Add("Access-Control-Allow-Headers", "Content-Type") //header 的类型
	svr.Handler.ServeHTTP(rw, r)
}

func (svr *FileServer) Start() {
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}
	tcpAddr := ln.Addr().(*net.TCPAddr)
	svr.addr = tcpAddr

	svr.Server = &http.Server{Handler: svr}
	go svr.Server.Serve(ln)
}

func (svr *FileServer) Stop(ctx context.Context) error {
	return svr.Server.Shutdown(ctx)
}
