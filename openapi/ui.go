package openapi

import (
	"bytes"
	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/tsinghua-cel/attacker-service/openapi/views"
	"github.com/tsinghua-cel/attacker-service/types"
	"net/http"
)

type uiHandler struct {
	backend types.ServiceBackend
}

func (api uiHandler) Home(c *gin.Context) {
	dash := views.DashboardInfo{
		CurSlot:           "100",
		StrategyCount:     "98",
		LatestBlockHeight: "90",
	}
	t1 := make([]views.StrategyWithReorgCount, 0)
	{
		t1 = append(t1, views.StrategyWithReorgCount{
			StrategyId:      "kkkkkkk111",
			ReorgCount:      "10",
			StrategyContent: "strategy content 1",
		})
	}
	t2 := make([]views.StrategyWithHonestLose, 0)
	{
		t2 = append(t2, views.StrategyWithHonestLose{
			StrategyId:        "kkkkkkk222",
			HonestLoseRateAvg: "0.1",
			StrategyContent:   "strategy content 2",
		})
	}
	t3 := make([]views.StrategyWithGreatHonestLose, 0)
	{
		t3 = append(t3, views.StrategyWithGreatHonestLose{
			StrategyId:           "kkkkkkk333",
			HonestLoseRateAvg:    "0.2",
			MaliciousLoseRateAvg: "0.1",
			StrategyContent:      "strategy content 3",
		})
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
