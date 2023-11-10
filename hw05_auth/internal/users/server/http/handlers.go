package users_internalhttp

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	users_app "github.com/julinserg/julinserg/OtusMicroserviceHomeWork/hw05_auth/internal/users/app"
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
	case "PUT":
		h.CreateOrUpdateUser(w, r)
	case "GET":
		h.FindUserById(w, r)
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

func (h *userHandler) FindUserById(w http.ResponseWriter, r *http.Request) {

	userId := w.Header().Get("X-UserId")
	if len(userId) == 0 {
		h.checkErrorAndSendResponse(fmt.Errorf("Not authenticated"), http.StatusInternalServerError, w)
		return
	}
	userFull := &users_app.UserFull{}
	id, _ := strconv.Atoi(userId)
	userFull.Id = int64(id)
	userFull.Login = w.Header().Get("X-User")
	userFull.FirstName = w.Header().Get("X-First-Name")
	userFull.LastName = w.Header().Get("X-Last-Name")
	userFull.Email = w.Header().Get("X-Email")

	userProfile, err := h.storage.FindUserById(userId)
	if userProfile.Id != 0 {
		userFull.Age = userProfile.Age
		userFull.Phone = userProfile.Phone
		userFull.AvatarUri = userProfile.AvatarUri
	}

	resBuf, err := json.Marshal(userFull)
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

func (h *userHandler) CreateOrUpdateUser(w http.ResponseWriter, r *http.Request) {
	userId := w.Header().Get("X-UserId")
	if len(userId) == 0 {
		h.checkErrorAndSendResponse(fmt.Errorf("Not authenticated"), http.StatusInternalServerError, w)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if !h.checkErrorAndSendResponse(err, http.StatusBadRequest, w) {
		return
	}

	user := &users_app.User{}
	err = json.Unmarshal(body, user)
	if !h.checkErrorAndSendResponse(err, http.StatusBadRequest, w) {
		return
	}
	id, _ := strconv.Atoi(userId)
	user.Id = int64(id)

	userProfile, err := h.storage.FindUserById(userId)
	if userProfile.Id != 0 {
		err = h.storage.UpdateUser(*user)
		if !h.checkErrorAndSendResponse(err, http.StatusInternalServerError, w) {
			return
		}
	} else {
		err = h.storage.CreateUser(*user)
		if !h.checkErrorAndSendResponse(err, http.StatusInternalServerError, w) {
			return
		}
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	return
}
