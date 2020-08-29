package logger

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"runtime"
	"time"

	grpc_logging "github.com/grpc-ecosystem/go-grpc-middleware/logging"
	jsoniter "github.com/json-iterator/go"
	rotateLogs "github.com/lestrrat-go/file-rotatelogs"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Fields map[string]interface{}

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type Logger interface {
	Debug(message string, fields ...zap.Field)
	Info(message string, fields ...zap.Field)
	Warn(message string, fields ...zap.Field)
	Error(message string, fields ...zap.Field)
	Fatal(message string, fields ...zap.Field)
	Panic(message string, fields ...zap.Field)
	TDR(tdr LogTdrModel)
}

func New(config Options) Logger {
	cores := []zapcore.Core{}

	var writer zapcore.WriteSyncer

	if config.Stdout {
		writer = zapcore.AddSync(os.Stdout)
	} else {
		rotate, err := rotateLogs.New(
			config.FileLocation+".%Y%m%d",
			rotateLogs.WithLinkName(config.FileLocation),
			rotateLogs.WithMaxAge(config.FileMaxAge*24*time.Hour),
			rotateLogs.WithRotationTime(time.Hour),
		)
		if err != nil {
			panic(err)
		}
		writer = zapcore.AddSync(rotate)
	}

	core := zapcore.NewCore(getEncoder(), writer, zapcore.InfoLevel)
	cores = append(cores, core)

	combinedCore := zapcore.NewTee(cores...)

	logger := zap.New(combinedCore,
		zap.AddCallerSkip(3),
		zap.AddCaller(),
	)

	//logger TDR
	var tdrWriter zapcore.WriteSyncer
	if config.Stdout {
		tdrWriter = zapcore.AddSync(os.Stdout)
	} else {
		rotateLogsTdr, err := rotateLogs.New(
			config.FileTdrLocation+".%Y%m%d",
			rotateLogs.WithLinkName(config.FileTdrLocation),
			rotateLogs.WithMaxAge(config.FileMaxAge*24*time.Hour),
			rotateLogs.WithRotationTime(time.Hour),
		)
		if err != nil {
			panic(err)
		}
		tdrWriter = zapcore.AddSync(rotateLogsTdr)
	}

	tdrCore := zapcore.NewCore(getTdrEncoder(), tdrWriter, zapcore.InfoLevel)
	loggerTdr := zap.New(tdrCore,
		zap.AddCallerSkip(2),
		zap.AddCaller(),
	)

	return &zapLogger{
		logger:    logger,
		loggerTdr: loggerTdr,
	}
}

type zapLogger struct {
	logger    *zap.Logger
	loggerTdr *zap.Logger
}

type LogTdrModel struct {
	AppName        string      `json:"app"`
	AppVersion     string      `json:"ver"`
	IP             string      `json:"ip"`
	Port           int         `json:"port"`
	SrcIP          string      `json:"srcIP"`
	RespTime       int64       `json:"rt"`
	Path           string      `json:"path"`
	Header         interface{} `json:"header"` // better to pass data here as is, don't cast it to string. use map or array
	Request        interface{} `json:"req"`
	Response       interface{} `json:"resp"`
	Error          string      `json:"error"`
	ThreadID       string      `json:"threadID"`
	AdditionalData interface{} `json:"addData"`
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.TimeKey = "xtime"
	encoderConfig.MessageKey = "xdata"
	return zapcore.NewJSONEncoder(encoderConfig)
}

func getTdrEncoder() zapcore.Encoder {
	tdrConfig := zapcore.EncoderConfig{
		TimeKey:        "xtime",
		MessageKey:     "xdata",
		EncodeDuration: MillisDurationEncoder,
		EncodeTime:     TDRLogTimeEncoder,
		LineEnding:     zapcore.DefaultLineEnding,
	}
	return zapcore.NewJSONEncoder(tdrConfig)
}

func TDRLogTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.999"))
}

func MillisDurationEncoder(d time.Duration, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendInt64(d.Nanoseconds() / 1000000)
}

func (l *zapLogger) Debug(message string, fields ...zap.Field) {
	l.logger.Debug(message, fields...)
}

func (l *zapLogger) Info(message string, fields ...zap.Field) {
	l.logger.Info(message, fields...)
}

func (l *zapLogger) Warn(message string, fields ...zap.Field) {
	l.logger.Warn(message, fields...)
}

func (l *zapLogger) Error(message string, fields ...zap.Field) {
	l.logger.Error(message, fields...)
}

func (l *zapLogger) Fatal(message string, fields ...zap.Field) {
	l.logger.Fatal(message, fields...)
}

func (l *zapLogger) Panic(message string, fields ...zap.Field) {
	l.logger.Panic(message, fields...)
}

func (l *zapLogger) TDR(model LogTdrModel) {
	l.loggerTdr.Info(
		"|",
		zap.String("xid", model.ThreadID),
		zap.Int64("rt", model.RespTime),
		zap.Int("port", model.Port),
		zap.String("ip", model.IP),
		zap.String("app", model.AppName),
		zap.String("ver", model.AppVersion),
		zap.String("path", model.Path),
		zap.Any("header", model.Header),
		FormatLog("req", model.Request),
		FormatLog("resp", model.Response),
		zap.String("srcIP", model.SrcIP),
		zap.String("error", model.Error),
		FormatLog("addData", model.AdditionalData),
	)
}

func toJSON(obj interface{}) interface{} {
	if obj == nil {
		return nil
	}
	if w, ok := obj.(string); ok {
		var js map[string]interface{}
		if err := json.Unmarshal([]byte(w), &js); err != nil {
			return w
		}
		return js
	}
	return obj
}

func FormatLog(key string, msg interface{}) (logRecord zap.Field) {
	if p, ok := msg.(proto.Message); ok {
		logRecord = zap.Object(key, &jsonpbObjectMarshaler{pb: p})
	} else {
		logRecord = zap.Any(key, toJSON(msg))
	}

	return
}

var (
	JsonPbMarshaller grpc_logging.JsonPbMarshaler = &jsonpb.Marshaler{}
)

type jsonpbObjectMarshaler struct {
	pb proto.Message
}

func (j *jsonpbObjectMarshaler) MarshalLogObject(e zapcore.ObjectEncoder) error {
	return e.AddReflected("msg", j)
}

func (j *jsonpbObjectMarshaler) MarshalJSON() ([]byte, error) {
	b := &bytes.Buffer{}
	if err := JsonPbMarshaller.Marshal(b, j.pb); err != nil {
		return nil, fmt.Errorf("jsonpb serializer failed: %v", err)
	}
	return b.Bytes(), nil
}

func ToField(key string, val interface{}) zap.Field {
	return zap.Any(key, val)
}

func Caller(level int) string {
	var callerFunc string
	pc, _, _, ok := runtime.Caller(level)
	d := runtime.FuncForPC(pc)

	if ok && d != nil {
		callerFunc = path.Base(d.Name())
	}

	return callerFunc
}
