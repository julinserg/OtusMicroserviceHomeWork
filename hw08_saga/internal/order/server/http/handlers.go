package order_internalhttp

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gofrs/uuid"
	order_app "github.com/julinserg/julinserg/OtusMicroserviceHomeWork/hw08_saga/internal/order/app"
)

func hellowHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("This is my orsers service!"))
}

type ordersHandler struct {
	logger   Logger
	storage  Storage
	srvOrder SrvOrder
}

type ResponseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (h *ordersHandler) createHandler(w http.ResponseWriter, r *http.Request) {
	switch method := r.Method; method {
	case "POST":
		h.CreateOrder(w, r)
	}
}

func (h *ordersHandler) countHandler(w http.ResponseWriter, r *http.Request) {
	switch method := r.Method; method {
	case "GET":
		h.CountOrder(w, r)
	}
}

func (h *ordersHandler) WriteResponseError(w http.ResponseWriter, resp *ResponseError) {
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

func (h *ordersHandler) checkErrorAndSendResponse(requestId string, err error, code int, w http.ResponseWriter) bool {
	if err != nil {
		resp := &ResponseError{}
		resp.Code = code
		resp.Message = err.Error()
		h.logger.Error(resp.Message)
		h.storage.UpdateRequest(order_app.Request{Id: requestId, Code: code, ErrorText: err.Error()})
		w.WriteHeader(code)
		h.WriteResponseError(w, resp)
		return false
	}
	return true
}

func (h *ordersHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	requestId := r.Header.Get("X-Request-Id")
	if len(requestId) == 0 {
		h.checkErrorAndSendResponse("", fmt.Errorf("Header X-Request-Id not set"), http.StatusBadRequest, w)
		return
	}
	reqId, err := h.storage.GetOrCreateRequest(requestId)
	if !reqId.IsNew {
		if len(reqId.ErrorText) != 0 {
			resp := &ResponseError{}
			resp.Code = reqId.Code
			resp.Message = reqId.ErrorText
			h.logger.Error(resp.Message)
			w.WriteHeader(reqId.Code)
			h.WriteResponseError(w, resp)
			return
		} else {
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(reqId.Code)
			return
		}
	}

	body, err := ioutil.ReadAll(r.Body)
	if !h.checkErrorAndSendResponse(requestId, err, http.StatusBadRequest, w) {
		return
	}

	order := &order_app.Order{}
	err = json.Unmarshal(body, order)
	if !h.checkErrorAndSendResponse(requestId, err, http.StatusBadRequest, w) {
		return
	}

	orderUUID, err := uuid.NewV4()
	if !h.checkErrorAndSendResponse(requestId, err, http.StatusInternalServerError, w) {
		return
	}
	order.Id = orderUUID.String()
	err = h.srvOrder.CreateOrder(*order)
	if !h.checkErrorAndSendResponse(requestId, err, http.StatusInternalServerError, w) {
		return
	}
	h.storage.UpdateRequest(order_app.Request{Id: requestId, Code: http.StatusOK, ErrorText: ""})
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	return
}

func (h *ordersHandler) CountOrder(w http.ResponseWriter, r *http.Request) {
	result, err := h.srvOrder.GetOrdersCount()
	if !h.checkErrorAndSendResponse("", err, http.StatusInternalServerError, w) {
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{ \"count\" :" + strconv.Itoa(result) + "}"))
	return
}
