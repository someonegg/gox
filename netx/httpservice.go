// Copyright 2015 someonegg. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package netx

import (
	"errors"
	"github.com/someonegg/gox/syncx"
	"golang.org/x/net/context"
	"net"
	"net/http"
	"sync"
	"time"
)

var (
	ErrUnknownPanic = errors.New("unknown panic")
)

type ContextHandler interface {
	ContextServeHTTP(context.Context, http.ResponseWriter, *http.Request)
}

// HTTPService is a wrapper of http.Server.
type HTTPService struct {
	err     error
	quitCtx context.Context
	quitF   context.CancelFunc
	stopD   syncx.DoneChan

	l   *net.TCPListener
	h   ContextHandler
	srv *http.Server

	reqWG sync.WaitGroup
}

// NewHTTPService is a short cut to use NewHTTPServiceEx.
func NewHTTPService(l *net.TCPListener, h http.Handler) *HTTPService {

	return NewHTTPServiceEx(l, NewHTTPHandler(h))
}

func NewHTTPServiceEx(l *net.TCPListener, h ContextHandler) *HTTPService {
	s := &HTTPService{}

	s.quitCtx, s.quitF = context.WithCancel(context.Background())
	s.stopD = syncx.NewDoneChan()
	s.l = l
	s.h = h
	s.srv = &http.Server{
		Addr:           s.l.Addr().String(),
		Handler:        s,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	return s
}

func (s *HTTPService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.reqWG.Add(1)
	defer s.reqWG.Done()
	s.h.ContextServeHTTP(s.quitCtx, w, r)
}

func (s *HTTPService) Start() {
	go s.serve()
}

func (s *HTTPService) serve() {
	defer s.ending()

	s.err = s.srv.Serve(TcpKeepAliveListener{s.l})
}

func (s *HTTPService) ending() {
	if e := recover(); e != nil {
		switch v := e.(type) {
		case error:
			s.err = v
		default:
			s.err = ErrUnknownPanic
		}
	}

	s.quitF()
	s.stopD.SetDone()
}

func (s *HTTPService) Err() error {
	return s.err
}

func (s *HTTPService) Stop() {
	s.srv.SetKeepAlivesEnabled(false)
	s.quitF()
	s.l.Close()
}

func (s *HTTPService) StopD() syncx.DoneChanR {
	return s.stopD.R()
}

func (s *HTTPService) Stopped() bool {
	return s.stopD.R().Done()
}

func (s *HTTPService) WaitRequests() {
	s.reqWG.Wait()
}

func (s *HTTPService) QuitCtx() context.Context {
	return s.quitCtx
}

type httpHandler struct {
	oh http.Handler
}

// The http handler type is an adapter to allow the use of
// ordinary http.Handler as ContextHandler.
func NewHTTPHandler(oh http.Handler) ContextHandler {
	return httpHandler{oh}
}

func (h httpHandler) ContextServeHTTP(ctx context.Context,
	w http.ResponseWriter, r *http.Request) {

	h.oh.ServeHTTP(w, r)
}
