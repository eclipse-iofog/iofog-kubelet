package api

import (
	"net/http"
)

func FogControllerHandlerStopFunc(stopFunc func(nodeId string)) http.HandlerFunc {
	return handleError(func(w http.ResponseWriter, req *http.Request) error {
		nodeId := req.FormValue("uuid")
		go stopFunc(nodeId)

		return nil
	})
}

func FogControllerHandlerStartFunc(startFunc func(nodeId string)) http.HandlerFunc {
	return handleError(func(w http.ResponseWriter, req *http.Request) error {
		nodeId := req.FormValue("uuid")
		go startFunc(nodeId)

		return nil
	})
}