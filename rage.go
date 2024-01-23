package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	formatPlain = "plain"
	formatJSON  = "json"

	redisKeyRageCount = "app:rage fucks"

	aboutRaw = `This app has a LONG history, and unfortunately it's referenceable past has experienced {{link_rot}}. In short, a co-worker of mine from many years ago had a silly bash alias that worked by pinging a service like this, and then it died. Soooo, I recreated it... and then it sat for 10+ years untouched. And yea, amazingly it ran without issues all that time.

Recently, {{i}} started some maintenance on some old projects and servers, and I found this app just sitting, still running all this time, but on a VERY old version of PHP, with it's version-controlled source having never been pushed to a remote... So, I pushed the source to {{my_gitHub}}, containerized the app, and now it's running on modern serverless fully-managed compute. Yay!

And, well, then I decided to re-write it in Go! Why? Well, because honestly the old PHP source would have taken a bit of effort just to update to be properly compatible (and secure) with modern PHP versions, and because PHP isn't super well-suited for these kinds of serverless workloads. Oh, yea, and just because I love Go. :)

Anyway, this is silly. Enjoy!

PS: Alias this in your shell environment for a good time, like this:

    alias fuck="curl -Ls {{app_base_url}}?format=plain"

`
	aboutTemplateTextLinkRot    = "link rot"
	aboutTemplateTextI          = "I (Trevor Suarez (Rican7))"
	aboutTemplateTextMyGitHub   = "my GitHub"
	aboutTemplateSourceLinkRot  = "https://en.wikipedia.org/wiki/Link_rot"
	aboutTemplateSourceI        = "https://trevorsuarez.com/"
	aboutTemplateSourceMyGitHub = "https://github.com/Rican7/rage"
)

var (
	serverHost    = "0.0.0.0"
	serverPort    = 80
	redisHost     = "127.0.0.1"
	redisPort     = 6379
	redisPassword = ""

	appBaseURL = "https://rage.metroserve.me/"

	apiErrNotFound = apiError{
		error: errors.New("unable to find the endpoint you requested"),

		StatusCode: http.StatusNotFound,
		Status:     "NOT_FOUND",
	}
)

type middleware func(next http.HandlerFunc) http.HandlerFunc

type apiResponse struct {
	Meta apiMeta `json:"meta"`
	Data any     `json:"data"`
}

func (r apiResponse) MarshalJSON() ([]byte, error) {
	if r.Meta.MoreInfo == nil {
		r.Meta.MoreInfo = map[string]any{}
	}
	if r.Data == nil {
		r.Data = struct{}{}
	}

	type alias apiResponse
	return json.Marshal(alias(r))
}

type apiMeta struct {
	StatusCode int            `json:"status_code"`
	Status     string         `json:"status"`
	Message    string         `json:"message"`
	MoreInfo   map[string]any `json:"more_info"`
}

type apiDataFucks struct {
	FucksGiven uint64 `json:"fucks_given"`
}

type apiDataAbout struct {
	Body struct {
		Raw    string `json:"raw"`
		Parsed string `json:"parsed"`
	} `json:"body"`
	Templates struct {
		Text    map[string]string `json:"text"`
		Sources map[string]string `json:"sources"`
	} `json:"templates"`
}

type apiError struct {
	error

	StatusCode int
	Status     string
	MoreInfo   map[string]any
}

func (e *apiError) Error() string {
	msg := e.error.Error()

	if e.Status != "" {
		msg = fmt.Sprintf("%s - %s", e.Status, msg)
	}

	if e.MoreInfo != nil {
		// The simplest, most readable text encoding of a generic value is JSON
		v, err := json.Marshal(e.MoreInfo)
		if err == nil {
			msg = fmt.Sprintf("%s; %s", msg, v)
		}
	}

	return msg
}

type routingErrorResponseWriter struct {
	logger *slog.Logger
	http.ResponseWriter

	format      string
	statusCode  int
	intercepted bool
}

func (w *routingErrorResponseWriter) WriteHeader(statusCode int) {
	if w.intercepted {
		return
	}

	w.statusCode = statusCode

	switch statusCode {
	case 404:
		respondError(w.logger, w.ResponseWriter, w.format, &apiErrNotFound)
		w.intercepted = true
	case 405:
		supportedMethods := strings.Split(
			w.ResponseWriter.Header().Get(http.CanonicalHeaderKey("Allow")),
			", ",
		)

		err := &apiError{
			error: errors.New("the wrong method was called on this endpoint"),

			StatusCode: http.StatusMethodNotAllowed,
			Status:     "METHOD_NOT_ALLOWED",
			MoreInfo: map[string]any{
				"possible_methods": supportedMethods,
			},
		}

		respondError(w.logger, w.ResponseWriter, w.format, err)
		w.intercepted = true
	default:
		w.ResponseWriter.WriteHeader(statusCode)
	}
}

func (w *routingErrorResponseWriter) Write(data []byte) (int, error) {
	if w.intercepted {
		return 0, nil
	}

	return w.ResponseWriter.Write(data)
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{AddSource: true}))

	initConfig(logger)

	redisClient := redis.NewClient(&redis.Options{
		Addr:     net.JoinHostPort(redisHost, strconv.Itoa(redisPort)),
		Password: redisPassword,
	})
	defer redisClient.Close()

	mainRoute := mainHandler(logger, redisClient)
	aboutRoute := aboutHandler(logger)

	router := http.NewServeMux()

	router.HandleFunc("GET /{$}", mainRoute)
	router.HandleFunc("POST /{$}", mainRoute)
	router.HandleFunc("GET /about/{$}", aboutRoute)

	serverAddress := net.JoinHostPort(serverHost, strconv.Itoa(serverPort))

	logMiddle := loggerMiddleware(logger)
	routingErrorResponseMiddle := routingErrorResponseMiddleware(logger)

	globalHandler := logMiddle(routingErrorResponseMiddle(router.ServeHTTP))

	logger.Info("starting HTTP server", "server_address", serverAddress)

	err := http.ListenAndServe(serverAddress, globalHandler)
	if err != nil {
		logger.Error("error listening and serving", "error", err)
	}
}

func initConfig(logger *slog.Logger) {
	if envServerHost := os.Getenv("HOST"); envServerHost != "" {
		serverHost = envServerHost
	}
	if envServerPort := os.Getenv("PORT"); envServerPort != "" {
		v, err := strconv.Atoi(envServerPort)
		if err != nil {
			logger.Error("error parsing env PORT", "error", err)
			return
		}

		serverPort = v
	}
	if envRedisHost := os.Getenv("REDIS_HOST"); envRedisHost != "" {
		redisHost = envRedisHost
	}
	if envRedisPort := os.Getenv("REDIS_PORT"); envRedisPort != "" {
		v, err := strconv.Atoi(envRedisPort)
		if err != nil {
			logger.Error("error parsing env REDIS_PORT", "error", err)
			return
		}

		redisPort = v
	}
	if envRedisPassword := os.Getenv("REDIS_PASSWORD"); envRedisPassword != "" {
		redisPassword = envRedisPassword
	}
	if envAppBaseURL := os.Getenv("APP_BASE_URL"); envAppBaseURL != "" {
		appBaseURL = envAppBaseURL
	}
}

func loggerMiddleware(logger *slog.Logger) middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, req *http.Request) {
			startTime := time.Now()

			ll := logger.With(
				"request_method", req.Method,
				"request_url", req.URL.String(),
				"request_user_agent", req.Header.Get("User-Agent"),
			)

			ll.Info("handling incoming request")

			next.ServeHTTP(w, req)

			ll.Info("finished handling request", "handle_duration", time.Now().Sub(startTime).String())
		}
	}
}

func routingErrorResponseMiddleware(logger *slog.Logger) middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, req *http.Request) {
			format, err := determineFormat(req)
			if err != nil {
				logger.Info("error determining format", "error", err)
				respondError(logger, w, format, err)
				return
			}

			w = &routingErrorResponseWriter{
				logger:         logger,
				ResponseWriter: w,

				format: format,
			}

			next.ServeHTTP(w, req)
		}
	}
}

func mainHandler(logger *slog.Logger, redisClient *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		format, err := determineFormat(req)
		if err != nil {
			logger.Info("error determining format", "error", err)
			respondError(logger, w, format, err)
			return
		}

		rageCount, err := redisClient.Incr(ctx, redisKeyRageCount).Result()
		if err != nil {
			logger.Error("error incrementing rage count", "error", err)
			respondError(logger, w, format, err)
			return
		}

		switch format {
		case formatPlain:
			msg := fmt.Sprintf("%v fucks given\n", rageCount)
			respondPlain(logger, w, http.StatusOK, msg)
		case formatJSON:
			apiResp := &apiResponse{
				Meta: apiMeta{
					StatusCode: http.StatusOK,
					Status:     "OK",
					MoreInfo: map[string]any{
						"about": appBaseURL + "about/",
					},
				},
				Data: apiDataFucks{
					FucksGiven: uint64(rageCount),
				},
			}

			respondJSON(logger, w, http.StatusOK, apiResp)
		}
	}
}

func aboutHandler(logger *slog.Logger) http.HandlerFunc {
	// Build the about data ONCE, since it's always the same
	aboutParsed := aboutRaw
	aboutTemplateTexts := map[string]string{
		"link_rot":     aboutTemplateTextLinkRot,
		"i":            aboutTemplateTextI,
		"my_gitHub":    aboutTemplateTextMyGitHub,
		"app_base_url": appBaseURL,
	}

	for key, replacement := range aboutTemplateTexts {
		tag := fmt.Sprintf("{{%s}}", key)

		aboutParsed = strings.ReplaceAll(aboutParsed, tag, replacement)
	}

	aboutData := apiDataAbout{}
	aboutData.Body.Raw = aboutRaw
	aboutData.Body.Parsed = aboutParsed
	aboutData.Templates.Text = aboutTemplateTexts
	aboutData.Templates.Sources = map[string]string{
		"link_rot":     aboutTemplateSourceLinkRot,
		"i":            aboutTemplateSourceI,
		"my_gitHub":    aboutTemplateSourceMyGitHub,
		"app_base_url": appBaseURL,
	}

	apiResp := &apiResponse{
		Meta: apiMeta{
			StatusCode: http.StatusOK,
			Status:     "OK",
		},
		Data: aboutData,
	}

	return func(w http.ResponseWriter, req *http.Request) {
		format, err := determineFormat(req)
		if err != nil {
			logger.Info("error determining format", "error", err)
			respondError(logger, w, format, err)
			return
		}

		switch format {
		case formatPlain:
			respondPlain(logger, w, http.StatusOK, aboutParsed)
		case formatJSON:
			respondJSON(logger, w, http.StatusOK, apiResp)
		}
	}
}

func determineFormat(req *http.Request) (string, error) {
	availableFormats := []string{formatPlain, formatJSON}

	// TODO: This should PROBABLY actually be based on content-negotiation
	// ("Accept" header and the like), but that can get a bit complicated...
	//
	// This is fine for now, and how it's worked for the past 10+ years 😅
	format := req.FormValue("format")
	if format == "" {
		format = formatJSON
	}

	if !slices.Contains(availableFormats, format) {
		err := &apiError{
			error: errors.New("invalid format type"),

			StatusCode: http.StatusBadRequest,
			Status:     "INVALID_FORMAT",
			MoreInfo: map[string]any{
				"valid_formats": availableFormats,
			},
		}
		return "", err
	}

	return format, nil
}

func respondPlain(logger *slog.Logger, w http.ResponseWriter, statusCode int, data string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(statusCode)

	if !strings.HasSuffix(data, "\n") {
		data += "\n"
	}

	_, err := w.Write([]byte(data))

	if err != nil {
		handleRespondError(logger, err)
	}
}

func respondJSON(logger *slog.Logger, w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)

	err := json.NewEncoder(w).Encode(data)

	if err != nil {
		handleRespondError(logger, err)
	}
}

func handleRespondError(logger *slog.Logger, err error) {
	if err == nil {
		return
	}

	logger.Error("error while responding", "error", err)
}

func respondError(logger *slog.Logger, w http.ResponseWriter, format string, err error) {
	var apiErr *apiError

	// If the err is an apiError, it'll be set to apiErr, otherwise...
	if !errors.As(err, &apiErr) {
		apiErr = &apiError{error: err}
	}

	if apiErr.StatusCode == 0 {
		apiErr.StatusCode = 500
	}

	if apiErr.Status == "" {
		apiErr.Status = "UNEXPECTED_ERROR"
	}

	logger.Debug("responding with error", "error", err)

	switch format {
	case formatPlain:
		respondPlain(logger, w, apiErr.StatusCode, apiErr.Error())
	case formatJSON:
		fallthrough
	default:
		apiResp := &apiResponse{
			Meta: apiMeta{
				StatusCode: apiErr.StatusCode,
				Status:     apiErr.Status,
				Message:    apiErr.error.Error(),
				MoreInfo:   apiErr.MoreInfo,
			},
		}

		respondJSON(logger, w, apiErr.StatusCode, apiResp)
	}
}
