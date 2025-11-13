package objs

import (
	"moviepilot-go/internal/modules/themoviedb/tmdbv3api"
)

// Authentication 认证对象
type Authentication struct {
	*tmdbv3api.TMDb
	username      string
	password      string
	expiresAt     string
	requestToken  string
}

// NewAuthentication 创建Authentication实例
// 对应Python中的__init__方法
func NewAuthentication(username, password string) *Authentication {
	auth := &Authentication{
		TMDb:     tmdbv3api.NewTMDb(false, nil),
		username: username,
		password: password,
	}
	
	// 执行初始化流�?	auth.requestToken = auth.createRequestToken()
	auth.authoriseRequestTokenWithLogin()
	auth.createSession()
	
	return auth
}

// createRequestToken 创建请求令牌
// 对应Python中的_create_request_token方法
func (a *Authentication) createRequestToken() string {
	/*
	   Create a temporary request token that can be used to validate a TMDb user login.
	*/
	result, err := a.TMDb.RequestObj("/authentication/token/new", "", "GET", nil, nil, nil)
	if err != nil {
		panic(err)
	}
	
	if resultMap, ok := result.(map[string]interface{}); ok {
		if expiresAt, ok := resultMap["expires_at"].(string); ok {
			a.expiresAt = expiresAt
		}
		if requestToken, ok := resultMap["request_token"].(string); ok {
			return requestToken
		}
	}
	
	return ""
}

// createSession 创建会话
// 对应Python中的_create_session方法
func (a *Authentication) createSession() {
	/*
	   You can use this method to create a fully valid session ID once a user has validated the request token.
	*/
	jsonData := map[string]interface{}{
		"request_token": a.requestToken,
	}
	
	result, err := a.TMDb.RequestObj("/authentication/session/new", "", "POST", nil, jsonData, nil)
	if err != nil {
		panic(err)
	}
	
	if resultMap, ok := result.(map[string]interface{}); ok {
		if sessionID, ok := resultMap["session_id"].(string); ok {
			a.TMDb.SetSessionID(sessionID)
		}
	}
}

// authoriseRequestTokenWithLogin 使用登录信息授权请求令牌
// 对应Python中的_authorise_request_token_with_login方法
func (a *Authentication) authoriseRequestTokenWithLogin() {
	/*
	   This method allows an application to validate a request token by entering a username and password.
	*/
	jsonData := map[string]interface{}{
		"username":      a.username,
		"password":      a.password,
		"request_token": a.requestToken,
	}
	
	_, err := a.TMDb.RequestObj("/authentication/token/validate_with_login", "", "POST", nil, jsonData, nil)
	if err != nil {
		panic(err)
	}
}

// DeleteSession 删除会话
// 对应Python中的delete_session方法
func (a *Authentication) DeleteSession() {
	/*
	   If you would like to delete (or "logout") from a session, call this method with a valid session ID.
	*/
	if a.TMDb.HasSession() {
		jsonData := map[string]interface{}{
			"session_id": a.TMDb.SessionID(),
		}
		
		_, err := a.TMDb.RequestObj("/authentication/session", "", "DELETE", nil, jsonData, nil)
		if err != nil {
			panic(err)
		}
		
		a.TMDb.SetSessionID("")
	}
}
