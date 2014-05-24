package nettao

import (
    "net"
    "sync"
    "time"
)

type ReadHeadInfo struct {
    headBuf      []byte
    pkgHeadField PkgHead
}

type ConnCtx struct {
    conn          net.Conn
    connCloseWait sync.WaitGroup
    toStop        bool
    lastRTime     time.Time // last read data from net
    lastWTime     time.Time // last write data to net
    pkgHead       ReadHeadInfo
    handlerInfo   *HandlerInfo

    SendQueue chan *TaskPkg
}

func NewConnCtx(conn net.Conn, handlerInfo *HandlerInfo) *ConnCtx {
    ctx := new(ConnCtx)
    ctx.toStop = false
    ctx.conn = conn
    ctx.lastRTime = time.Now()
    ctx.lastWTime = time.Now()
    ctx.pkgHead.headBuf = make([]byte, PKG_HEAD_LEN)
    ctx.handlerInfo = handlerInfo
    ctx.SendQueue = make(chan *TaskPkg)

    return ctx
}

// one package
type TaskPkg struct {
    ctx      *ConnCtx
    DataTime time.Time
    Data     []byte
}

func NewTaskPkg(ctx *ConnCtx) *TaskPkg {
    taskPkg := new(TaskPkg)
    taskPkg.ctx = ctx
    taskPkg.DataTime = time.Now()

    return taskPkg
}
