package openapi

import (
	"bytes"
	"fmt"
	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/tsinghua-cel/attacker-service/dbmodel"
	"github.com/tsinghua-cel/attacker-service/openapi/views"
	"github.com/tsinghua-cel/attacker-service/types"
	"net/http"
	"strconv"
)

type uiHandler struct {
	backend types.ServiceBackend
}

func (api uiHandler) Home(c *gin.Context) {
	limit := 10
	dash := views.DashboardInfo{}
	curSlot := api.backend.GetCurSlot()
	stCount := dbmodel.GetStrategyCount()
	latest, err := api.backend.GetLatestBeaconHeader()
	if err != nil {
		log.WithError(err).Error("Could not get latest beacon header")
	} else {
		dash.LatestBlockHeight = latest.Header.Message.Slot
	}
	dash.CurSlot = strconv.FormatInt(int64(curSlot), 10)
	dash.StrategyCount = strconv.FormatInt(stCount, 10)

	t1 := make([]views.StrategyWithReorgCount, 0)
	{
		list := dbmodel.GetStrategyListByReorgCount(limit)
		for _, s := range list {
			t1 = append(t1, views.StrategyWithReorgCount{
				StrategyId:      s.UUID,
				ReorgCount:      strconv.FormatInt(int64(s.ReorgCount), 10),
				StrategyContent: s.Content,
			})
		}

	}

	t2 := make([]views.StrategyWithHonestLose, 0)
	{
		list := dbmodel.GetStrategyListByHonestLoseRateAvg(limit)
		for _, s := range list {
			rate := strconv.FormatFloat(s.HonestLoseRateAvg*100, 'f', 4, 64)
			t2 = append(t2, views.StrategyWithHonestLose{
				StrategyId:        s.UUID,
				HonestLoseRateAvg: fmt.Sprintf("%s%%", rate),
				StrategyContent:   s.Content,
			})
		}
	}
	t3 := make([]views.StrategyWithGreatHonestLose, 0)
	{
		list := dbmodel.GetStrategyListByGreatLostRatio(limit)
		for _, s := range list {
			rate1 := strconv.FormatFloat(s.HonestLoseRateAvg*100, 'f', 4, 64)
			rate2 := strconv.FormatFloat(s.AttackerLoseRateAvg*100, 'f', 4, 64)
			ratio := s.HonestLoseRateAvg / s.AttackerLoseRateAvg
			rate_ratio := strconv.FormatFloat(ratio*100, 'f', 4, 64)
			t3 = append(t3, views.StrategyWithGreatHonestLose{
				StrategyId:           s.UUID,
				HonestLoseRateAvg:    fmt.Sprintf("%s%%", rate1),
				MaliciousLoseRateAvg: fmt.Sprintf("%s%%", rate2),
				Ratio:                fmt.Sprintf("%s%%", rate_ratio),
				StrategyContent:      s.Content,
			})
		}
	}
	data, _ := renderHtml(c, views.MakeStrategy("BunnyFinder Testing View", dash, t1, t2, t3))
	c.Data(http.StatusOK, "text/html; charset=utf-8", data)

}

// / This function will render the templ component into
// / a gin context's Response Writer
func render(c *gin.Context, status int, template templ.Component) error {
	c.Status(status)
	return template.Render(c.Request.Context(), c.Writer)
}

func renderHtml(c *gin.Context, template templ.Component) ([]byte, error) {
	html_buffer := bytes.NewBuffer(nil)

	err := template.Render(c, html_buffer)
	if err != nil {
		log.WithError(err).Error("Could not render index")
		return []byte{}, err
	}

	return html_buffer.Bytes(), nil
}
