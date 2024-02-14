package background

import (
	"example.com/greetings/internal/service"
	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
)

func RunWorker(sv service.CompareService) {
	c := cron.New()

	c.AddFunc("@every 5m", func() {
		sv.WatchTopChange(uuid.NewString())
	})
	c.AddFunc("@every 1m", sv.RefreshConn)
	c.Start()

	go sv.WatchTopChange(uuid.NewString())
	sv.SendNoti()
}
