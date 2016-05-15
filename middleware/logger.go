package middleware

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"time"
)

// LoggerConfig contains logger info such as formatting and output writer
type LoggerConfig struct {
	Output      io.Writer
	EnableColor bool
}

type loggerInterptor struct {
	w          http.ResponseWriter
	statusCode int
}

func (li *loggerInterptor) Header() http.Header {
	return li.w.Header()
}

func (li *loggerInterptor) Write(data []byte) (int, error) {
	return li.w.Write(data)
}

func (li *loggerInterptor) WriteHeader(statusCode int) {
	li.w.WriteHeader(statusCode)
	li.statusCode = statusCode
}

// Logger - writes information about http requests to the log
func Logger(config LoggerConfig, next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Start timer
		start := time.Now()

		// Process request
		li := &loggerInterptor{w: w, statusCode: 200}
		next.ServeHTTP(li, r)
		method := r.Method
		statusCode := li.statusCode
		path := r.URL.Path

		// Stop timer
		end := time.Now()
		latency := end.Sub(start)

		// Parse remote IP
		remoteAddr := r.RemoteAddr
		if ip := r.Header.Get("X-Real-IP"); ip != "" {
			remoteAddr = ip
		} else if ip = r.Header.Get("X-Forwarded-For"); ip != "" {
			remoteAddr = ip
		} else {
			remoteAddr, _, _ = net.SplitHostPort(remoteAddr)
		}

		//Parse API version
		version := r.Header.Get("accept-version")
		if len(version) != 0 {
			version = fmt.Sprintf("/v%s", version)
		}

		//Write log string to writer
		if config.EnableColor {
			fmt.Fprintf(config.Output, "[REQUEST] %v | %3s | %22v | %24s | %17s %s%s\n",
				grey(end.Format("2006/01/02 - 15:04:05")),
				colorForStatus(statusCode),
				grey(latency.String()),
				grey(remoteAddr),
				colorForMethod(method),
				version,
				white(path),
			)
		} else {
			fmt.Fprintf(config.Output, "[REQUEST] %v | %3s | %13v | %15s | %7s %s%s\n",
				end.Format("2006/01/02 - 15:04:05"),
				strconv.Itoa(statusCode),
				latency,
				remoteAddr,
				method,
				version,
				path,
			)
		}
	}
}

func colorForStatus(code int) string {
	switch {
	case code >= 200 && code < 300:
		return green(strconv.Itoa(code))
	case code >= 300 && code < 400:
		return cyan(strconv.Itoa(code))
	case code >= 400 && code < 500:
		return yellow(strconv.Itoa(code))
	default:
		return red(strconv.Itoa(code))
	}
}

func colorForMethod(method string) string {
	switch method {
	case "GET":
		return blue(method)
	case "POST":
		return cyan(method)
	case "PUT":
		return yellow(method)
	case "DELETE":
		return red(method)
	case "PATCH":
		return green(method)
	case "HEAD":
		return magenta(method)
	case "OPTIONS":
		return white(method)
	default:
		return white(method)
	}
}

const (
	blackColor   = "30"
	redColor     = "31"
	greenColor   = "32"
	yellowColor  = "33"
	blueColor    = "34"
	magentaColor = "35"
	cyanColor    = "36"
	whiteColor   = "37"
	greyColor    = "90"
)

func grey(msg string) string {
	return color(msg, greyColor)
}

func blue(msg string) string {
	return color(msg, blueColor)
}

func cyan(msg string) string {
	return color(msg, cyanColor)
}

func yellow(msg string) string {
	return color(msg, yellowColor)
}

func red(msg string) string {
	return color(msg, redColor)
}

func green(msg string) string {
	return color(msg, greenColor)
}

func magenta(msg string) string {
	return color(msg, magentaColor)
}

func white(msg string) string {
	return color(msg, whiteColor)
}

func color(msg string, n string) string {
	b := new(bytes.Buffer)
	b.WriteString("\x1b[")
	b.WriteString(n)
	b.WriteString("m")
	return fmt.Sprintf("%s%v\x1b[0m", b.String(), msg)
}
