package api

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/panjjo/gosip/db"
	"github.com/panjjo/gosip/m"
	sipapi "github.com/panjjo/gosip/sip"
)

// @Summary     通道对讲
// @Description 对讲一个通道最多存在一个流
// @Tags        streams
// @Accept      x-www-form-urlencoded
// @Produce     json
// @Param       id     path     string true  "通道id"
// @Success     0      {object} sipapi.Streams
// @Failure     1000 {object} string
// @Failure     1001 {object} string
// @Failure     1002 {object} string
// @Failure     1003 {object} string
// @Router      /channels/{id}/start_talk [post]
func StartTalk(c *gin.Context) {
	channelid := c.Param("id")
	pm := &sipapi.Streams{S: time.Time{}, E: time.Time{}, ChannelID: channelid, Ttag: db.M{}, Ftag: db.M{}}
	res, err := sipapi.SipTalk(pm)
	if err != nil {
		m.JsonResponse(c, m.StatusParamsERR, err.Error())
		return
	}
	m.JsonResponse(c, m.StatusSucc, res)
}
