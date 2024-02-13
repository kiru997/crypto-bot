package background

import (
	"example.com/greetings/internal/service"
	"github.com/robfig/cron/v3"
)

func RunWorker(sv service.CompareService) {
	c := cron.New()

	c.AddFunc("@every 10m", sv.WatchTopChange)
	c.AddFunc("@every 1m", sv.RefreshConn)
	c.Start()

	sv.WatchTopChange()
	sv.SendNoti()
}
