package router
import (
    "github.com/gin-gonic/gin"
    
    "github.com/NCKU-NASA/nasa-judge-lib/schema/user"

    "github.com/NCKU-NASA/nasa-judge-wireguard/middlewares/auth"
    "github.com/NCKU-NASA/nasa-judge-wireguard/utils/errutil"
    "github.com/NCKU-NASA/nasa-judge-wireguard/models/wireguard"
)

var router *gin.RouterGroup

func Init(r *gin.RouterGroup) {
    router = r
    router.POST("/update", auth.CheckIsTrust, update)
    router.GET("/get", auth.CheckIsTrust, get)
}

func update(c *gin.Context) {
    var userdata user.User
    err := c.ShouldBindJSON(&userdata)
    if err != nil {
        errutil.AbortAndStatus(c, 400)
        return
    }
    userdata = user.User{
        Username: userdata.Username,
    }
    userdata.Fix()
    userdata, err = user.GetUser(userdata)
    if err != nil {
        errutil.AbortAndError(c, &errutil.Err{
            Code: 409,
            Msg: "username not exist",
        })
        return
    }
    wireguard.Create(userdata)
    wireguard.Reload()
    c.String(200, "Success")
}

func get(c *gin.Context) {
    userdata := user.User{
        Username: c.Query("username"),
    }
    userdata.Fix()
    userdata, err := user.GetUser(userdata)
    if err != nil {
        errutil.AbortAndError(c, &errutil.Err{
            Code: 409,
            Msg: "username not exist",
        })
        return
    }
    confs := wireguard.GetPeerConfig(userdata)
    c.JSON(200, confs)
}
