package pp

import (
    "encoding/binary"
    //"fmt"
    "nettao"
    "sync/atomic"
)

type PPHandler struct {
    nettao.CmdHandlerBase
}

var RecvPkgCount int32

func (p *PPHandler) HandlePkg(ctx *nettao.ConnCtx, pkg *nettao.TaskPkg) error {

    cmd := binary.BigEndian.Uint32(pkg.Data[4:])
    // fmt.Printf("[Nettao]HandlePkg|cmd=%d, pp msg=%s\n", cmd, pkg.Data[nettao.PKG_HEAD_LEN:])

    if 100 == cmd {
        binary.BigEndian.PutUint32(pkg.Data[4:], 101)
    } else {
        binary.BigEndian.PutUint32(pkg.Data[4:], 100)
    }
    ctx.SendQueue <- pkg

    atomic.AddInt32(&RecvPkgCount, 1)

    return nil
}
