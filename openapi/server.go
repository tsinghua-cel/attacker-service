package openapi

import (
	"fmt"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/tsinghua-cel/attacker-service/config"
	_ "github.com/tsinghua-cel/attacker-service/docs"
	"github.com/tsinghua-cel/attacker-service/types"
)

type OpenAPI struct {
	backend types.ServiceBackend
	conf    *config.Config
}

// generate swagger docs for the api in one step
// @title Attacker Service API
// @version 1
// @description This is the attacker service API server.
// @host localhost:20001
// @BasePath /v1
// @accept json

func NewOpenAPI(backend types.ServiceBackend, conf *config.Config) *OpenAPI {
	return &OpenAPI{backend: backend, conf: conf}
}

func (s *OpenAPI) Start() {
	go s.startHttp(s.conf.HttpPort + 1)
}

func (s *OpenAPI) startHttp(port int) {
	router := gin.Default()
	router.Use(cors())
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	// 创建v1组
	v1 := router.Group("/v1")
	{
		// 在v1这个分组下，注册路由
		v1.GET("/duties/:epoch", apiHandler{backend: s.backend}.GetDutiesByEpoch)
		v1.GET("/reward/:epoch", apiHandler{backend: s.backend}.GetRewardByEpoch)
		v1.GET("/strategy", apiHandler{backend: s.backend}.GetStrategy)
		v1.POST("/update-strategy", apiHandler{backend: s.backend}.UpdateStrategy)
	}

	router.Run(fmt.Sprintf(":%d", port))
}

// enable cors
func cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
		} else {
			c.Next()
		}
	}
}
