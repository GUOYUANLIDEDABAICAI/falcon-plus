package cron

import (
	"bytes"
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/open-falcon/falcon-plus/common/model"
	"github.com/open-falcon/falcon-plus/modules/alarm/api"
	"github.com/open-falcon/falcon-plus/modules/alarm/g"
	"github.com/open-falcon/falcon-plus/modules/alarm/redi"
	"github.com/toolkits/net/httplib"
	"strings"
	"time"
)

func HandleAlarmCallback(event *model.Event, action *api.Action) {
	teams := action.Uic
	ims := []string{}

	if teams != "" {
		ims = api.ParseTeams(teams)
	}
	message := AlarmCallback(event, action, ims)
}

func AlarmCallback(event *model.Event, action *api.Action, tos []string) string {
	if action.Url == "" {
		return "AlarmCallback url is blank"
	}

	L := make([]string, 0)
	if len(event.PushedTags) > 0 {
		for k, v := range event.PushedTags {
			L = append(L, fmt.Sprintf("%s:%s", k, v))
		}
	}

	tags := ""
	if len(L) > 0 {
		tags = strings.Join(L, ",")
	}

	req := httplib.Post(g.Config().Api.AlarmCallback).SetTimeout(3*time.Second, 20*time.Second)

	// 构建请求体
	data := map[string]interface{}{
		"endpoint":  event.Endpoint,
		"metric":    event.Metric(),
		"status":    event.Status,
		"step":      event.CurrentStep,
		"priority":  event.Priority(),
		"leftValue": event.LeftValue,
		"time":      event.FormattedTime(),
		"tplId":     event.TplId(),
		"expId":     event.ExpressionId(),
		"straId":    event.StrategyId(),
		"tags":      tags,
		"tos":       tos,
	}

	body, err := json.Marshal(data)
	if err != nil {
		log.Errorf("构建请求体失败：%v", err)
		return
	}
	// 设置请求体和请求头
	req.Body(body)
	req.Header("Content-Type", "application/json")
	// 发送请求
	resp, e := req.String()

	success := "success"
	if e != nil {
		log.Errorf("AlarmCallback fail, action:%v, event:%s, error:%s", action, event.String(), e.Error())
		success = fmt.Sprintf("fail:%s", e.Error())
	}
	message := fmt.Sprintf("curl %s %s. resp: %s", action.Url, success, resp)
	log.Debugf("AlarmCallback to url:%s, event:%s, resp:%s", action.Url, event.String(), resp)

	return message
}
