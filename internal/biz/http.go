package biz

import (
	"encoding/json"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
)

// 未完成 - 建议使用 chromeDp 完成，简单的可以使用curl command完成

type HttpInspectClientUseCase struct {
	logger *zap.Logger
}

func NewHttpInspectClientUseCase(logger *zap.Logger) *HttpInspectClientUseCase {
	return &HttpInspectClientUseCase{
		logger: logger,
	}
}

type HttpInspectReply struct {
	Url        string `json:"url,omitempty"`
	StatusCode int    `json:"statusCode,omitempty"`
	Response   []byte `json:"response,omitempty"`
}

func (httpInspectReply *HttpInspectReply) String() string {
	b, err := json.Marshal(httpInspectReply)
	if err != nil {
		return err.Error()
	}
	
	return string(b)
}

func (httpInspectClientUseCase *HttpInspectClientUseCase) GetHttpUrlResponse(url string) (HttpInspectReply, error) {
	var httpInspectReply HttpInspectReply
	resp, err := http.Get(url)
	if err != nil {
		return httpInspectReply, err
	}
	
	respByte, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return httpInspectReply, err
	}
	
	httpInspectReply.Response = respByte
	defer resp.Body.Close()
	
	httpInspectReply.StatusCode = resp.StatusCode
	httpInspectReply.Url = url
	return httpInspectReply, nil
}
