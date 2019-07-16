package main

import (
	"flag"
	"time"
)

var (
	mode     string
	addr     string
	file     string
	out      string
	interval int64
)

func init() {
	flag.StringVar(&mode, "m", "server", "client or server")
	flag.StringVar(&addr, "a", "[::1]:2000", "address")
	flag.StringVar(&file, "f", "", "file to upload, default is stdin")
	flag.StringVar(&out, "o", "", "the output folder, default isstdout")
	flag.Int64Var(&interval, "i", int64(time.Millisecond), "the interval")
	flag.Parse()
}

func main() {
	// switch mode {
	// case "client":
	// 	client()
	// 	break
	// case "server":
	// 	server()
	// 	break
	// }
}
