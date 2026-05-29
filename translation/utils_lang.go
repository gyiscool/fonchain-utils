package translation

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sort"
	"strings"
	"time" // 引入超时依赖
)

// 补充Response相关结构体定义（原代码缺失，确保编译通过）
type TranslateData struct {
	Content string `json:"content"` // 翻译结果内容
}

type Response struct {
	Status  int            `json:"status"`  // 响应状态码（HTTP状态码）
	Message string         `json:"message"` // 响应描述信息
	Code    int            `json:"code"`    // 业务错误码（0表示成功）
	Data    *TranslateData `json:"data"`    // 翻译结果数据（可能为nil）
}

type translation struct {
	ClientID string
	Secret   string
	Url      string
}

var translationObj *translation

func NewTranslation(ak, sk, url string) {
	if translationObj == nil {
		translationObj = &translation{
			ClientID: ak,
			Secret:   sk,
			Url:      url,
		}
	}
}

/*
 * GetConvertString 函数用于调用翻译API将消息内容翻译成指定语言
 * 参数:
 *   msg: 需要翻译的消息内容
 *   lang: 目标语言，支持 "Ja"(日语)、"ZhTW"(繁体中文)、"EN"(英语)、"Check"(仅校验)、"ZhCN"(简体中文，目前是仅仅校验)
 *   sys: 系统标识
 * 返回值:
 *   string: 翻译后的消息内容，如果翻译失败则返回原始消息  正常仅仅接受中文
 *   error: 错误信息，翻译成功时为nil
 * 特别注意 如果lang 不是以上五种字符串，会原样返回
 * 我建议不要考虑error返回，直接获取 string
 */
func GetConvertString(msg, lang, sys string) (string, error) {
	// 1. 入参简单校验（避免空请求浪费资源）
	if translationObj == nil {
		return "", errors.New("没有初始化")
	}
	var requestUrl = translationObj.Url

	// 检查消息内容是否为空
	if msg == "" {
		return "", nil
	}

	// 检查语言参数是否为空
	if lang == "" {
		return msg, nil
	}

	// 检查语言参数是否为支持的值
	if lang != "Ja" && lang != "ZhTW" && lang != "EN" && lang != "Check" && lang != "ZhCN" {
		return msg, nil
	}

	// 2. 定义请求参数

	// 构建请求数据
	data := map[string]string{
		"lang": lang,
		"msg":  msg,
		"sys":  sys,
	}

	fmt.Println("1---------", data)

	// 生成时间戳和随机数
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	nonce := generateNonce(16)

	// 生成签名: SHA1(sorted([secret, timestamp, nonce, client]))
	sign := generateSign(translationObj.Secret, timestamp, nonce, translationObj.ClientID)

	// 构建请求URL
	realuUrl := fmt.Sprintf("%s?timestamp=%s&nonce=%s&client=%s&sign=%s",
		requestUrl, timestamp, nonce, translationObj.ClientID, sign)
	// 3. JSON编码（错误直接返回，日志记录详细信息）
	jsonData, err := json.Marshal(data)
	if err != nil {
		return msg, fmt.Errorf("JSON编码失败: %w", err)
	}

	// 4. 创建HTTP请求（POST方法）
	req, err := http.NewRequest("POST", realuUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		return msg, fmt.Errorf("创建请求失败: %w", err)
	}

	// 5. 设置请求头（指定JSON格式）
	req.Header.Set("Content-Type", "application/json; charset=utf-8") // 补充charset避免编码问题

	// 6. 创建HTTP客户端（核心：设置3秒超时）
	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	// 7. 发送请求（超时会返回 *url.Error，错误信息含"context deadline exceeded"）
	resp, err := client.Do(req)
	if err != nil {
		// 超时错误单独标注（方便调用方判断）
		if errors.Is(err, context.DeadlineExceeded) {
			return msg, errors.New("翻译请求超时，请稍后重试")
		}
		return msg, fmt.Errorf("发送请求失败: %w", err)
	}
	// 关键：确保resp非nil时关闭Body（避免资源泄露）
	if resp != nil {
		defer resp.Body.Close()
	}

	// 8. 校验HTTP状态码（非200视为请求失败）
	if resp.StatusCode != http.StatusOK {
		// 读取错误响应体（辅助排查问题）
		//errBody, _ := io.ReadAll(resp.Body)
		return msg, fmt.Errorf("服务器异常，状态码: %d", resp.StatusCode)
	}

	// 9. 读取响应体（替换废弃的ioutil.ReadAll）
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return msg, fmt.Errorf("读取响应体失败: %w", err)
	}

	// 10. 解析响应JSON
	var response Response
	if err := json.Unmarshal(body, &response); err != nil {
		return msg, fmt.Errorf("解析翻译结果失败: %w", err)
	}

	// 12. 业务逻辑校验（code≠0视为失败）
	if response.Code != 0 {
		return msg, fmt.Errorf("翻译失败: %s（错误码：%d）", response.Message, response.Code)
	}

	// 13. 校验数据非空
	if response.Data == nil || response.Data.Content == "" {
		return msg, errors.New("翻译结果为空")
	}

	// 成功返回翻译结果
	return response.Data.Content, nil
}

func generateSign(token string, timestamp string, nonce string, client string) (ret string) {
	strs := []string{token, timestamp, nonce, client}

	sort.Strings(strs)
	vv := strings.Join(strs, "")
	t := sha1.New()
	t.Write([]byte(vv))

	return fmt.Sprintf("%x", t.Sum(nil))
}

// generateNonce 生成随机nonce
func generateNonce(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
