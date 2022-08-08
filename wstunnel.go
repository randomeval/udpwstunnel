// wstunnel tool
package main

import (
    "flag"
    "wstunnel/mlog"
    "os"
    "os/signal"
    "syscall"
)


type Config struct {
    tunnelListenUrl  string     // Tunnel endpoint to listen.
    tunnelConnectUrl  string     // Tunnel endpoint to connect.
    listenUrl        string      
    connectUrl        string

    front_end  int
}

///////////////////////////////////////////////////////////////
// Global Variables

// Config info.
var gConfig Config
///////////////////////////////////////////////////////////////

func parseArgs(cfg *Config) bool {
    flag.StringVar(&cfg.tunnelListenUrl, "tl", "", "Address to tunnel listen.")
    flag.StringVar(&cfg.tunnelConnectUrl, "tc", "", "Address to tunnel connect.")
    flag.StringVar(&cfg.listenUrl, "l", "", "Address to listen.")
    flag.StringVar(&cfg.connectUrl, "c", "", "Address to connect.")
    
    flag.Parse()

    // TODO
    // 检查参数.
    return true
}

func run() chan struct{} {
    end := make(chan struct{})
    
    go func () {
        if len(gConfig.tunnelListenUrl) != 0 {
            runTunnelService(gConfig.tunnelListenUrl, gConfig.connectUrl)
        } else {
            runTunnelClient(gConfig.listenUrl, gConfig.tunnelConnectUrl)
        }

        close(end)
    }()

    return end;
}

func main() {
    // mlog.LogInit("./wstunnel.log", 10, 3, true, mlog.LEVEL_DEBUG) 
    mlog.LogInit("./wstunnel.log", 10, 3, true, mlog.LEVEL_INFO) 

    if !parseArgs(&gConfig) {
        // 失败.
        return ;
    }

    mlog.DebugLog("tunnel listen address = %s\n", gConfig.tunnelListenUrl)
    mlog.DebugLog("tunnel connect address = %s\n", gConfig.tunnelConnectUrl)
    mlog.DebugLog("listen address = %s\n", gConfig.listenUrl)
    mlog.DebugLog("connect address = %s\n", gConfig.connectUrl)
    
    // setup signal.
    sigChan := make(chan os.Signal)
    signal.Notify(sigChan, syscall.SIGINT)

    // init

    // run
    end := run();

    // wait to finish.
    select {
        case <-sigChan:
        case <-end:
    }

    // clean.

    mlog.DebugLog("wstunnel exit.\n");
}
