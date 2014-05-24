package main

import (
    "flag"
    "fmt"
    "runtime"
    //"log"
    "net"
    "nettao"
    "os"
    "pp"
    "reflect"
)

func main() {

    flag.Parse()

    if 1 != flag.NArg() {
        fmt.Printf("Usage:pp_server conf_file\n", flag.Arg(0))
        return
    }

    nettao.LoadConfig(flag.Arg(0))

    runtime.GOMAXPROCS(2 * runtime.NumCPU())

    ipport := nettao.AppConf.String("ipport")
    //service := ":9190"
    listener, err := net.Listen("tcp", ipport)
    checkError(err)

    var cmdRange nettao.CmdRange

    cmdRange.CmdMin = 100
    cmdRange.CmdMax = 100
    cmdHandlerType := make(map[reflect.Type]nettao.CmdRange)
    cmdHandlerType[reflect.TypeOf(pp.PPHandler{})] = cmdRange
    nettao.RegisterCmdHandlers("pp", cmdHandlerType)

    ppHinfo := nettao.GetHandlerInfo("pp")
    if nil == ppHinfo {
        fmt.Fprintf(os.Stderr, "Fatal error: ppHinfo is nil\n")
        os.Exit(1)
    }

    for {
        conn, err := listener.Accept()
        if err != nil {
            continue
        }
        ctx := nettao.NewConnCtx(conn, ppHinfo)
        // run as a goroutine
        go nettao.HandleSession(ctx)
    }
}

func checkError(err error) {
    if err != nil {
        fmt.Fprintf(os.Stderr, "Fatal error: %s\n", err.Error())
        os.Exit(1)
    }
}
