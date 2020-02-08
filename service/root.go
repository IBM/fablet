package service

import (
	"net/http"
)

func HandleRoot(res http.ResponseWriter, req *http.Request) {
	PlainOutput(res, req, []byte("Fablet is running!"))
}
