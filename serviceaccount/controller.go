package serviceaccount

import (
	"net/http"
	"strconv"

	"github.com/cloudogu/k8s-prometheus/auth/prometheus"
	"github.com/gin-gonic/gin"
)

type manager interface {
	// CreateOrUpdateServiceAccount creates a new - or updates an existing service account. The parameters [params] are
	// likely to be ignored because prometheus does not support any special parameters for service account handling. The
	// behavior parameters modify this manager's actions during service account interaction.
	CreateOrUpdateServiceAccount(consumer string, params map[string]string, behaviorParams prometheus.BehaviorParams) (credentials map[string]string, changed prometheus.CredentialChangeType, err error)
	DeleteServiceAccount(consumer string) error
	GetServiceAccount(consumer string) (user map[string]string, found bool, err error)
}

// Controller provides HTTP endpoint functionality for service account handling.
type Controller struct {
	manager manager
}

// NewController creates a new service account controller.
func NewController(manager manager) *Controller {
	return &Controller{manager: manager}
}

type createOrUpdateRequest struct {
	// Consumer contains the identifier of the Service Account Consumer. This field is required.
	Consumer string `json:"consumer"`
	// Params contains key/value parameters upon the producer modifies the service account creation/update. These
	// parameters are usually optional, anyhow developers of Service Account consumers are strongly asked to check the
	// producers requirements.
	Params map[string]string `json:"params"`
	// BehaviorParams contain information in which the Service Account producer may be triggered for an action. This field is strictly optional.
	BehaviorParams prometheus.BehaviorParams `json:"behaviorParams"`
}

// GetAccount provides an endpoint for checking the account for a given consumer.
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

// CreateOrUpdateAccount provides an idempotent endpoint of creating or updating service accounts.
func (ctrl *Controller) CreateOrUpdateAccount(c *gin.Context) {
	if c.Request.Method != http.MethodPut {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "wrong method"})
		return
	}

	var request createOrUpdateRequest
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if request.Consumer == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "consumer must not be empty"})
		return
	}

	credentials, changeType, err := ctrl.manager.CreateOrUpdateServiceAccount(request.Consumer, request.Params, request.BehaviorParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	switch changeType {
	case prometheus.CredentialCreated:
		c.JSON(http.StatusCreated, credentials)
	case prometheus.CredentialUpdated:
		c.JSON(http.StatusOK, credentials)
	case prometheus.CredentialNoChange:
		c.JSON(http.StatusNoContent, "")
	default:
		panic("unknown change type" + strconv.Itoa(int(changeType)))
	}
}

// DeleteAccount provides an endpoint for service account deletion.
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
