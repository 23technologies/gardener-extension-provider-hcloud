package mock

import (
	"strings"

	"net/http"
)

const (
	token = "dummy-token"
)

func SetupTestTokenEndpointOnMux(mux *http.ServeMux) {
	mux.HandleFunc("/testtokenendpoint", func(res http.ResponseWriter, req *http.Request) {

		auth_bearer_token := req.Header["Authorization"][0]
		auth_bearer_token = strings.Split(auth_bearer_token, " ")[1]
		if auth_bearer_token == token {
			res.WriteHeader(http.StatusOK)
		}	else {
			res.WriteHeader(http.StatusForbidden)
		}

	})
}
