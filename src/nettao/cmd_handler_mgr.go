package nettao

import (
    "container/list"
    "fmt"
    "reflect"
    "sync"
)

type CmdRange struct {
    CmdMin uint32
    CmdMax uint32
}

type ICmdHandler interface {
    Init(cmd uint32) error
    HandlePkg(ctx *ConnCtx, pkg *TaskPkg) error
}

type CmdHandlerBase struct {
    Cmd uint32
}

func (p *CmdHandlerBase) Init(cmd uint32) error {
    p.Cmd = cmd

    return nil
}

type HandlerInfo struct {
    handlers     map[uint32]ICmdHandler
    cmdRangeList *list.List // store CmdRange that CmdMin is not equal CmdMax
}

func NewHandlerInfo() *HandlerInfo {
    var hinfo *HandlerInfo

    hinfo = new(HandlerInfo)
    hinfo.handlers = make(map[uint32]ICmdHandler)
    hinfo.cmdRangeList = list.New()

    return hinfo
}

func (p *HandlerInfo) GetCmdHandler(cmd uint32) ICmdHandler {
    var ok bool
    var h ICmdHandler

    if h, ok = p.handlers[cmd]; ok {
        return h
    }
    for e := p.cmdRangeList.Front(); nil != e; e = e.Next() {
        if cmd >= e.Value.(CmdRange).CmdMin && cmd <= e.Value.(CmdRange).CmdMax {
            cmd = e.Value.(CmdRange).CmdMin
            break
        }
    }
    if h, ok = p.handlers[cmd]; ok {
        return h
    }

    return nil
}

type cmdHandlerMgr struct {
    handlerInfosLock sync.Mutex
    handlerInfos     map[string]*HandlerInfo
}

var mgr *cmdHandlerMgr

func RegisterCmdHandlers(name string, cmdHandlerType map[reflect.Type]CmdRange) error {
    mgr.handlerInfosLock.Lock()
    defer mgr.handlerInfosLock.Unlock()

    if _, ok := mgr.handlerInfos[name]; ok {
        panic(fmt.Errorf("RegisterCmdHandlers|%s already registered", name))
    }
    hinfo := NewHandlerInfo()
    for creatorType, cmdRange := range cmdHandlerType {
        var h ICmdHandler
        var cmd uint32

        h = reflect.New(creatorType).Interface().(ICmdHandler)
        cmd = cmdRange.CmdMin
        if cmdRange.CmdMin > cmdRange.CmdMax {
            panic(fmt.Errorf("RegisterCmdHandlers|%s CmdRange min=%d, max=%d", name, cmdRange.CmdMin, cmdRange.CmdMax))
        }
        if cmdRange.CmdMin < cmdRange.CmdMax {
            hinfo.cmdRangeList.PushBack(cmdRange)
        }
        h.Init(cmd)
        hinfo.handlers[cmd] = h
    }
    mgr.handlerInfos[name] = hinfo

    return nil
}

func GetHandlerInfo(name string) *HandlerInfo {
    mgr.handlerInfosLock.Lock()
    defer mgr.handlerInfosLock.Unlock()

    hinfo, _ := mgr.handlerInfos[name]

    return hinfo
}

func init() {
    mgr = new(cmdHandlerMgr)
    mgr.handlerInfos = make(map[string]*HandlerInfo)
}
