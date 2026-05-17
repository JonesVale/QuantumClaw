package perplexity

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/quantumclaw/quantumclaw/relay/adaptor"
	"github.com/quantumclaw/quantumclaw/relay/meta"
	"github.com/quantumclaw/quantumclaw/relay/model"
)

type Adaptor struct {
	ChannelType int
}

var ModelList = []string{
	"llama-3-sonar-small-32k-chat", "llama-3-sonar-small-32k-online",
	"llama-3-sonar-large-32k-chat", "llama-3-sonar-large-32k-online",
	"llama-3-8b-instruct", "llama-3-70b-instruct", "mixtral-8x7b-instruct",
	"sonar", "sonar-pro", "sonar-reasoning",
}

const ChannelName = "perplexity"

func (a *Adaptor) Init(meta *meta.Meta) {
	a.ChannelType = meta.ChannelType
}

func (a *Adaptor) GetRequestURL(meta *meta.Meta) (string, error) {
	return fmt.Sprintf("%s/chat/completions", meta.BaseURL), nil
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Request, meta *meta.Meta) error {
	adaptor.SetupCommonRequestHeader(c, req, meta)
	req.Header.Set("Authorization", "Bearer "+meta.APIKey)
	return nil
}

func (a *Adaptor) ConvertRequest(c *gin.Context, relayMode int, request *model.GeneralOpenAIRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}
	if request.TopP != nil && *request.TopP >= 1 {
		*request.TopP = 0.99
	}
	return request, nil
}

func (a *Adaptor) ConvertImageRequest(request *model.ImageRequest) (any, error) {
	return nil, errors.New("not implemented")
}

func (a *Adaptor) DoRequest(c *gin.Context, meta *meta.Meta, requestBody io.Reader) (*http.Response, error) {
	return adaptor.DoRequestHelper(a, c, meta, requestBody)
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, meta *meta.Meta) (usage *model.Usage, err *model.ErrorWithStatusCode) {
	return nil, nil
}

func (a *Adaptor) GetModelList() []string {
	return ModelList
}

func (a *Adaptor) GetChannelName() string {
	return ChannelName
}