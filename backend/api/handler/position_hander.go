package handler

import (
	"minilend/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PositionHandler struct {
	positionService *service.PositionService
}

func NewPositionHandler() *PositionHandler {
	return &PositionHandler{
		positionService: &service.PositionService{},
	}
}

func (r *PositionHandler) RegisterRoutes(router *gin.RouterGroup) {
	positionGroup := router.Group("/positions")
	{
		positionGroup.GET("/:user_address", r.getUserPosition)
	}
}

func (r *PositionHandler) getUserPosition(c *gin.Context) {
	userAddress := c.Param("user_address")
	if userAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户地址不能为空"})
		return
	}

	position, err := r.positionService.GetUserPosition(userAddress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, position)
}
