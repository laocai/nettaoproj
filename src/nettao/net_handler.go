package nettao

import (
    "bytes"
    "encoding/binary"
    "fmt"
    //"net"
    "time"
)

func HandleSession(ctx *ConnCtx) {
    go handleSessionRead(ctx)
    go handleSessionWrite(ctx)

    return
}

func handleSessionRead(ctx *ConnCtx) {
    var errState string

    ctx.connCloseWait.Add(1)
    defer func() {
        ctx.toStop = true
        if r := recover(); r != nil {
            fmt.Printf("[nettao]handleSessionRead|%s|Runtime error caught: %v\n", errState, r)
        }
        ctx.connCloseWait.Done()
        ctx.connCloseWait.Wait()
        ctx.conn.Close()
        fmt.Printf("[nettao]handleSessionRead|finish\n")
    }()

    for {
        errState = "handleSessionRead::readPkgHead"
        // package = head + body
        // first read the package head
        err := readPkgHead(ctx)
        if nil != err {
            panic(fmt.Errorf("peer addr = %s|%s", ctx.conn.RemoteAddr().String(), err.Error()))
            break
        }

        errState = "handleSessionRead::readPkgBody"
        // and than read the package body
        taskPkg, errReadPkgData := readPkgBody(ctx)
        if nil != errReadPkgData {
            panic(fmt.Errorf("peer addr = %s|%s", ctx.conn.RemoteAddr().String(), err.Error()))
            break
        }

        errState = "handleSessionRead::handlePkg"
        h := ctx.handlerInfo.GetCmdHandler(ctx.pkgHead.pkgHeadField.Cmd)
        if nil != h {
            // do handle the package
            errHandle := h.HandlePkg(ctx, taskPkg)
            if nil != errHandle {
                panic(fmt.Errorf("peer addr = %s|%s", ctx.conn.RemoteAddr().String(), errHandle.Error()))
                break
            }
        } else {
            panic(fmt.Errorf("peer addr = %s|%s|cmd=%d handler not found",
                ctx.conn.RemoteAddr().String(), ctx.pkgHead.pkgHeadField.Cmd))
            break
        }
    }
}

func handleSessionWrite(ctx *ConnCtx) {
    ctx.connCloseWait.Add(1)
    timeout := time.NewTimer(time.Second)

    defer func() {
        timeout.Stop()
        ctx.toStop = true
        if r := recover(); r != nil {
            fmt.Printf("[Nettao]handleSessionWrite|Runtime error caught: %v\n", r)
        }
        ctx.connCloseWait.Done()
    }()

    for {
        timeout.Reset(time.Second)

        select {
        case pkg := <-ctx.SendQueue:
            writeDataPkg(ctx, pkg)
        case <-timeout.C:
            fmt.Printf("[Nettao]handleSessionWrite|no data timeout\n")
        }

        if ctx.toStop {
            fmt.Printf("[Nettao]handleSessionWrite|stoped|peer addr= %s \n", ctx.conn.RemoteAddr().String())
            break
        }
    }
}

func readPkgHead(ctx *ConnCtx) error {
    headReaded := 0

    for {
        n, err := ctx.conn.Read(ctx.pkgHead.headBuf[headReaded:])
        if nil != err {
            //LogError("[Nettao]HandleSession::readPkgHead|peer addr=" + ctx.conn.RemoteAddr().String() +
            //    "|read pkg head err=" + err.Error() + "\n")
            return err
        }
        headReaded += n
        if headReaded == PKG_HEAD_LEN {
            break
        }
    }

    buf := bytes.NewBuffer(ctx.pkgHead.headBuf)
    errReadHead := binary.Read(buf, binary.BigEndian, &ctx.pkgHead.pkgHeadField)
    //fmt.Printf("[Nettao]HandleSession::readPkgHead|headReaded=%d|%v\n", headReaded, ctx.pkgHead.pkgHeadField)
    if nil != errReadHead {
        return errReadHead
    }

    return nil
}

func readPkgBody(ctx *ConnCtx) (*TaskPkg, error) {
    bodyReaded := 0
    taskPkg := NewTaskPkg(ctx)

    taskPkg.Data = make([]byte, ctx.pkgHead.pkgHeadField.Size)
    for {
        if ctx.pkgHead.pkgHeadField.Size == uint32(bodyReaded+PKG_HEAD_LEN) {
            break
        }
        n, err := ctx.conn.Read(taskPkg.Data[(bodyReaded + PKG_HEAD_LEN):])
        if nil != err {
            return nil, err
        }
        bodyReaded += n
    }
    copy(taskPkg.Data, ctx.pkgHead.headBuf)
    taskPkg.DataTime = time.Now()

    return taskPkg, nil
}

func writeDataPkg(ctx *ConnCtx, pkg *TaskPkg) error {
    dataWrited := 0
    dataLen := len(pkg.Data)
    for {
        if dataWrited == dataLen {
            break
        }
        n, err := ctx.conn.Write(pkg.Data[dataWrited:])
        if nil != err {
            ctx.toStop = true
            return err
        }
        dataWrited = dataWrited + n
        ctx.lastWTime = time.Now()
    }

    return nil
}
