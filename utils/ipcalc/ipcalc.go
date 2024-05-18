package ipcalc
import (
    "log"
    "net/netip"
    "math/big"
)

func PrefixIPGet(prefix netip.Prefix, index int64) netip.Addr {
    prefix = prefix.Masked()
    prefixnum := new(big.Int)
    prefixnum.GobDecode(append([]byte{2}, prefix.Addr().AsSlice()...))
    ipnum := new(big.Int)
    ipnum = ipnum.Add(prefixnum, big.NewInt(index))
    ipbyte, _ := ipnum.GobEncode()
    if ip, ok := netip.AddrFromSlice(ipbyte[1:]); ok {
        if !prefix.Contains(ip) {
            log.Panicln("PrefixIPGet index out of range")
        }
        return ip
    } else {
        log.Panicln("PrefixIPGet error")
    }
    return netip.Addr{}
}
