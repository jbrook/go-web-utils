package web

import (
    "fmt"
    "net/http"
    "time"

    "github.com/elazarl/go-bindata-assetfs"
    "github.com/julienschmidt/httprouter"
)

type FileServerConfig struct {
    Asset           func(name string) ([]byte, error)
    AssetDir        func(name string) ([]string, error)
    StaticFilesPath string
}

var staticFilePath string

func GetFileServer(config FileServerConfig) httprouter.Handle {
    fileServer := http.FileServer(&assetfs.AssetFS{
        Asset:    config.Asset,
        AssetDir: config.AssetDir,
        Prefix:   config.StaticFilesPath,
    })

    f := func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
        w.Header().Set("Cache-Control", "max-age=2592000") // 30 days

        r.URL.Path = ps.ByName("filepath")
        fileServer.ServeHTTP(w, r)
    }

    return f
}

func GetStaticPath(path string) string {
    return staticFilePath + path
}

func initStaticPath(root string) {
    staticFilePath = fmt.Sprintf("%sstatic/%d/", root, time.Now().Unix())
}