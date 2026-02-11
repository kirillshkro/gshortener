package shortener

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog"
)

func HandlerWithLog(h http.Handler) http.Handler {
	logger := zerolog.New(os.Stderr)
	infoFn := func(resp http.ResponseWriter, req *http.Request) {
		uri := req.RequestURI
		method := req.Method
		tRec := time.Now()
		logger.Info().Msg(fmt.Sprintf("URI request: %s\t, method: %s\t, time: %v\n", uri, method, tRec))
	}
	return http.HandlerFunc(infoFn)
}
