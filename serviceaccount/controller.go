package serviceaccount

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type manager interface {
	CreateServiceAccount(consumer string, params []string) (credentials map[string]string, err error)
	DeleteServiceAccount(consumer string) error
	GetServiceAccount(consumer string) (user map[string]string, found bool, err error)
}

type Controller struct {
	manager manager
}

func NewController(manager manager) *Controller {
	return &Controller{manager: manager}
}

type createRequest struct {
	Consumer string   `json:"consumer"`
	Params   []string `json:"params"`
}

func (ctrl *Controller) GetAccount(c *gin.Context) {
	consumer := c.Param("consumer")
	if consumer == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "consumer cannot be empty"})
		return
	}

	user, found, err := ctrl.manager.GetServiceAccount(consumer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	switch c.Request.Method {
	case http.MethodGet:
		c.JSON(http.StatusOK, user)
	case http.MethodHead:
		c.Status(http.StatusOK)
	default:
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "wrong method"})
	}
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

	c.JSON(http.StatusCreated, credentials)
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
