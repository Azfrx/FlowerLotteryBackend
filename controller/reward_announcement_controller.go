package controller

import (
	"encoding/json"
	"flower-lottery-backend/response"
	"flower-lottery-backend/service"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const rewardAnnouncementEventName = "reward-announcement"

func (x *GameController) RewardAnnouncements(c *gin.Context) {
	if x.announcements == nil {
		response.Error(c, http.StatusServiceUnavailable, 15001, "中奖播报服务暂不可用")
		return
	}
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		response.Error(c, http.StatusInternalServerError, 15002, "当前服务不支持流式响应")
		return
	}

	updates, recent, unsubscribe := x.announcements.Subscribe(rewardAnnouncementLastEventID(c))
	defer unsubscribe()

	c.Header("Content-Type", "text/event-stream; charset=utf-8")
	c.Header("Cache-Control", "no-cache, no-transform")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	c.Status(http.StatusOK)
	if _, err := io.WriteString(c.Writer, "retry: 3000\n\n"); err != nil {
		return
	}
	for _, announcement := range recent {
		if err := writeRewardAnnouncementEvent(c.Writer, announcement); err != nil {
			return
		}
	}
	flusher.Flush()

	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()
	for {
		select {
		case <-c.Request.Context().Done():
			return
		case announcement := <-updates:
			if err := writeRewardAnnouncementEvent(c.Writer, announcement); err != nil {
				return
			}
			flusher.Flush()
		case now := <-heartbeat.C:
			if _, err := fmt.Fprintf(c.Writer, ": heartbeat %d\n\n", now.Unix()); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

func rewardAnnouncementLastEventID(c *gin.Context) uint64 {
	return parseRewardAnnouncementLastEventID(c.GetHeader("Last-Event-ID"), c.Query("last_event_id"))
}

func parseRewardAnnouncementLastEventID(values ...string) uint64 {
	for _, value := range values {
		id, err := strconv.ParseUint(strings.TrimSpace(value), 10, 64)
		if err == nil {
			return id
		}
	}
	return 0
}

func writeRewardAnnouncementEvent(writer io.Writer, announcement service.RewardAnnouncement) error {
	payload, err := json.Marshal(announcement)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(
		writer,
		"id: %d\nevent: %s\ndata: %s\n\n",
		announcement.ID,
		rewardAnnouncementEventName,
		payload,
	)
	return err
}
