package config

import (
    "net"
    "os"
    "strconv"
    "encoding/json"
)

var Debug bool
var Trust []string
var Port string
var Secret string
var Sessionname string
var WGpath string
var Server []string

func init() {
    loadenv()
    var err error
    debugstr, exists := os.LookupEnv("DEBUG")
    if !exists {
        Debug = false
    } else {
        Debug, err = strconv.ParseBool(debugstr)
        if err != nil {
            Debug = false
        }
    }
    truststr := os.Getenv("TRUST")
    if truststr == "" {
        Trust = []string{"127.0.0.1", "::1"}
    } else {
        var tmp []string
        err = json.Unmarshal([]byte(truststr), &tmp)
        if err != nil {
            panic(err)
        }
        for _, now := range tmp {
            ips, _ := net.LookupIP(now)
            for _, ip := range ips {
                Trust = append(Trust, ip.String())
            }
        }
    }
    serverstr := os.Getenv("SERVER")
    if serverstr != "" {
        err = json.Unmarshal([]byte(serverstr), &Server)
        if err != nil {
            panic(err)
        }
    }
    Port = os.Getenv("PORT")
    Secret = os.Getenv("SECRET")
    Sessionname = os.Getenv("SESSIONNAME")
    WGpath = os.Getenv("WGPATH")
}
