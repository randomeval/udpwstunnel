package main

import (
    "net"
    "fmt"
    "sync"
    "net/url"
    "net/http"
    "strconv"
    "time"
    "wstunnel/mlog"
    "github.com/gorilla/websocket"
)

type tunnelClient struct {
    lock        sync.Mutex
    remoteUrl   string
    conn        *websocket.Conn
}

func (wsconn *tunnelClient) Connect(connUrl string) error {
    var dialer = &websocket.Dialer{}
    c, resp, err := dialer.Dial(connUrl, nil);

    wsconn.remoteUrl = connUrl

    if err != nil {
        mlog.ErrorLog("connect to websocket failed. remote=%s, err=%s.\n", connUrl, err)
        return err 
    }

    resp.Body.Close()
    if resp.StatusCode != http.StatusSwitchingProtocols {
        mlog.ErrorLog("switch proto failed. remote=%s.\n", connUrl)
        return fmt.Errorf("switch proto failed.")
    }
    
    wsconn.conn = c
    return nil
}

func (wsconn *tunnelClient) ReConnect() error {
    wsconn.lock.Lock()
    defer wsconn.lock.Unlock()

    if wsconn.conn == nil {
        wsconn.Connect(wsconn.remoteUrl)

        if wsconn.conn == nil {
            return fmt.Errorf("Can't connet to remote ws. url=%s.", wsconn.remoteUrl)
        }
    }
    return nil
}

func (wsconn *tunnelClient) CloseConn() {
    if wsconn.conn != nil {
        wsconn.conn.Close()
        wsconn.conn = nil
    }
}

func (wsconn *tunnelClient) SendData(addr *net.UDPAddr, data []byte) error {
    if wsconn.ReConnect() != nil {
        return fmt.Errorf("Can't connet to remote ws. url=%s.", wsconn.remoteUrl)
    }

    var msg = &TunnelMsg{Key: *addr, Data: data}
    
    buf, _ := msg.ToBytes()
    err := wsconn.conn.WriteMessage(websocket.BinaryMessage, buf)
    if err != nil {
        wsconn.CloseConn()
        return err
    }
    return nil
}

func (wsconn *tunnelClient) RecvData() (net.UDPAddr, []byte, error) {
    if wsconn.ReConnect() != nil {
        return net.UDPAddr{}, nil, fmt.Errorf("Can't connet to remote ws. url=%s.", wsconn.remoteUrl)
    }

    var msg = &TunnelMsg{}
    
    _, buf, err := wsconn.conn.ReadMessage()
    if err != nil {
        wsconn.CloseConn()
        return net.UDPAddr{}, nil, err
    }

    msg.FromBytes(buf)
        
    return msg.Key, msg.Data, nil
}

func runTunnelClient(listenUrl string, tunnelClientUrl string) {
    wsconn := &tunnelClient{}
    wsconn.Connect(tunnelClientUrl)

    mlog.DebugLog("tunnel listen address = %s\n", listenUrl)

    url, _ := url.Parse(listenUrl)
    port, _ := strconv.Atoi(url.Port())
    addr := &net.UDPAddr{
                IP: net.ParseIP(url.Hostname()), 
                Port: port, 
            }
    
    conn, err := net.ListenUDP(url.Scheme, addr)
    if err != nil {
        mlog.ErrorLog("tunnel client listen UDP failed. address=%s, err=%s.\n", listenUrl, err)
        return ;
    }

    mlog.ErrorLog("client listen UDP OK. address=%s.\n", listenUrl)
    
    defer conn.Close()

    UdpToWS := func(done chan bool, udpconn *net.UDPConn, wsconn *tunnelClient) {
        mlog.InfoLog("tunnel client start UDP to WS.\n");
        for {
            msg := make([]byte, 2048)
            n, rmt, err := udpconn.ReadFromUDP(msg)
            if err != nil {
                mlog.ErrorLog("tunnel client read from udp failed. err=%s.\n", err)
                break;
            }

            mlog.DebugLog("tunnel client read from udp. data=%s, addr=%s.\n", msg[:n], rmt.String());

            err = wsconn.SendData(rmt, msg[:n])
            if err != nil {
                mlog.ErrorLog("tunnel client send to ws server failed. err=%s.\n", err)
                break;
            }
        }
        done <- true
    }

    WSToUdp := func(done chan bool, udpconn *net.UDPConn, wsconn *tunnelClient) {
        mlog.InfoLog("tunnel client start WS to UDP.\n");
        for {
            addr, data, err := wsconn.RecvData()
            if err != nil {
                mlog.ErrorLog("tunnel client recv from ws server failed. err=%s.\n", err)
                break;
            }

            mlog.DebugLog("tunnel client read from ws. data=%s, addr=%s.\n", data, addr.String());

            _, err = udpconn.WriteToUDP(data, &addr)
            if err != nil {
                mlog.ErrorLog("tunnel client write to udp failed. err=%s.\n", err)
                break;
            }
        }
        done <- true
    }

    done_UdpToWS := make(chan bool)
    done_WSToUdp := make(chan bool)

    go UdpToWS(done_UdpToWS, conn, wsconn)
    go WSToUdp(done_WSToUdp, conn, wsconn)

    cnt_UdpToWS := 1
    cnt_WSToUdp := 1
    
    /** check if goroutine finish. When finish, restart it after sleep. */
    for {
        select {
            case <- done_UdpToWS:
                time.Sleep(time.Duration(cnt_UdpToWS) * time.Second)
                cnt_UdpToWS++
                go UdpToWS(done_UdpToWS, conn, wsconn)
            case <- done_WSToUdp:
                time.Sleep(time.Duration(cnt_WSToUdp) * time.Second)
                cnt_WSToUdp++
                go WSToUdp(done_WSToUdp, conn, wsconn)
        }
    }
}


