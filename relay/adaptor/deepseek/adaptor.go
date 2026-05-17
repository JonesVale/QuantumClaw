package deepseek

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/quantumclaw/quantumclaw/relay/adaptor"
	"github.com/quantumclaw/quantumclaw/relay/meta"
	"github.com/quantumclaw/quantumclaw/relay/model"
	"github.com/quantumclaw/quantumclaw/relay/relaymode"
)

type Adaptor struct {
	ChannelType int
}

func (a *Adaptor) Init(meta *meta.Meta) {
	a.ChannelType = meta.ChannelType
}

func (a *Adaptor) GetRequestURL(meta *meta.Meta) (string, error) {
	switch meta.Mode {
	case relaymode.Completions:
		return fmt.Sprintf("%s/completions", meta.BaseURL), nil
	default:
		return fmt.Sprintf("%s/v1/chat/completions", meta.BaseURL), nil
	}
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

const ChannelName = "deepseek"

func init() {
	ModelList = append(ModelList,
		"deepseek-v4-flash", "deepseek-v4-flash-none", "deepseek-v4-flash-max",
		"deepseek-v4-pro", "deepseek-v4-pro-none", "deepseek-v4-pro-max",
	)
}