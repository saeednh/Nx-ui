package controller

import (
	"strconv"
	"x-ui/database/model"
	"x-ui/web/service"

	"github.com/gin-gonic/gin"
)

type TelegramController struct {
	telegramService service.TelegramService
	xrayService     service.XrayService
}

func NewTelegramController(g *gin.RouterGroup) *TelegramController {
	a := &TelegramController{}
	a.initRouter(g)
	// a.startTask()
	return a
}

func (a *TelegramController) initRouter(g *gin.RouterGroup) {
	g = g.Group("/tgClients")

	g.POST("/list", a.getClients)
	g.POST("/sendMsg", a.sendMsg)
	g.POST("/sendMsgToAll", a.sendMsg)
	g.POST("/update", a.updateClient)
	g.POST("/registerClient", a.approveClient)
	g.POST("/renewClient", a.renewClient)
	g.POST("/del/:id", a.delClient)

	g.POST("/listMsgs", a.getClientMsgs)
	g.POST("/msg/del/:id", a.delMsg)
}

func (a *TelegramController) getClients(c *gin.Context) {
	// user := session.GetLoginUser(c)
	clients, err := a.telegramService.GetTgClients()
	if err != nil {
		jsonMsg(c, I18n(c, "pages.inbounds.toasts.obtain"), err)
		return
	}
	jsonObj(c, clients, nil)
}

func (a *TelegramController) getClientMsgs(c *gin.Context) {
	// user := session.GetLoginUser(c)
	msgs, err := a.telegramService.GetTgClientMsgs()
	if err != nil {
		jsonMsg(c, I18n(c, "pages.inbounds.toasts.obtain"), err)
		return
	}
	jsonObj(c, msgs, nil)
}

func (a *TelegramController) sendMsg(c *gin.Context) {
	// user := session.GetLoginUser(c)
	clientMsg := &model.TgClientMsg{}
	err := c.ShouldBind(clientMsg)
	if err != nil {
		jsonMsg(c, I18n(c, "pages.inbounds.addTo"), err)
		return
	}

	if clientMsg.ChatID > 0 {
		err = a.telegramService.SendMsgToTgBot(clientMsg.ChatID, clientMsg.Msg)
	} else {
		err = a.telegramService.BroadcastMsgToBot(clientMsg.Msg)
	}
	jsonMsgObj(c, I18n(c, "sendMsg"), clientMsg.ChatID, err)
	if err != nil {
		jsonMsg(c, I18n(c, "pages.inbounds.toasts.obtain"), err)
		return
	}
}

func (a *TelegramController) updateClient(c *gin.Context) {
	// user := session.GetLoginUser(c)
	client := &model.TgClient{}
	err := c.ShouldBind(client)
	if err != nil {
		jsonMsg(c, I18n(c, "pages.inbounds.addTo"), err)
		return
	}
	err = a.telegramService.UpdateClient(client)
	jsonMsgObj(c, I18n(c, "pages.inbounds.tg.update"), client, err)
	if err != nil {
		jsonMsg(c, I18n(c, "pages.inbounds.toasts.obtain"), err)
		return
	}
}

func (a *TelegramController) approveClient(c *gin.Context) {
	// user := session.GetLoginUser(c)
	client := &model.TgClient{}
	err := c.ShouldBind(client)
	if err != nil {
		jsonMsg(c, I18n(c, "pages.inbounds.addTo"), err)
		return
	}
	err = a.telegramService.RegisterClient(client)
	jsonMsgObj(c, I18n(c, "pages.inbounds.tg.update"), client, err)
	if err != nil {
		jsonMsg(c, I18n(c, "pages.inbounds.toasts.obtain"), err)
		return
	}
}

func (a *TelegramController) renewClient(c *gin.Context) {
	// user := session.GetLoginUser(c)
	client := &model.TgClient{}
	err := c.ShouldBind(client)
	if err != nil {
		jsonMsg(c, I18n(c, "pages.inbounds.addTo"), err)
		return
	}
	err = a.telegramService.RenewClient(client)
	jsonMsgObj(c, I18n(c, "pages.inbounds.tg.update"), client, err)
	if err != nil {
		jsonMsg(c, I18n(c, "pages.inbounds.toasts.obtain"), err)
		return
	}
}

func (a *TelegramController) delClient(c *gin.Context) {
	// user := session.GetLoginUser(c)
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		jsonMsg(c, I18n(c, "get"), err)
		return
	}
	err = a.telegramService.DeleteClient(id)
	jsonMsgObj(c, I18n(c, "delete"), id, err)
	if err != nil {
		jsonMsg(c, I18n(c, "pages.inbounds.toasts.obtain"), err)
		return
	}
}

func (a *TelegramController) delMsg(c *gin.Context) {
	// user := session.GetLoginUser(c)
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		jsonMsg(c, I18n(c, "get"), err)
		return
	}
	err = a.telegramService.DeleteMsg(id)
	jsonMsgObj(c, I18n(c, "delete"), id, err)
	if err != nil {
		jsonMsg(c, I18n(c, "pages.inbounds.toasts.obtain"), err)
		return
	}
}
