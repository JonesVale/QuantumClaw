package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/quantumclaw/quantumclaw/common/config"
	"github.com/quantumclaw/quantumclaw/common/ctxkey"
	"github.com/quantumclaw/quantumclaw/common/helper"
	"github.com/quantumclaw/quantumclaw/common/logger"
	"github.com/quantumclaw/quantumclaw/common/message"
	"github.com/quantumclaw/quantumclaw/middleware"
	"github.com/quantumclaw/quantumclaw/model"
	"github.com/quantumclaw/quantumclaw/monitor"
	"github.com/quantumclaw/quantumclaw/relay"
	"github.com/quantumclaw/quantumclaw/relay/adaptor/openai"
	"github.com/quantumclaw/quantumclaw/relay/channeltype"
	"github.com/quantumclaw/quantumclaw/relay/quantum/azure"
	"github.com/quantumclaw/quantumclaw/relay/quantum/braket"
	"github.com/quantumclaw/quantumclaw/relay/quantum/ibmq"
	"github.com/quantumclaw/quantumclaw/relay/quantum/ionq"
	"github.com/quantumclaw/quantumclaw/relay/quantum/rigetti"
	"github.com/quantumclaw/quantumclaw/relay/controller"
	"github.com/quantumclaw/quantumclaw/relay/meta"
	relaymodel "github.com/quantumclaw/quantumclaw/relay/model"
	"github.com/quantumclaw/quantumclaw/relay/relaymode"
)

func buildTestRequest(model string) *relaymodel.GeneralOpenAIRequest {
	if model == "" {
		// 默认使用 auto 模式, 优先免费/低成本模型
		model = "gpt-3.5-turbo" // 回落默认值
		// 环境变量可覆盖: TEST_MODEL=auto 或 TEST_MODEL=deepseek-chat
		if envModel := os.Getenv("TEST_MODEL"); envModel != "" {
			model = envModel
		}
	}
	testRequest := &relaymodel.GeneralOpenAIRequest{
		Model: model,
	}
	testMessage := relaymodel.Message{
		Role:    "user",
		Content: config.TestPrompt,
	}
	testRequest.Messages = append(testRequest.Messages, testMessage)
	return testRequest
}

func parseTestResponse(resp string) (*openai.TextResponse, string, error) {
	var response openai.TextResponse
	err := json.Unmarshal([]byte(resp), &response)
	if err != nil {
		return nil, "", err
	}
	if len(response.Choices) == 0 {
		return nil, "", errors.New("response has no choices")
	}
	stringContent, ok := response.Choices[0].Content.(string)
	if !ok {
		return nil, "", errors.New("response content is not string")
	}
	return &response, stringContent, nil
}

func testChannel(ctx context.Context, channel *model.Channel, request *relaymodel.GeneralOpenAIRequest) (responseMessage string, err error, openaiErr *relaymodel.Error) {
	startTime := time.Now()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = &http.Request{
		Method: "POST",
		URL:    &url.URL{Path: "/v1/chat/completions"},
		Body:   nil,
		Header: make(http.Header),
	}
	c.Request = c.Request.WithContext(ctx)
	c.Request.Header.Set("Authorization", "Bearer "+channel.Key)
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set(ctxkey.Channel, channel.Type)
	c.Set(ctxkey.BaseURL, channel.GetBaseURL())
	cfg, _ := channel.LoadConfig()
	c.Set(ctxkey.Config, cfg)
	middleware.SetupContextForSelectedChannel(c, channel, "")
	meta := meta.GetByContext(c)
	apiType := channeltype.ToAPIType(channel.Type)
	adaptor := relay.GetAdaptor(apiType)
	if adaptor == nil {
		return "", fmt.Errorf("invalid api type: %d, adaptor is nil", apiType), nil
	}
	adaptor.Init(meta)
	modelName := request.Model
	modelMap := channel.GetModelMapping()
	if modelName == "" || !strings.Contains(channel.Models, modelName) {
		modelNames := strings.Split(channel.Models, ",")
		if len(modelNames) > 0 && modelNames[0] != "" {
			modelName = modelNames[0]
		} else {
			// 渠道没有配置模型，使用 buildTestRequest 传入的默认模型
			if modelName == "" {
				modelName = "gpt-3.5-turbo"
			}
		}
	}
	if modelMap != nil && modelMap[modelName] != "" {
		modelName = modelMap[modelName]
	}
	meta.OriginModelName, meta.ActualModelName = request.Model, modelName
	request.Model = modelName
	convertedRequest, err := adaptor.ConvertRequest(c, relaymode.ChatCompletions, request)
	if err != nil {
		return "", err, nil
	}
	jsonData, err := json.Marshal(convertedRequest)
	if err != nil {
		return "", err, nil
	}
	defer func() {
		logContent := fmt.Sprintf("渠道 %s 测试成功，响应：%s", channel.Name, responseMessage)
		if err != nil || openaiErr != nil {
			errorMessage := ""
			if err != nil {
				errorMessage = err.Error()
			} else {
				errorMessage = openaiErr.Message
			}
			logContent = fmt.Sprintf("渠道 %s 测试失败，错误：%s", channel.Name, errorMessage)
		}
		go model.RecordTestLog(ctx, &model.Log{
			ChannelId:   channel.Id,
			ModelName:   modelName,
			Content:     logContent,
			ElapsedTime: helper.CalcElapsedTime(startTime),
		})
	}()
	logger.SysLog(string(jsonData))
	requestBody := bytes.NewBuffer(jsonData)
	c.Request.Body = io.NopCloser(requestBody)
	resp, err := adaptor.DoRequest(c, meta, requestBody)
	if err != nil {
		return "", err, nil
	}
	if resp != nil && resp.StatusCode != http.StatusOK {
		err := controller.RelayErrorHandler(resp)
		errorMessage := err.Error.Message
		if errorMessage != "" {
			errorMessage = ", error message: " + errorMessage
		}
		return "", fmt.Errorf("http status code: %d%s", resp.StatusCode, errorMessage), &err.Error
	}
	usage, respErr := adaptor.DoResponse(c, resp, meta)
	if respErr != nil {
		return "", fmt.Errorf("%s", respErr.Error.Message), &respErr.Error
	}
	if usage == nil {
		return "", errors.New("usage is nil"), nil
	}
	rawResponse := w.Body.String()
	_, responseMessage, err = parseTestResponse(rawResponse)
	if err != nil {
		logger.SysError(fmt.Sprintf("failed to parse error: %s, \nresponse: %s", err.Error(), rawResponse))
		return "", err, nil
	}
	result := w.Result()
	// print result.Body
	respBody, err := io.ReadAll(result.Body)
	if err != nil {
		return "", err, nil
	}
	logger.SysLog(fmt.Sprintf("testing channel #%d, response: \n%s", channel.Id, string(respBody)))
	return responseMessage, nil, nil
}

func TestChannel(c *gin.Context) {
	ctx := c.Request.Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	channel, err := model.GetChannelById(id, true)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	// 权限校验: 非管理员只能测试自己的渠道
	userId := c.GetInt("id")
	role := c.GetInt("role")
	if role < model.RoleAdminUser && channel.UserId != userId {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "仅可测试自己的渠道",
		})
		return
	}
	// 量子渠道使用独立的测试逻辑
	if channel.Type >= 100 {
		tik := time.Now()
		responseMessage, err := testQuantumChannel(ctx, channel)
		tok := time.Now()
		milliseconds := tok.Sub(tik).Milliseconds()
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error(), "time": float64(milliseconds) / 1000.0})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "message": responseMessage, "time": float64(milliseconds) / 1000.0})
		return
	}

	modelName := c.Query("model")
	testRequest := buildTestRequest(modelName)

	// 测试连接设置 30 秒超时上下文，防止挂起
	testCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	tik := time.Now()
	responseMessage, err, _ := testChannel(testCtx, channel, testRequest)
	tok := time.Now()
	milliseconds := tok.Sub(tik).Milliseconds()
	if err != nil {
		milliseconds = 0
	}
	go channel.UpdateResponseTime(milliseconds)
	consumedTime := float64(milliseconds) / 1000.0
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error(), "time": consumedTime, "modelName": modelName})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": responseMessage, "time": consumedTime, "modelName": modelName})
	return
}

var testAllChannelsLock sync.Mutex
var testAllChannelsRunning bool = false

func testChannels(ctx context.Context, notify bool, scope string) error {
	if config.RootUserEmail == "" {
		config.RootUserEmail = model.GetRootUserEmail()
	}
	testAllChannelsLock.Lock()
	if testAllChannelsRunning {
		testAllChannelsLock.Unlock()
		return errors.New("测试已在运行中")
	}
	testAllChannelsRunning = true
	testAllChannelsLock.Unlock()
	channels, err := model.GetAllChannels(0, 0, scope)
	if err != nil {
		return err
	}
	var disableThreshold = int64(config.ChannelDisableThreshold * 1000)
	if disableThreshold == 0 {
		disableThreshold = 10000000 // a impossible value
	}
	go func() {
		for _, channel := range channels {
			isChannelEnabled := channel.Status == model.ChannelStatusEnabled

			// 量子渠道：用量子测试逻辑
			if channel.Type >= 100 {
				_, testErr := testQuantumChannel(ctx, channel)
				if testErr != nil && isChannelEnabled {
					monitor.DisableChannel(channel.Id, channel.Name, testErr.Error())
				}
				continue
			}

			tik := time.Now()
			testRequest := buildTestRequest("")
			_, err, openaiErr := testChannel(ctx, channel, testRequest)
			tok := time.Now()
			milliseconds := tok.Sub(tik).Milliseconds()
			if isChannelEnabled && milliseconds > disableThreshold {
				err = fmt.Errorf("响应时间 %.2fs 超过阈值 %.2fs", float64(milliseconds)/1000.0, float64(disableThreshold)/1000.0)
				if config.AutomaticDisableChannelEnabled {
					monitor.DisableChannel(channel.Id, channel.Name, err.Error())
				} else {
					_ = message.Notify(message.ByAll, fmt.Sprintf("渠道 %s （%d）测试超时", channel.Name, channel.Id), "", err.Error())
				}
			}
			if isChannelEnabled && monitor.ShouldDisableChannel(openaiErr, -1) {
				monitor.DisableChannel(channel.Id, channel.Name, err.Error())
			}
			if !isChannelEnabled && monitor.ShouldEnableChannel(err, openaiErr) {
				monitor.EnableChannel(channel.Id, channel.Name)
			}
			// 保存测试结果
			testPassed := err == nil && (openaiErr == nil)
			errMsg := ""
			if err != nil {
				errMsg = err.Error()
			} else if openaiErr != nil {
				errMsg = openaiErr.Message
			}
			model.DB.Model(&model.Channel{}).Where("id = ?", channel.Id).Updates(map[string]interface{}{
				"last_test_passed":    testPassed,
				"last_error_message":  errMsg,
			})
			channel.UpdateResponseTime(milliseconds)
			time.Sleep(config.RequestInterval)
		}
		testAllChannelsLock.Lock()
		testAllChannelsRunning = false
		testAllChannelsLock.Unlock()
		if notify {
			err := message.Notify(message.ByAll, "渠道测试完成", "", "渠道测试完成，如果没有收到禁用通知，说明所有渠道都正常")
			if err != nil {
				logger.SysError(fmt.Sprintf("failed to send email: %s", err.Error()))
			}
		}
	}()
	return nil
}

func TestChannels(c *gin.Context) {
	ctx := c.Request.Context()
	scope := c.Query("scope")
	if scope == "" {
		scope = "all"
	}
	err := testChannels(ctx, true, scope)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
	return
}

func AutomaticallyTestChannels(frequency int) {
	ctx := context.Background()
	for {
		time.Sleep(time.Duration(frequency) * time.Minute)
		logger.SysLog("testing all channels")
		_ = testChannels(ctx, false, "all")
		logger.SysLog("channel test finished")
	}
}

// testQuantumChannel 测试量子渠道连通性：列举可用量子后端
func testQuantumChannel(ctx context.Context, channel *model.Channel) (string, error) {
	qAdaptor, err := relay.GetQuantumAdaptor(channel.Type)
	if err != nil {
		return "", fmt.Errorf("quantum adaptor error: %w", err)
	}

	// 注入 API Key 和 BaseURL
	switch a := qAdaptor.(type) {
	case *ionq.Adaptor:
		a.APIKey = channel.Key
		baseURL := channel.GetBaseURL()
		if baseURL == "" {
			baseURL = "https://api.ionq.co"
		}
		a.BaseURL = baseURL
	case *ibmq.Adaptor:
		a.APIKey = channel.Key
		baseURL := channel.GetBaseURL()
		if baseURL == "" {
			baseURL = "https://api.quantum.ibm.com"
		}
		a.BaseURL = baseURL
	case *rigetti.Adaptor:
		a.APIKey = channel.Key
		baseURL := channel.GetBaseURL()
		if baseURL == "" {
			baseURL = "https://api.qcs.rigetti.com"
		}
		a.BaseURL = baseURL
	case *braket.Adaptor:
		a.APIKey = channel.Key
		baseURL := channel.GetBaseURL()
		if baseURL == "" {
			baseURL = "https://braket.us-west-1.amazonaws.com"
		}
		a.BaseURL = baseURL
	case *azure.Adaptor:
		a.APIKey = channel.Key
		baseURL := channel.GetBaseURL()
		if baseURL == "" {
			baseURL = "https://quantum.azure.com/api"
		}
		a.BaseURL = baseURL
	}

	// 列举后端以验证连通性
	backends, err := qAdaptor.ListBackends(ctx)
	if err != nil {
		return "", fmt.Errorf("quantum test failed: %w", err)
	}
	return fmt.Sprintf("connected, backends: %s", strings.Join(backends, ", ")), nil
}
