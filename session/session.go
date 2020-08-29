package session

import (
	Logger "github.com/edukacloud-api/go-pkg/logger"
	Map "github.com/orcaman/concurrent-map"

	"fmt"
	"github.com/pkg/errors"
	"github.com/spf13/cast"
	"go.uber.org/zap"
	"time"
)

type Session struct {
	Map                     Map.ConcurrentMap
	Logger                  Logger.Logger
	RequestTime             time.Time
	ThreadID                string
	AppName, AppVersion, IP string
	Port                    int
	SrcIP, URL, Method      string
	Header, Request         interface{}
	ErrorMessage            string
}

func New(logger Logger.Logger) *Session {
	return &Session{
		RequestTime: time.Now(),
		Logger:      logger,
		Map:         Map.New(),
		Header:      map[string]interface{}{},
		Request:     struct{}{},
	}
}

func (session *Session) SetThreadID(threadID string) *Session {
	session.ThreadID = threadID
	return session
}

func (session *Session) SetMethod(method string) *Session {
	session.Method = method
	return session
}

func (session *Session) SetAppName(appName string) *Session {
	session.AppName = appName
	return session
}

func (session *Session) SetAppVersion(appVersion string) *Session {
	session.AppVersion = appVersion
	return session
}

func (session *Session) SetURL(url string) *Session {
	session.URL = url
	return session
}

func (session *Session) SetIP(ip string) *Session {
	session.IP = ip
	return session
}

func (session *Session) SetPort(port int) *Session {
	session.Port = port
	return session
}

func (session *Session) SetSrcIP(srcIp string) *Session {
	session.SrcIP = srcIp
	return session
}

func (session *Session) SetHeader(header interface{}) *Session {
	session.Header = header
	return session
}

func (session *Session) SetRequest(request interface{}) *Session {
	session.Request = request
	return session
}

func (session *Session) SetErrorMessage(errorMessage string) *Session {
	session.ErrorMessage = errorMessage
	return session
}

func (session *Session) Get(key string) (data interface{}, err error) {
	data, ok := session.Map.Get(key)
	if !ok {
		err = errors.New("not found")
	}
	return
}

func (session *Session) Put(key string, data interface{}) {
	session.Map.Set(key, data)
}

func (session *Session) T1(message ...interface{}) {
	logRecord := []zap.Field{
		zap.String("_app_tag", "T1"),
		zap.String("_app_thread_id", session.ThreadID),
		zap.String("_app_method", session.Method),
		zap.String("_app_uri", session.URL),
	}

	msg := formatLogs(message...)
	logRecord = append(logRecord, msg...)

	session.Logger.Info("|", logRecord...)
}

func (session *Session) T2(message ...interface{}) time.Time {
	logRecord := []zap.Field{
		zap.String("_app_tag", "T2"),
		zap.String("_app_thread_id", session.ThreadID),
		zap.String("_app_method", session.Method),
		zap.String("_app_uri", session.URL),
	}

	msg := formatLogs(message...)
	logRecord = append(logRecord, msg...)

	session.Logger.Info("|", logRecord...,
	)

	return time.Now()
}

func (session *Session) T3(startProcessTime time.Time, message ...interface{}) {
	stop := time.Now()

	logRecord := []zap.Field{
		zap.String("_app_tag", "T3"),
		zap.String("_app_thread_id", session.ThreadID),
		zap.String("_app_method", session.Method),
		zap.String("_app_uri", session.URL),
	}

	msg := formatLogs(message...)
	logRecord = append(logRecord, msg...)
	logRecord = append(logRecord, zap.String("_process_time", fmt.Sprintf("%d ms", stop.Sub(startProcessTime).Nanoseconds()/1000000)))

	session.Logger.Info("|", logRecord...)
}

func (session *Session) T4(message ...interface{}) {
	stop := time.Now()
	rt := stop.Sub(session.RequestTime).Nanoseconds() / 1000000

	logRecord := []zap.Field{
		zap.String("_app_tag", "T4"),
		zap.String("_app_thread_id", session.ThreadID),
		zap.String("_app_method", session.Method),
		zap.String("_app_uri", session.URL),
	}

	msg := formatLogs(message...)
	logRecord = append(logRecord, msg...)
	logRecord = append(logRecord, zap.String("_response_time", fmt.Sprintf("%d ms", rt)))

	session.Logger.Info("|", logRecord...)

	session.Logger.TDR(Logger.LogTdrModel{
		AppName:        session.AppName,
		AppVersion:     session.AppVersion,
		IP:             session.IP,
		Port:           session.Port,
		SrcIP:          session.SrcIP,
		RespTime:       rt,
		Path:           session.URL,
		Header:         session.Header,
		Request:        session.Request,
		Response:       message,
		Error:          session.ErrorMessage,
		ThreadID:       session.ThreadID,
		AdditionalData: session.Map,
	})
}

func (session *Session) Info(message ...interface{}) {
	logRecord := []zap.Field{
		zap.String("_app_tag", "INFO"),
		zap.String("_app_thread_id", session.ThreadID),
		zap.String("_app_method", session.Method),
		zap.String("_app_uri", session.URL),
	}

	msg := formatLogs(message...)
	logRecord = append(logRecord, msg...)

	session.Logger.Info("|", logRecord...)
}

func (session *Session) Error(message ...interface{}) {
	logRecord := []zap.Field{
		zap.String("_app_tag", "ERROR"),
		zap.String("_app_thread_id", session.ThreadID),
		zap.String("_app_method", session.Method),
		zap.String("_app_uri", session.URL),
	}

	msg := formatLogs(message...)
	logRecord = append(logRecord, msg...)

	session.Logger.Error("|", logRecord...)
}

func formatLogs(message ...interface{}) (logRecord []zap.Field) {
	for index, msg := range message {
		logRecord = append(logRecord, Logger.FormatLog("_message_"+cast.ToString(index), msg))
	}

	return
}
