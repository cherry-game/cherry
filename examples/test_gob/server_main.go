package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
)

func main() {

}

type MathService struct {
}

type Args struct {
	A, B int
}

func (m *MathService) Add(args Args, reply *int) error {
	*reply = args.A + args.B
	return nil
}

func server() {
	rpc.RegisterName("MathService", new(MathService))
	l, err := net.Listen("tcp", ":8088") //注意 “：” 不要忘了写
	if err != nil {
		log.Fatal("listen error", err)
	}
	rpc.Accept(l)
}

func client() {
	client, err := rpc.Dial("tcp", "localhost:8088")
	if err != nil {
		log.Fatal("dialing")
	}
	args := Args{A: 1, B: 2}
	var reply int
	err = client.Call("MathService.Add", args, &reply)
	if err != nil {
		log.Fatal("MathService.Add error", err)
	}
	fmt.Printf("MathService.Add: %d+%d=%d", args.A, args.B, reply)
}
