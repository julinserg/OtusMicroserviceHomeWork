package pay_internalhttp

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gofrs/uuid"
	pay_app "github.com/julinserg/julinserg/OtusMicroserviceHomeWork/hw08_saga/internal/pay/app"
)

func hellowHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("This is my auth service 3!"))
}

type userHandler struct {
	logger   Logger
	storage  Storage
	sessions map[string]pay_app.UserAuth
}

type ResponseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
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

func (h *userHandler) registerHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if !h.checkErrorAndSendResponse(err, http.StatusBadRequest, w) {
		return
	}

	user := &pay_app.UserAuth{}
	err = json.Unmarshal(body, user)
	if !h.checkErrorAndSendResponse(err, http.StatusBadRequest, w) {
		return
	}
	id, err := h.storage.RegisterUser(*user)
	if !h.checkErrorAndSendResponse(err, http.StatusBadRequest, w) {
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("{\"id\": %d}", id)))
	return
}

func (h *userHandler) loginHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if !h.checkErrorAndSendResponse(err, http.StatusBadRequest, w) {
		return
	}

	login := &pay_app.LoginAuth{}
	err = json.Unmarshal(body, login)
	if !h.checkErrorAndSendResponse(err, http.StatusBadRequest, w) {
		return
	}
	user, err := h.storage.GetUser(login.Login, login.Password)
	if !h.checkErrorAndSendResponse(err, http.StatusUnauthorized, w) {
		return
	}
	if len(user.Login) == 0 {
		h.checkErrorAndSendResponse(fmt.Errorf("User not found"), http.StatusUnauthorized, w)
		return
	}
	uuid, err := uuid.NewV4()
	uuidStr := uuid.String()
	h.sessions[uuidStr] = user

	cookie := http.Cookie{Name: "session_id", Value: uuidStr, HttpOnly: true}
	http.SetCookie(w, &cookie)

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{\"status\": \"ok\"}"))
	return
}

func (h *userHandler) authHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if !h.checkErrorAndSendResponse(err, http.StatusUnauthorized, w) {
		return
	}
	user, ok := h.sessions[cookie.Value]
	if !ok {
		h.checkErrorAndSendResponse(fmt.Errorf("User not found"), http.StatusUnauthorized, w)
		return
	}
	resBuf, err := json.Marshal(user)
	if !h.checkErrorAndSendResponse(err, http.StatusInternalServerError, w) {
		return
	}

	w.Header().Set("X-User-Id", strconv.Itoa(int(user.Id)))
	w.Header().Set("X-User", user.Login)
	w.Header().Set("X-Email", user.Email)
	w.Header().Set("X-First-Name", user.FirstName)
	w.Header().Set("X-Last-Name", user.LastName)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	_, err = w.Write(resBuf)
	if !h.checkErrorAndSendResponse(err, http.StatusInternalServerError, w) {
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *userHandler) logoutHandler(w http.ResponseWriter, r *http.Request) {

	cookie := http.Cookie{Name: "session_id", Value: ""}
	http.SetCookie(w, &cookie)

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{\"status\": \"ok\"}"))
	return
}
