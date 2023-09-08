package main

import (
	"context"
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

	"github.com/redis/go-redis/v9"
)

const (
	formatPlain = "plain"
	formatJSON  = "json"

	redisKeyRageCount = "app:rage fucks"

	aboutRaw = `This app has a LONG history, and unfortunately it's referenceable past has experienced {{link_rot}}. In short, a co-worker of mine from many years ago had a silly bash alias that worked like this, and then it died. Soooo, I recreated it... and then it sat for 10+ years untouched. And yea, amazingly it ran without issues all that time.

Recently, {{i}} started some maintenance on some old projects and servers, and I found this app just sitting, still running all this time, but on a VERY old version of PHP, with it's version-controlled source having never been pushed to a remote... So, I pushed the source to {{my_gitHub}}, containerized the app, and now it's running on modern serverless fully-managed compute. Yay!

And, well, then I decided to re-write it in Go! Why? Well, because honestly the old PHP source would have taken a bit of effort just to update to be properly compatible (and secure) with modern PHP versions, and because PHP isn't super well-suited for these kinds of serverless workloads. Oh, yea, and just because I love Go. :)

Anyway, this is silly. Enjoy!

PS: Alias this in your shell environment for a good time, like this:

    alias fuck="curl -Ls http://rage.metroserve.me/?format=plain"

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

	logger = slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{AddSource: true}))
)

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
		msg = fmt.Sprintf("%s {%v}", msg, e.MoreInfo)
	}

	return msg
}

func main() {
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

	ctx := context.Background()

	redisClient := redis.NewClient(&redis.Options{
		Addr:     net.JoinHostPort(redisHost, strconv.Itoa(redisPort)),
		Password: redisPassword,
	})
	defer redisClient.Close()

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		format, err := determineFormat(req)
		if err != nil {
			logger.Info("error determining format", "error", err)
			respondError(w, format, err)
			return
		}

		if req.URL.Path != "/" {
			err := &apiError{
				error: errors.New("unable to find the endpoint you requested"),

				StatusCode: http.StatusNotFound,
				Status:     "NOT_FOUND",
			}
			respondError(w, format, err)
			return
		}

		supportedMethods := []string{http.MethodHead, http.MethodGet, http.MethodPost}
		if !slices.Contains(supportedMethods, req.Method) {
			err := &apiError{
				error: errors.New("the wrong method was called on this endpoint"),

				StatusCode: http.StatusMethodNotAllowed,
				Status:     "METHOD_NOT_ALLOWED",
				MoreInfo: map[string]any{
					"possible_methods": supportedMethods,
				},
			}
			respondError(w, format, err)
			return
		}

		rageCount, err := redisClient.Incr(ctx, redisKeyRageCount).Result()
		if err != nil {
			logger.Error("error incrementing rage count", "error", err)
			respondError(w, format, err)
			return
		}

		var respErr error

		switch format {
		case formatPlain:
			msg := fmt.Sprintf("%v fucks given\n", rageCount)
			respErr = respondPlain(w, http.StatusOK, []byte(msg))
		case formatJSON:
			apiResp := &apiResponse{
				Meta: apiMeta{
					StatusCode: http.StatusOK,
					Status:     "OK",
				},
				Data: apiDataFucks{
					FucksGiven: uint64(rageCount),
				},
			}

			respErr = respondJSON(w, http.StatusOK, apiResp)
		}

		if respErr != nil {
			logger.Error("error while responding", "error", respErr)
		}
	})

	http.HandleFunc("/about/", func(w http.ResponseWriter, req *http.Request) {
		format, err := determineFormat(req)
		if err != nil {
			logger.Info("error determining format", "error", err)
			respondError(w, format, err)
			return
		}

		if req.URL.Path != "/about/" {
			err := &apiError{
				error: errors.New("unable to find the endpoint you requested"),

				StatusCode: http.StatusNotFound,
				Status:     "NOT_FOUND",
			}
			respondError(w, format, err)
			return
		}

		supportedMethods := []string{http.MethodHead, http.MethodGet}
		if !slices.Contains(supportedMethods, req.Method) {
			err := &apiError{
				error: errors.New("the wrong method was called on this endpoint"),

				StatusCode: http.StatusMethodNotAllowed,
				Status:     "METHOD_NOT_ALLOWED",
				MoreInfo: map[string]any{
					"possible_methods": supportedMethods,
				},
			}
			respondError(w, format, err)
			return
		}

		aboutParsed := aboutRaw
		aboutTemplateTexts := map[string]string{
			"link_rot":  aboutTemplateTextLinkRot,
			"i":         aboutTemplateTextI,
			"my_gitHub": aboutTemplateTextMyGitHub,
		}

		for key, replacement := range aboutTemplateTexts {
			tag := fmt.Sprintf("{{%s}}", key)

			aboutParsed = strings.ReplaceAll(aboutParsed, tag, replacement)
		}

		var respErr error

		switch format {
		case formatPlain:
			msg := aboutParsed
			respErr = respondPlain(w, http.StatusOK, []byte(msg))
		case formatJSON:
			aboutData := apiDataAbout{}
			aboutData.Body.Raw = aboutRaw
			aboutData.Body.Parsed = aboutParsed
			aboutData.Templates.Text = aboutTemplateTexts
			aboutData.Templates.Sources = map[string]string{
				"link_rot":  aboutTemplateSourceLinkRot,
				"i":         aboutTemplateSourceI,
				"my_gitHub": aboutTemplateSourceMyGitHub,
			}

			apiResp := &apiResponse{
				Meta: apiMeta{
					StatusCode: http.StatusOK,
					Status:     "OK",
				},
				Data: aboutData,
			}

			respErr = respondJSON(w, http.StatusOK, apiResp)
		}

		if respErr != nil {
			logger.Error("error while responding", "error", respErr)
		}
	})

	serverAddress := net.JoinHostPort(serverHost, strconv.Itoa(serverPort))

	logger.Info("starting HTTP server", "server_address", serverAddress)

	err := http.ListenAndServe(serverAddress, nil)
	if err != nil {
		logger.Error("error listening and serving", "error", err)
	}
}

func determineFormat(req *http.Request) (string, error) {
	availableFormats := []string{formatPlain, formatJSON}

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

func respondPlain(w http.ResponseWriter, statusCode int, data []byte) error {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(statusCode)

	_, err := w.Write(data)

	return err
}

func respondJSON(w http.ResponseWriter, statusCode int, data any) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)

	return json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, format string, err error) {
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

	var respErr error

	switch format {
	case formatPlain:
		respErr = respondPlain(w, apiErr.StatusCode, []byte(apiErr.Error()))
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

		respErr = respondJSON(w, apiErr.StatusCode, apiResp)
	}

	if respErr != nil {
		logger.Error("error while responding", "error", respErr)
	}
}
