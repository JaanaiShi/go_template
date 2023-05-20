package requests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"io"
	"net/http"
	"time"
)

func Get(url string, header map[string]string, lg *zap.Logger) (r []byte, err error) {
	defer func() {
		if e := recover(); e != nil {
			if lg != nil {
				lg.Error("HttpGetJson失败")
			}
		}
	}()

	var client = &http.Client{
		Timeout: 3 * time.Second,
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		if lg != nil {
			lg.Error("NewRequest error")
			return
		}
	}

	for k, v := range header {
		req.Header.Add(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		if lg != nil {
			lg.Error("构建get-http请求失败", zap.Error(err))
		}
		return
	}
	defer resp.Body.Close()
	r, err = io.ReadAll(resp.Body)
	if err != nil {
		if lg != nil {
			lg.Error(err.Error(), zap.Error(err))
		}
		return
	}

	if resp.StatusCode != http.StatusOK {
		e := fmt.Sprintf("%s：[%d] %s", url, resp.StatusCode, string(r))
		if lg != nil {
			lg.Error("获取错误状态码:"+e, zap.Int("response.StatusCode", resp.StatusCode))
		}
		return
	}

	return
}

func Post(url string, header map[string]string, data interface{}, lg *zap.Logger) (r []byte, err error) {
	buf := bytes.NewBuffer(nil)
	encoder := json.NewEncoder(buf)
	if err = encoder.Encode(data); err != nil {
		if lg != nil {
			lg.Error("解析data错误", zap.Error(err))
		}
		return nil, err
	}

	var timeout = 3 * time.Second

	var client = &http.Client{
		Timeout: timeout,
	}
	request, err := http.NewRequest(http.MethodPost, url, buf)
	if err != nil {
		if lg != nil {
			lg.Error("构建http请求失败", zap.Error(err))
		}
		return nil, err
	}

	request.Header.Add("Content-Type", "application/json")
	for k, v := range header {
		request.Header.Add(k, v)
	}

	response, err := client.Do(request)

	if response == nil {
		if err != nil {
			if lg != nil {
				lg.Error("请求失败："+err.Error(), zap.Any("response", response), zap.Error(err))
			}
			return
		}
		return
	}

	defer func() {
		_ = response.Body.Close()
	}()

	r, err = io.ReadAll(response.Body)
	if err != nil {
		if lg != nil {
			lg.Error("获取响应失败", zap.Error(err))
		}
		return
	}

	if response != nil && response.StatusCode != http.StatusOK {
		if lg != nil {
			lg.Error("http状态码异常", zap.Int("response.StatusCode", response.StatusCode))
		}
		err = fmt.Errorf("%s：[%d] %s", url, response.StatusCode, string(r))
		return
	}

	if err != nil {
		if lg != nil {
			lg.Error("请求失败："+err.Error(), zap.Any("response", response), zap.Error(err))
		}
		return
	}

	return
}
