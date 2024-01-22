package order_internalhttp

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	order_app "github.com/julinserg/julinserg/OtusMicroserviceHomeWork/hw08_saga/internal/order/app"
)

func hellowHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("This is my orsers service!"))
}

type ordersHandler struct {
	logger   Logger
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

func (h *ordersHandler) cancelHandler(w http.ResponseWriter, r *http.Request) {
	switch method := r.Method; method {
	case "POST":
		h.CancelOrder(w, r)
	}
}

func (h *ordersHandler) statusHandler(w http.ResponseWriter, r *http.Request) {
	switch method := r.Method; method {
	case "GET":
		h.StatusOrder(w, r)
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

func (h *ordersHandler) checkErrorAndSendResponse(err error, code int, w http.ResponseWriter) bool {
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

func (h *ordersHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if !h.checkErrorAndSendResponse(err, http.StatusBadRequest, w) {
		return
	}

	order := &order_app.Order{}
	err = json.Unmarshal(body, order)
	if !h.checkErrorAndSendResponse(err, http.StatusBadRequest, w) {
		return
	}
	err = h.srvOrder.CreateOrder(*order)
	if !h.checkErrorAndSendResponse(err, http.StatusInternalServerError, w) {
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	return
}

func (h *ordersHandler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if !h.checkErrorAndSendResponse(err, http.StatusBadRequest, w) {
		return
	}

	orderID := &order_app.OrderID{}
	err = json.Unmarshal(body, orderID)
	if !h.checkErrorAndSendResponse(err, http.StatusBadRequest, w) {
		return
	}
	err = h.srvOrder.CancelOrder(orderID.Id)
	if !h.checkErrorAndSendResponse(err, http.StatusInternalServerError, w) {
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	return
}

func (h *ordersHandler) StatusOrder(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	status, err := h.srvOrder.StatusOrder(id)
	if !h.checkErrorAndSendResponse(err, http.StatusInternalServerError, w) {
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{ \"status\" :" + status + "}"))
	return
}
