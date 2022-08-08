package main

import (
    "net"
    "sync"
    "net/url"
    "net/http"
    "strconv"
    "wstunnel/mlog"
    "github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

type tunnelServer struct {
    cnt         int
    listenUrl   string
    remoteUrl   string
}

func (wsserv *tunnelServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    wsconn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        mlog.ErrorLog("upgrade failed. err=%s.\n", err)
        return
    }   

    var udp = udpClient {
        wsconn : wsconn,
        remoteUrl : wsserv.remoteUrl,
    }

    go WSToUdp(&udp)
}

type udpClient struct {
    wsconn   *websocket.Conn
    dstAddr  net.UDPAddr
    remoteUrl string
    lock      sync.Mutex
    connMap   map[string]*net.UDPConn
}

func (udp *udpClient) createConnect(key net.UDPAddr) *net.UDPConn {
    url, _ := url.Parse(udp.remoteUrl)
    port, _ := strconv.Atoi(url.Port())
    rmtaddr := &net.UDPAddr{
                IP: net.ParseIP(url.Hostname()), 
                Port: port, 
            }
    
    udpconn, err := net.DialUDP(url.Scheme, nil, rmtaddr) 
    if err != nil {
        mlog.ErrorLog("connect to udp server failed. url=%s.\n", udp.remoteUrl);
        return nil
    }
    
    udp.dstAddr = *rmtaddr
    if udp.connMap == nil {
        udp.connMap = make(map[string]*net.UDPConn)
    }
    udp.connMap[key.String()] = udpconn
    return udpconn
}

func WSToUdp(udp *udpClient) {

    UdpToWS := func(key net.UDPAddr, udpconn *net.UDPConn, udp *udpClient) {
        defer func() { 
            delete(udp.connMap, key.String())
            udpconn.Close()
        }()

        mlog.InfoLog("tunnel server start UDP to WS. addr=%s.\n", key.String());
        for {
            buf := make([]byte, 2048)
            n, _, err := udpconn.ReadFromUDP(buf)
            if err != nil {
                mlog.ErrorLog("tunnel server read from udp conn failed. err=%s.\n", err)
                break;
            }

            mlog.DebugLog("tunnel server read from udp. data=%s.\n", buf[:n]);

            var msg = &TunnelMsg{Key: key, Data: buf[:n]}
            
            d, _ := msg.ToBytes()
            err = udp.wsconn.WriteMessage(websocket.BinaryMessage, d)
            if err != nil {
                mlog.ErrorLog("tunnel server write to ws conn failed. err=%s.\n", err)
                break;
            }
        }
    }

    realWSToUdp := func(done chan bool, udp *udpClient) {
        var msg = &TunnelMsg{}

        mlog.InfoLog("tunnel server start WS to UDP.\n");

        for {
            _, buf, err := udp.wsconn.ReadMessage()
            if err != nil {
                mlog.ErrorLog("tunnel server read from ws failed. err=%s.\n", err)
                break;
            }

            mlog.DebugLog("tunnel server read from ws. data=%s.\n", buf);

            msg.FromBytes(buf)
            conn, ok := udp.connMap[msg.Key.String()]
            if !ok {
                conn = udp.createConnect(msg.Key)
                if conn == nil {
                    continue
                }

                go UdpToWS(msg.Key, conn, udp)
            }

            // _, err = conn.WriteToUDP(msg.Data, &udp.dstAddr)
            _, err = conn.Write(msg.Data)
            if err != nil {
                mlog.ErrorLog("tunnel server write to udp failed. err=%s.\n", err)
                break;
            }
        }
        done <- true
    }

    done_WSToUdp := make(chan bool)

    go realWSToUdp(done_WSToUdp, udp)

    
    /** check if goroutine finish. When finish, restart it after sleep. */
    select {
        case <- done_WSToUdp:
    }

    udp.wsconn.Close()
    for _, v := range udp.connMap {
        v.Close()
    }
}

func runTunnelService(tunnelListenUrl string, connectUrl string) {
    url, _ := url.Parse(tunnelListenUrl)
    wsserv := tunnelServer{
                    listenUrl : tunnelListenUrl,
                    remoteUrl : connectUrl,
                 }
    
    http.Handle("/udp", &wsserv)

    mlog.InfoLog("tunnel listen address = %s\n", tunnelListenUrl)

    http.ListenAndServe(url.Host, nil)
}
