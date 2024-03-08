package serviceaccount

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type Manager interface {
	CreateServiceAccount(consumer string, params []string) (credentials map[string]string, err error)
	DeleteServiceAccount(consumer string) error
}

type Controller struct {
	manager Manager
}

func NewController(manager Manager) *Controller {
	return &Controller{manager: manager}
}

type createRequest struct {
	Consumer string   `json:"consumer"`
	Params   []string `json:"params"`
}

func (ctrl *Controller) CreateAccount(c *gin.Context) {
	var request createRequest
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if request.Consumer == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "consumer must not be empty"})
		return
	}

	credentials, err := ctrl.manager.CreateServiceAccount(request.Consumer, request.Params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, credentials)
}

func (ctrl *Controller) DeleteAccount(c *gin.Context) {
	consumer := c.Param("consumer")
	if consumer == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "consumer cannot be empty"})
		return
	}

	if err := ctrl.manager.DeleteServiceAccount(consumer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
