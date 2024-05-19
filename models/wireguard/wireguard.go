package wireguard

import (
    "fmt"
    "log"
    "os"
    "os/exec"
    "bufio"
    "encoding/base64"
    "strings"
    "path/filepath"
    "regexp"
    "net/netip"
    "sync"

    "github.com/google/uuid"
    
    "github.com/NCKU-NASA/nasa-judge-lib/utils/database"
    "github.com/NCKU-NASA/nasa-judge-lib/schema/user"

    "github.com/NCKU-NASA/nasa-judge-wireguard/utils/config"
    "github.com/NCKU-NASA/nasa-judge-wireguard/utils/ipcalc"
    "github.com/NCKU-NASA/nasa-judge-wireguard/utils/privatekey"
)

var lock *sync.RWMutex

type Wireguard struct {
    Key string `gorm:"primaryKey" json:"key"`
    UserID uint `json:"-"`
    User user.User `json:"user"`
}

func init() {
    lock = new(sync.RWMutex)
    _, err := exec.LookPath("wg")
    if err != nil {
        log.Panicln("Please install wireguard.")
    }
    database.GetDB().AutoMigrate(&Wireguard{})
    Reload()
}

func getConfig(name string) (conf string, data map[string]string) {
    f, err := os.Open(filepath.Join(config.WGpath, fmt.Sprintf("%s.conf", name)))
    if err != nil {
        log.Panicln(err)
    }
    defer f.Close()

    scanner := bufio.NewScanner(f)
    conf = ""
    data = make(map[string]string)
    ininterface := false
    rmcomment := regexp.MustCompile(`^#\s*`)
    spliteq := regexp.MustCompile(`\s*=\s*`)
    matchend := regexp.MustCompile(`^# BEGIN .*$`)
    for scanner.Scan() {
        now := scanner.Text()
        if now == "" {
            continue
        } else if now == "[Interface]" {
            ininterface = true
        } else if !ininterface {
            nowvar := rmcomment.ReplaceAllString(now, "")
            if spliteq.MatchString(nowvar) {
                data[spliteq.Split(nowvar, 2)[0]] = spliteq.Split(nowvar, 2)[1]
            }
        } else if matchend.MatchString(now) {
            break
        } else {
            if spliteq.MatchString(now) {
                data[spliteq.Split(now, 2)[0]] = spliteq.Split(now, 2)[1]
            }
        }
        conf += fmt.Sprintf("\n%s", now)
    }

    if err := scanner.Err(); err != nil {
        log.Panicln(err)
    }

    return
}

func setConfig(name, conf string) {
    f, err := os.OpenFile(filepath.Join(config.WGpath, fmt.Sprintf("%s.conf", name)), os.O_RDWR|os.O_TRUNC, 0600)
    if err != nil {
        log.Panicln(err)
    }
    _, err = f.WriteString(conf)
    if err != nil {
        log.Panicln(err)
    }
    f.Close()
    out, _ := exec.Command("wg-quick", "strip", name).Output()

    tmpname := uuid.New().String()
    f, err = os.Create(fmt.Sprintf("/tmp/%s", tmpname))
    if err != nil {
        log.Panicln(err)
    }
    _, err = f.Write(out)
    if err != nil {
        log.Panicln(err)
    }
    f.Close()
    defer os.Remove(fmt.Sprintf("/tmp/%s", tmpname))
    exec.Command("wg", "syncconf", name, fmt.Sprintf("/tmp/%s", tmpname)).Run()
}

func Create(userdata user.User) {
    data := Wireguard{
        UserID: userdata.ID,
        User: userdata,
    }
    result := database.GetDB().Model(&Wireguard{}).Preload("User").Where(data).First(&data)
    if result.Error == nil {
        return
    }
    data = Wireguard{
        Key: privatekey.Generate(),
        UserID: userdata.ID,
        User: userdata,
    }
    result = database.GetDB().Model(&Wireguard{}).Preload("User").Create(&data)
    if result.Error != nil {
        log.Panicln(result.Error)
    }
}

func Reload() bool {
    var datas []Wireguard
    result := database.GetDB().Model(&Wireguard{}).Preload("User").Find(&datas)
    if result.Error != nil {
        log.Panicln(result.Error)
    }

    lock.Lock()
    defer lock.Unlock()

    splitcom := regexp.MustCompile(`\s*,\s*`)
    
    for _, name := range config.Server {
        conf, servervar := getConfig(name)
        conf += "\n"
        addressesarr := splitcom.Split(servervar["Address"], -1)
        for _, data := range datas {
            addresses := []string{}
            for _, address := range addressesarr {
                nowipaddr := ipcalc.PrefixIPGet(netip.MustParsePrefix(address), int64(data.User.ID))
                addresses = append(addresses, netip.PrefixFrom(nowipaddr, nowipaddr.BitLen()).String())
            }
            conf += fmt.Sprintf(
`
# BEGIN %s
[Peer]
AllowedIPs = %s
PublicKey = %s
PersistentKeepalive = %s
# END %s`, 
                data.User.Username,
                strings.Join(addresses, ","),
                privatekey.Pubkey(data.Key),
                servervar["PersistentKeepalive"],
                data.User.Username,
            )
        }
        conf += "\n"
        setConfig(name, conf)
    }
    return true
}

func GetPeerConfig(userdata user.User) map[string]string {
    data := Wireguard{
        UserID: userdata.ID,
        User: userdata,
    }
    result := database.GetDB().Model(&Wireguard{}).Preload("User").Where(data).First(&data)
    if result.Error != nil {
        log.Panicln(result.Error)
    }

    lock.RLock()
    defer lock.RUnlock()
    
    confs := make(map[string]string)

    for _, name := range config.Server {
        _, servervar := getConfig(name)
        splitcom := regexp.MustCompile(`\s*,\s*`)
        addressesarr := splitcom.Split(servervar["Address"], -1)
        addresses := []string{}
        for _, address := range addressesarr {
            nowipaddr := ipcalc.PrefixIPGet(netip.MustParsePrefix(address), int64(data.User.ID))
            addresses = append(addresses, netip.PrefixFrom(nowipaddr, netip.MustParsePrefix(address).Bits()).String())
        }
        conf := fmt.Sprintf(
`[Interface]
Address = %s
PrivateKey = %s
`,
            strings.Join(addresses, ","),
            data.Key,
        )
        if _, ok := servervar["DNS"]; ok {
            conf += fmt.Sprintf(`DNS = %s
`,
                servervar["DNS"],
            )
        }
        conf += fmt.Sprintf(`
[Peer]
Endpoint = %s
AllowedIPs = %s
PublicKey = %s
`, 
            fmt.Sprintf("%s:%s", servervar["Host"], servervar["ListenPort"]),
            servervar["AllowedIPs"],
            privatekey.Pubkey(servervar["PrivateKey"]),
        )
        if _, ok := servervar["PersistentKeepalive"]; ok {
            conf += fmt.Sprintf(`PersistentKeepalive = %s
`,
                servervar["PersistentKeepalive"],
            )
        }
        confs[fmt.Sprintf("%s.conf", name)] = base64.StdEncoding.EncodeToString([]byte(conf))
    }
    return confs
}
