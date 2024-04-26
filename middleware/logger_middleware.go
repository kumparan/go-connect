package middleware

import (
	"bytes"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasttemplate"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	headerXForwardedFor = "X-Forwarded-For"
	headerXRealIP       = "X-Real-Ip"
	headerXRequestID    = "X-Request-Id"
	headerContentLength = "Content-Length"
)

// CustomTagFunc :nodoc:
type CustomTagFunc func(r *http.Request, buf *bytes.Buffer) (int, error)

// LoggerMiddleware :nodoc:
type LoggerMiddleware struct {
	template      *fasttemplate.Template
	customTagFunc CustomTagFunc
	timeFormat    string
	logger        *logrus.Logger
}

// LoggerConfig :nodoc:
type LoggerConfig struct {
	Format        string
	CustomTagFunc CustomTagFunc
	TimeFormat    string
}

// NewLoggerMiddleware :nodoc:
func NewLoggerMiddleware(cfg *LoggerConfig) *LoggerMiddleware {
	return &LoggerMiddleware{
		template:      fasttemplate.New(cfg.Format, "${", "}"),
		customTagFunc: cfg.CustomTagFunc,
		logger:        logrus.StandardLogger(),
		timeFormat:    cfg.TimeFormat,
	}
}

func (l *LoggerMiddleware) Intercept(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		start := time.Now()
		next.ServeHTTP(writer, req)
		stop := time.Now()

		buf := bytes.NewBuffer([]byte{})
		if _, err := l.template.ExecuteFunc(buf, func(w io.Writer, tag string) (int, error) {
			switch tag {
			case "custom":
				if l.customTagFunc == nil {
					return 0, nil
				}
				return l.customTagFunc(req, buf)
			case "time_unix":
				return buf.WriteString(strconv.FormatInt(time.Now().Unix(), 10))
			case "time_unix_milli":
				// go 1.17 or later, it supports time#UnixMilli()
				return buf.WriteString(strconv.FormatInt(time.Now().UnixNano()/1000000, 10))
			case "time_unix_micro":
				// go 1.17 or later, it supports time#UnixMicro()
				return buf.WriteString(strconv.FormatInt(time.Now().UnixNano()/1000, 10))
			case "time_unix_nano":
				return buf.WriteString(strconv.FormatInt(time.Now().UnixNano(), 10))
			case "time_rfc3339":
				return buf.WriteString(time.Now().Format(time.RFC3339))
			case "time_rfc3339_nano":
				return buf.WriteString(time.Now().Format(time.RFC3339Nano))
			case "time_custom":
				return buf.WriteString(time.Now().Format(l.timeFormat))
			case "id":
				id := req.Header.Get(headerXRequestID)
				if id == "" {
					id = req.Header.Get(headerXRequestID)
				}
				return buf.WriteString(id)
			case "remote_ip":
				return buf.WriteString(ipExtractor(req))
			case "host":
				return buf.WriteString(req.Host)
			case "uri":
				return buf.WriteString(req.RequestURI)
			case "method":
				return buf.WriteString(req.Method)
			case "path":
				p := req.URL.Path
				if p == "" {
					p = "/"
				}
				return buf.WriteString(p)
			case "protocol":
				return buf.WriteString(req.Proto)
			case "referer":
				return buf.WriteString(req.Referer())
			case "user_agent":
				return buf.WriteString(req.UserAgent())
			case "latency":
				l := stop.Sub(start)
				return buf.WriteString(strconv.FormatInt(int64(l), 10))
			case "latency_human":
				return buf.WriteString(stop.Sub(start).String())
			case "bytes_in":
				cl := req.Header.Get(headerContentLength)
				if cl == "" {
					cl = "0"
				}
				return buf.WriteString(cl)
			default:
				switch {
				case strings.HasPrefix(tag, "header:"):
					return buf.Write([]byte(req.Header.Get(tag[7:])))
				case strings.HasPrefix(tag, "query:"):
					params := req.URL.Query()[tag[6:]]
					return buf.Write([]byte(strings.Join(params, ",")))
				case strings.HasPrefix(tag, "form:"):
					return buf.Write([]byte(req.FormValue(tag[5:])))
				case strings.HasPrefix(tag, "cookie:"):
					cookie, err := req.Cookie(tag[7:])
					if err == nil {
						return buf.Write([]byte(cookie.Value))
					}
				}
			}
			return 0, nil
		}); err != nil {
			return
		}
		_, err := l.logger.Out.Write(buf.Bytes())
		if err != nil {
			logrus.Error(err)
		}
	})
}

func ipExtractor(r *http.Request) string {
	if ip := r.Header.Get(headerXForwardedFor); ip != "" {
		i := strings.IndexAny(ip, ",")
		if i > 0 {
			xffip := strings.TrimSpace(ip[:i])
			xffip = strings.TrimPrefix(xffip, "[")
			xffip = strings.TrimSuffix(xffip, "]")
			return xffip
		}
		return ip
	}
	if ip := r.Header.Get(headerXRealIP); ip != "" {
		ip = strings.TrimPrefix(ip, "[")
		ip = strings.TrimSuffix(ip, "]")
		return ip
	}

	return ""
}
