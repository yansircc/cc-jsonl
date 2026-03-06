//go:build ignore

package main

import (
    "embed"
    "fmt"
    "io/fs"
)

//go:embed all:ui/dist
var uiFS embed.FS

func main() {
    count := 0
    fs.WalkDir(uiFS, ".", func(path string, d fs.DirEntry, err error) error {
        if err != nil { return nil }
        if !d.IsDir() {
            count++
            if count <= 20 {
                fmt.Println(path)
            }
        }
        return nil
    })
    fmt.Printf("\n总共 %d 个文件\n", count)
    
    // 测试 Sub
    distFS, _ := fs.Sub(uiFS, "ui/dist")
    f, err := distFS.Open("_app/immutable/chunks/_aMjao2f.js")
    if err != nil {
        fmt.Printf("Open error: %v\n", err)
    } else {
        buf := make([]byte, 80)
        n, _ := f.Read(buf)
        fmt.Printf("OK: %s\n", string(buf[:n]))
        f.Close()
    }
}
