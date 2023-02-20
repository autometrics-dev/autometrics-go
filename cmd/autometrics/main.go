package main

import (
    "fmt"
    "os"
)

func main() {
    fmt.Println(len(os.Args), os.Args)
}
