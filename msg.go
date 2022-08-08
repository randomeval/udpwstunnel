package main

import (
    "bytes"
    "encoding/gob"
    "net"
    "wstunnel/mlog"
)

type TunnelMsg struct {
    Key  net.UDPAddr
    Data []byte
}

func (msg *TunnelMsg) ToBytes() ([]byte, error) {
    var buf bytes.Buffer
    enc := gob.NewEncoder(&buf)
    err := enc.Encode(*msg)
    if err != nil {
        mlog.ErrorLog("msg to bytes failed. key=%s, data=%s, err=%s.\n", msg.Key.String(), msg.Data, err)
        return nil, err
    }

    return buf.Bytes(), nil
}

func (msg *TunnelMsg) FromBytes(data []byte) error {
    buf := bytes.NewBuffer(data)
    dec := gob.NewDecoder(buf) 
    err := dec.Decode(msg)
    if err != nil {
        mlog.ErrorLog("bytes to msg failed. err=%s.\n", err)
        return err
    }
    
    return nil
}
