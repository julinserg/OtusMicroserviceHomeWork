package internalhttp

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/julinserg/julinserg/OtusMicroserviceHomeWork/hw04_prometheus/internal/app"
)

func hellowHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("This is my user service!"))
}

type userHandler struct {
	logger  Logger
	storage Storage
}

type ResponseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (h *userHandler) commonHandler(w http.ResponseWriter, r *http.Request) {
	switch method := r.Method; method {
	case "POST":
		h.CreateUser(w, r)
	case "PUT":
		h.UpdateUser(w, r)
	case "GET":
		h.FindUserById(w, r)
	case "DELETE":
		h.DeleteUser(w, r)
	}
}

func (h *userHandler) WriteResponseError(w http.ResponseWriter, resp *ResponseError) {
	resBuf, err := json.Marshal(resp)
	if err != nil {
		h.logger.Error("response marshal error: " + err.Error())
	}
	_, err = w.Write(resBuf)
	if err != nil {
		h.logger.Error("response marshal error: " + err.Error())
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return
}

func (h *userHandler) checkErrorAndSendResponse(err error, code int, w http.ResponseWriter) bool {
	if err != nil {
		resp := &ResponseError{}
		resp.Code = code
		resp.Message = err.Error()
		h.logger.Error(resp.Message)
		w.WriteHeader(code)
		h.WriteResponseError(w, resp)
		return false
	}
	return true
}

func (h *userHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if !h.checkErrorAndSendResponse(err, http.StatusBadRequest, w) {
		return
	}

	user := &app.User{}
	err = json.Unmarshal(body, user)
	if !h.checkErrorAndSendResponse(err, http.StatusBadRequest, w) {
		return
	}
	err = h.storage.CreateUser(*user)
	if !h.checkErrorAndSendResponse(err, http.StatusInternalServerError, w) {
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	return
}

func (h *userHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	url := strings.Split(r.URL.Path, "/")
	userId := url[len(url)-1]
	err := h.storage.DeleteUser(userId)
	if !h.checkErrorAndSendResponse(err, http.StatusInternalServerError, w) {
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusNoContent)
}

func (h *userHandler) FindUserById(w http.ResponseWriter, r *http.Request) {
	url := strings.Split(r.URL.Path, "/")
	userId := url[len(url)-1]
	user, err := h.storage.FindUserById(userId)
	if !h.checkErrorAndSendResponse(err, http.StatusInternalServerError, w) {
		return
	}
	resBuf, err := json.Marshal(user)
	if err != nil {
		h.logger.Error("response marshal error: " + err.Error())
	}
	_, err = w.Write(resBuf)
	if err != nil {
		h.logger.Error("response marshal error: " + err.Error())
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}

func (h *userHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	url := strings.Split(r.URL.Path, "/")
	userId := url[len(url)-1]
	userIdInt, _ := strconv.Atoi(userId)
	body, err := ioutil.ReadAll(r.Body)
	if !h.checkErrorAndSendResponse(err, http.StatusBadRequest, w) {
		return
	}

	user := &app.User{}
	err = json.Unmarshal(body, user)
	if !h.checkErrorAndSendResponse(err, http.StatusBadRequest, w) {
		return
	}
	user.Id = int64(userIdInt)
	err = h.storage.UpdateUser(*user)
	if !h.checkErrorAndSendResponse(err, http.StatusInternalServerError, w) {
		return
	}
}
