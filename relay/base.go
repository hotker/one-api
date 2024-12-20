package relay

import (
	"one-api/types"

	providersBase "one-api/providers/base"

	"github.com/gin-gonic/gin"
)

type relayBase struct {
	c             *gin.Context
	provider      providersBase.ProviderInterface
	originalModel string
	modelName     string
}

type RelayBaseInterface interface {
	send() (err *types.OpenAIErrorWithStatusCode, done bool)
	getPromptTokens() (int, error)
	setRequest() error
	getRequest() any
	setProvider(modelName string) error
	getProvider() providersBase.ProviderInterface
	getOriginalModel() string
	getModelName() string
	getContext() *gin.Context
	IsStream() bool
}

func (r *relayBase) getRequest() interface{} {
	return nil
}

func (r *relayBase) IsStream() bool {
	return false
}

func (r *relayBase) setProvider(modelName string) error {
	provider, modelName, fail := GetProvider(r.c, modelName)
	if fail != nil {
		return fail
	}
	r.provider = provider
	r.modelName = modelName
	return nil
}

func (r *relayBase) getContext() *gin.Context {
	return r.c
}

func (r *relayBase) getProvider() providersBase.ProviderInterface {
	return r.provider
}

func (r *relayBase) getOriginalModel() string {
	return r.originalModel
}

func (r *relayBase) getModelName() string {
	return r.modelName
}
