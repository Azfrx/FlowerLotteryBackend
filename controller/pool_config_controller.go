package controller

import (
	"encoding/json"
	"flower-lottery-backend/response"
	"flower-lottery-backend/service"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const poolConfigEventName = "pool-config-updated"

func (x *GameController) PoolConfigUpdates(c *gin.Context) {
	if x.poolConfigs == nil {
		response.Error(c, http.StatusServiceUnavailable, 15003, "奖池配置推送服务暂不可用")
		return
	}
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		response.Error(c, http.StatusInternalServerError, 15004, "当前服务不支持流式响应")
		return
	}

	updates, recent, unsubscribe := x.poolConfigs.Subscribe()
	defer unsubscribe()
	c.Header("Content-Type", "text/event-stream; charset=utf-8")
	c.Header("Cache-Control", "no-cache, no-transform")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	c.Status(http.StatusOK)
	if _, err := io.WriteString(c.Writer, "retry: 3000\n\n"); err != nil {
		return
	}
	for _, update := range recent {
		if err := writePoolConfigEvent(c.Writer, update); err != nil {
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
		case update := <-updates:
			if err := writePoolConfigEvent(c.Writer, update); err != nil {
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

func writePoolConfigEvent(writer io.Writer, update service.PoolConfigUpdate) error {
	payload, err := json.Marshal(update)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(
		writer,
		"id: %d\nevent: %s\ndata: %s\n\n",
		update.ID,
		poolConfigEventName,
		payload,
	)
	return err
}
