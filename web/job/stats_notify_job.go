package job

import (
	"fmt"
	"net"
	"os"
	"time"
	"x-ui/logger"
	"x-ui/util/common"
	"x-ui/web/service"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type LoginStatus byte

const (
	LoginSuccess LoginStatus = 1
	LoginFail    LoginStatus = 0
)

type StatsNotifyJob struct {
	xrayService     service.XrayService
	inboundService  service.InboundService
	settingService  service.SettingService
	telegramService service.TelegramService
	bot             *tgbotapi.BotAPI
}

func NewStatsNotifyJob() *StatsNotifyJob {
	statsNotifyJob := new(StatsNotifyJob)
	statsNotifyJob.telegramService.InitI18n()
	return statsNotifyJob
}

func (j *StatsNotifyJob) SendMsgToTgbot(msg string) {
	//Telegram bot basic info
	tgBottoken, err := j.settingService.GetTgBotToken()
	if err != nil || tgBottoken == "" {
		logger.Warning("sendMsgToTgbot failed, GetTgBotToken fail:", err)
		return
	}
	tgBotid, err := j.settingService.GetTgBotChatId()
	if err != nil {
		logger.Warning("sendMsgToTgbot failed, GetTgBotChatId fail:", err)
		return
	}

	bot, err := tgbotapi.NewBotAPI(tgBottoken)
	if err != nil {
		fmt.Println("get tgbot error:", err)
		return
	}
	bot.Debug = true
	fmt.Printf("Authorized on account %s", bot.Self.UserName)
	info := tgbotapi.NewMessage(int64(tgBotid), msg)
	//msg.ReplyToMessageID = int(tgBotid)
	bot.Send(info)
}

// Here run is a interface method of Job interface
func (j *StatsNotifyJob) Run() {

	if !j.xrayService.IsXrayRunning() {
		return
	}
	var info string
	//get hostname
	name, err := os.Hostname()
	if err != nil {
		fmt.Println("get hostname error:", err)
		return
	}
	info = fmt.Sprintf("Hostname:%s\r\n", name)
	//get ip address
	var ip string
	netInterfaces, err := net.Interfaces()
	if err != nil {
		fmt.Println("net.Interfaces failed, err:", err.Error())
		return
	}

	for i := 0; i < len(netInterfaces); i++ {
		if (netInterfaces[i].Flags & net.FlagUp) != 0 {
			addrs, _ := netInterfaces[i].Addrs()

			for _, address := range addrs {
				if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
					if ipnet.IP.To4() != nil {
						ip = ipnet.IP.String()
						break
					} else {
						ip = ipnet.IP.String()
						break
					}
				}
			}
		}
	}
	info += fmt.Sprintf("IP:%s\r\n \r\n", ip)

	//get traffic
	inbouds, err := j.inboundService.GetAllInbounds()
	if err != nil {
		logger.Warning("StatsNotifyJob run failed:", err)
		return
	}

	for _, inbound := range inbouds {
		info += fmt.Sprintf("Node name:%s\r\nPort:%d\r\nUpload↑:%s\r\nDownload↓:%s\r\nTotal:%s\r\n", inbound.Remark, inbound.Port, common.FormatTraffic(inbound.Up), common.FormatTraffic(inbound.Down), common.FormatTraffic((inbound.Up + inbound.Down)))
		if inbound.ExpiryTime == 0 {
			info += fmt.Sprintf("Expire date:unlimited\r\n \r\n")
		} else {
			info += fmt.Sprintf("Expire date:%s\r\n \r\n", time.Unix((inbound.ExpiryTime/1000), 0).Format("2006-01-02 15:04:05"))
		}
	}
	j.SendMsgToTgbot(info)

	j.telegramService.NotifyUsersAboutToExpire()
}

func (j *StatsNotifyJob) UserLoginNotify(username string, ip string, time string, status LoginStatus) {
	if username == "" || ip == "" || time == "" {
		logger.Warning("UserLoginNotify failed,invalid info")
		return
	}
	var msg string
	//get hostname
	name, err := os.Hostname()
	if err != nil {
		fmt.Println("get hostname error:", err)
		return
	}
	if status == LoginSuccess {
		msg = fmt.Sprintf("Successfully logged-in to the panel\r\nHostname:%s\r\n", name)
	} else if status == LoginFail {
		msg = fmt.Sprintf("Login to the panel was unsuccessful\r\nHostname:%s\r\n", name)
	}
	msg += fmt.Sprintf("Time:%s\r\n", time)
	msg += fmt.Sprintf("Username:%s\r\n", username)
	msg += fmt.Sprintf("IP:%s\r\n", ip)
	j.SendMsgToTgbot(msg)
}

func (j *StatsNotifyJob) OnReceive() *StatsNotifyJob {
	tgBottoken, err := j.settingService.GetTgBotToken()
	if err != nil || tgBottoken == "" {
		logger.Warning("sendMsgToTgbot failed,GetTgBotToken fail:", err)
		return j
	}
	bot, err := tgbotapi.NewBotAPI(tgBottoken)
	if err != nil {
		fmt.Println("got tgbot error:", err)
		return j
	}
	j.bot = bot
	bot.Debug = false
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 10

	updates := bot.GetUpdatesChan(u)

	config := tgbotapi.NewSetMyCommands(service.CreateChatMenu()...)
	if _, err := bot.Request(config); err != nil {
		logger.Warning(err)
	}

	for update := range updates {
		if update.Message == nil {

			if update.CallbackQuery == nil {
				continue
			}
			callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "Success!")
			if _, err := bot.Request(callback); err != nil {
				logger.Error(err)
			}

			resp, del, upd := j.telegramService.HandleCallback(update.CallbackQuery)

			if resp != nil {
				if upd {
					updateMsg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, resp.Text)
					keyboard := resp.ReplyMarkup.(tgbotapi.InlineKeyboardMarkup)
					updateMsg.ReplyMarkup = &keyboard

					if _, err := bot.Request(updateMsg); err != nil {
						logger.Warning(err)
					}
				} else {
					_, err := bot.Send(resp)
					if err != nil {
						logger.Warning(err)
					}
				}
			}
			if del {
				deleteMsg := tgbotapi.NewDeleteMessage(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID)
				if _, err := bot.Request(deleteMsg); err != nil {
					logger.Error(err)
				}
			}
			continue
		}

		if update.Message.Photo != nil && j.telegramService.CanAcceptPhoto(update.Message.Chat.ID) {
			adminId, _ := j.settingService.GetTgBotChatId()

			fConfig := tgbotapi.NewForward(int64(adminId), update.Message.Chat.ID, update.Message.MessageID)
			if _, err := bot.Send(fConfig); err != nil {
				logger.Error(err)
			}
		}

		resp := j.telegramService.HandleMessage(update.Message)
		if resp == nil {
			continue
		}

		if _, err := bot.Send(resp); err != nil {
			logger.Error(err)
		}

	}
	return j
}

func (j *StatsNotifyJob) StopReceiving() {

	if j.bot != nil {
		logger.Debug("Stop receiving Telegram updates")
		j.bot.StopReceivingUpdates()
	}
}