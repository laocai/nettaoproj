package main

import (
    "encoding/binary"
    "flag"
    "fmt"
    "runtime"
    //"log"
    "net"
    "nettao"
    "os"
    "pp"
    "reflect"
    "time"
)

const (
    PP_MSG string = "PP Cmd Ping Pong Msg"
)

func main() {

    flag.Parse()

    if 1 != flag.NArg() {
        fmt.Printf("Usage:pp_server conf_file\n", flag.Arg(0))
        return
    }

    nettao.LoadConfig(flag.Arg(0))

    ipport := nettao.AppConf.String("ipport")
    maxconn, _ := nettao.AppConf.Int("maxconn")
    maxpkgcount, _ := nettao.AppConf.Int("maxpkgcount")

    runtime.GOMAXPROCS(2 * runtime.NumCPU())

    var cmdRange nettao.CmdRange

    cmdRange.CmdMin = 100
    cmdRange.CmdMax = 101
    cmdHandlerType := make(map[reflect.Type]nettao.CmdRange)
    cmdHandlerType[reflect.TypeOf(pp.PPHandler{})] = cmdRange
    nettao.RegisterCmdHandlers("pp", cmdHandlerType)

    ppHinfo := nettao.GetHandlerInfo("pp")
    if nil == ppHinfo {
        fmt.Fprintf(os.Stderr, "Fatal error: ppHinfo is nil\n")
        os.Exit(1)
    }

    var connCount int
    for {
        conn, err := net.Dial("tcp", ipport)
        if err != nil {
            continue
        }
        ctx := nettao.NewConnCtx(conn, ppHinfo)
        // run as a goroutine
        go beginPingPong(ctx)
        go nettao.HandleSession(ctx)
        connCount++
        if connCount >= maxconn {
            break
        }
    }

    timeout := time.NewTimer(time.Second)
    for {
        timeout.Reset(time.Second)
        select {
        case <-timeout.C:
            fmt.Printf("Client do RecvPkgCount=%d\n", pp.RecvPkgCount)
        }
        if pp.RecvPkgCount >= int32(maxpkgcount) {
            break
        }
    }
    fmt.Printf("Client finish\n")
}

func beginPingPong(ctx *nettao.ConnCtx) {
    var pkgsize uint32
    var pkg *nettao.TaskPkg

    pkgsize = uint32(nettao.PKG_HEAD_LEN) + uint32(len(PP_MSG))
    pkg = nettao.NewTaskPkg(ctx)
    pkg.Data = make([]byte, pkgsize)
    binary.BigEndian.PutUint32(pkg.Data, pkgsize)
    binary.BigEndian.PutUint32(pkg.Data[4:], 100)
    copy(pkg.Data[nettao.PKG_HEAD_LEN:], PP_MSG)
    ctx.SendQueue <- pkg
}

func checkError(err error) {
    if err != nil {
        fmt.Fprintf(os.Stderr, "Fatal error: %s\n", err.Error())
        os.Exit(1)
    }
}
