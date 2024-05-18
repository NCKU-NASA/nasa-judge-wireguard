package privatekey

import (
    //"fmt"
    "log"
    "io"
    "os/exec"
    "strings"
)

func Pubkey(privkey string) string {
    cmd := exec.Command("wg", "pubkey")
    stdin, err := cmd.StdinPipe()
    if err != nil {
        log.Panicln(err)
    }
    stdout, err := cmd.StdoutPipe()
    if err != nil {
        log.Panicln(err)
    }
    io.WriteString(stdin, privkey)
    stdin.Close()
    if err = cmd.Start(); err != nil {
        log.Panicln(err)
    }
    out, _ := io.ReadAll(stdout)
    if err = cmd.Wait(); err != nil {
        log.Panicln(err)
    }
    return strings.TrimSpace(string(out))
}

func Generate() string {
    out, _ := exec.Command("wg", "genkey").Output()
    return strings.TrimSpace(string(out))
}
