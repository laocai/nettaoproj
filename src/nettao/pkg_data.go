package nettao

type PkgHead struct {
    // size == pkg_head + pkg_body
    Size uint32
    Cmd  uint32

    /* padding */
    Spare0 uint32
    Spare1 uint32
    Spare2 uint32
    Spare3 uint32
    Spare4 uint32
    Spare5 uint32
}

const (
    PKG_HEAD_LEN int = 32

    // sys command
    CMD_KEEP_ALIVE uint32 = 1
)
