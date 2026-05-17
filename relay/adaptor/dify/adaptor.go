package dify

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/quantumclaw/quantumclaw/relay/adaptor"
	channelhelper "github.com/quantumclaw/quantumclaw/relay/adaptor"
	"github.com/quantumclaw/quantumclaw/relay/meta"
	"github.com/quantumclaw/quantumclaw/relay/model"
	relaymodel "github.com/quantumclaw/quantumclaw/relay/model"
)

var _ adaptor.Adaptor = new(Adaptor)

type Adaptor struct{}

func (a *Adaptor) Init(meta *meta.Meta) {
}

func (a *Adaptor) GetRequestURL(meta *meta.Meta) (string, error) {
	// Dify API uses its own RESTful endpoints, not OpenAI-style.
	// For passthrough, we forward the full request URL path to the base URL.
	requestURL := meta.BaseURL + meta.RequestURLPath
	return requestURL, nil
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Request, meta *meta.Meta) error {
	// Copy all headers from original request
	for k, v := range c.Request.Header {
		req.Header.Set(k, v[0])
	}

	// Remove headers that might interfere
	req.Header.Del("Host")
	req.Header.Del("Content-Length")
	req.Header.Del("Accept-Encoding")
	req.Header.Del("Connection")

	// Dify uses "Authorization: Bearer <api-key>" or plain api-key via header.
	// The meta.APIKey already carries the user's API key.
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", meta.APIKey))

	return nil
}

func (a *Adaptor) ConvertRequest(c *gin.Context, relayMode int, request *model.GeneralOpenAIRequest) (any, error) {
	// For passthrough mode, we don't convert the request.
	// Dify ChatFlow API has a different format than OpenAI.
	// We pass the request body through unchanged.
	return nil, errors.New("dify adaptor does not support request conversion; use passthrough mode")
}

func (a *Adaptor) ConvertImageRequest(request *model.ImageRequest) (any, error) {
	return nil, errors.Errorf("dify adaptor does not support image request conversion")
}

func (a *Adaptor) DoRequest(c *gin.Context, meta *meta.Meta, requestBody io.Reader) (*http.Response, error) {
	return channelhelper.DoRequestHelper(a, c, meta, requestBody)
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, meta *meta.Meta) (usage *model.Usage, err *model.ErrorWithStatusCode) {
	// Passthrough: copy all response headers and body
	for k, v := range resp.Header {
		for _, vv := range v {
			c.Writer.Header().Set(k, vv)
		}
	}

	c.Writer.WriteHeader(resp.StatusCode)
	if _, gerr := io.Copy(c.Writer, resp.Body); gerr != nil {
		return nil, &relaymodel.ErrorWithStatusCode{
			StatusCode: http.StatusInternalServerError,
			Error: relaymodel.Error{
				Message: gerr.Error(),
			},
		}
	}

	return nil, nil
}

func (a *Adaptor) GetModelList() []string {
	return ModelList
}

func (a *Adaptor) GetChannelName() string {
	return ChannelName
}
