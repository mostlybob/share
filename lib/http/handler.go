// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/shuLhan/share/lib/debug"
	"github.com/shuLhan/share/lib/strings"
)

type handler struct {
	reqType RequestType
	resType ResponseType
	cb      Callback
}

func (h *handler) call(res http.ResponseWriter, req *http.Request) {
	var (
		e       error
		reqBody []byte
	)

	switch h.reqType {
	case RequestTypeForm:
		e = req.ParseForm()

	case RequestTypeQuery:
		e = req.ParseForm()

	case RequestTypeMultipartForm:
		e = req.ParseMultipartForm(0)

	case RequestTypeJSON:
		e = req.ParseForm()
		if e != nil {
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		reqBody, e = ioutil.ReadAll(req.Body)
	}
	if e != nil {
		log.Printf("handler.call: %d %s %s %s\n",
			http.StatusBadRequest, req.Method, req.URL.Path, e)
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	if debug.Value > 0 {
		log.Printf("> request body: %s\n", reqBody)
	}

	rspb, e := h.cb(req, reqBody)
	if e != nil {
		log.Printf("handler.call: %d %s %s %s\n",
			http.StatusInternalServerError,
			req.Method, req.URL.Path, e)
		h.error(res, e)
		return
	}

	switch h.resType {
	case ResponseTypeNone:
		res.WriteHeader(http.StatusNoContent)
		return
	case ResponseTypeBinary:
		res.Header().Set(contentType, contentTypeBinary)
	case ResponseTypeJSON:
		res.Header().Set(contentType, contentTypeJSON)
	case ResponseTypePlain:
		res.Header().Set(contentType, contentTypePlain)
	}

	res.WriteHeader(http.StatusOK)

	_, e = res.Write(rspb)
	if e != nil {
		log.Printf("handler.call: %s %s %s\n", req.Method, req.URL.Path, e)
	}
}

func (h *handler) error(res http.ResponseWriter, e error) {
	se, ok := e.(*StatusError)
	if !ok {
		se = &StatusError{
			Code:    http.StatusInternalServerError,
			Message: e.Error(),
		}
	} else {
		if se.Code == 0 {
			se.Code = http.StatusInternalServerError
		}
	}

	var rsp string

	switch h.resType {
	case ResponseTypeNone:
		res.WriteHeader(se.Code)
		return

	case ResponseTypeBinary:
		res.Header().Set(contentType, contentTypeBinary)
		rsp = se.Message

	case ResponseTypePlain:
		res.Header().Set(contentType, contentTypePlain)
		rsp = se.Message

	case ResponseTypeJSON:
		res.Header().Set(contentType, contentTypeJSON)
		rsp = fmt.Sprintf(`{"code":%d,"message":"%s"}`, se.Code,
			strings.JSONEscape(se.Message))
	}

	res.WriteHeader(se.Code)

	_, e = res.Write([]byte(rsp))
	if e != nil {
		log.Printf("handler.error: %s\n", e)
	}
}