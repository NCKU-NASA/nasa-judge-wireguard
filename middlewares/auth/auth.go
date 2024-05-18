package auth

import (
//    "fmt"
    "github.com/gin-gonic/gin"
    "github.com/gin-contrib/sessions"
    "golang.org/x/exp/slices"
    
    "github.com/NCKU-NASA/nasa-judge-lib/schema/user"

    "github.com/NCKU-NASA/nasa-judge-wireguard/utils/errutil"
    "github.com/NCKU-NASA/nasa-judge-wireguard/utils/config"
)

func CheckSignIn(c *gin.Context) {
    if isSignIn, exist := c.Get("isSignIn"); !exist || !isSignIn.(bool) {
        errutil.AbortAndStatus(c, 401)
    }
}

func CheckIsAdmin(c *gin.Context) {
    if isAdmin, exist := c.Get("isAdmin"); !exist || !isAdmin.(bool) {
        errutil.AbortAndStatus(c, 401)
    }
}

func CheckIsTrust(c *gin.Context) {
    ip := c.ClientIP()
    if !slices.Contains(config.Trust, ip) {
        errutil.AbortAndStatus(c, 401)
    }
}

func AddMeta(c *gin.Context) {
    session := sessions.Default(c)
    username := session.Get("user")
    if username == nil {
        c.Set("isSignIn", false)
    } else {
        userdata := user.User{
            Username: username.(string),
        }
        userdata.Fix()
        if userdata.Username == "" {
            c.Set("isSignIn", false)
            return
        }
        var err error
        userdata, err = user.GetUser(userdata)
        if err != nil {
            c.Set("isSignIn", false)
        } else {
            c.Set("user", userdata)
            c.Set("isSignIn", true)
            c.Set("isAdmin", userdata.ContainGroup("admin"))
        }
    }
}
