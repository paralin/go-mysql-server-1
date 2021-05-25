// Copyright 2020-2021 Dolthub, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"net"
	"time"

	"github.com/dolthub/vitess/go/mysql"
	"go.opentelemetry.io/otel/trace"

	sqle "github.com/dolthub/go-mysql-server"
	"github.com/dolthub/go-mysql-server/sql"
)

// ProtocolListener handles connections based on the configuration it was given. These listeners also implement
// their own protocol, which by default will be the MySQL wire protocol, but another protocol may be provided.
type ProtocolListener interface {
	Addr() net.Addr
	Accept()
	Close()
}

type ServerEventListener interface {
	ClientConnected()
	ClientDisconnected()
	QueryStarted()
	QueryCompleted(success bool, duration time.Duration)
}

// NewServer creates a server with the given protocol, address, authentication
// details given a SQLe engine and a session builder.
func NewServer(cfg Config, e *sqle.Engine, sb SessionBuilder, listener ServerEventListener) (*Server, error) {
	var tracer trace.Tracer
	if cfg.Tracer != nil {
		tracer = cfg.Tracer
	} else {
		tracer = sql.NoopTracer
	}

	sm := NewSessionManager(sb, tracer, e.Analyzer.Catalog.Database, e.MemoryManager, e.ProcessList, cfg.Address)
	handler := &Handler{
		e:                 e,
		sm:                sm,
		readTimeout:       cfg.ConnReadTimeout,
		disableMultiStmts: cfg.DisableClientMultiStatements,
		maxLoggedQueryLen: cfg.MaxLoggedQueryLen,
		encodeLoggedQuery: cfg.EncodeLoggedQuery,
		sel:               listener,
	}
	//handler = NewHandler_(e, sm, cfg.ConnReadTimeout, cfg.DisableClientMultiStatements, cfg.MaxLoggedQueryLen, cfg.EncodeLoggedQuery, listener)
	return newServerFromHandler(cfg, e, sm, handler)
}

// HandlerWrapper provides a way for clients to wrap the mysql.Handler used by the server with a custom implementation
// that wraps it.
type HandlerWrapper func(h mysql.Handler) (mysql.Handler, error)

// NewServerWithHandler creates a Server with a handler wrapped by the provided wrapper function.
func NewServerWithHandler(
	cfg Config,
	e *sqle.Engine,
	sb SessionBuilder,
	listener ServerEventListener,
	wrapper HandlerWrapper,
) (*Server, error) {
	var tracer trace.Tracer
	if cfg.Tracer != nil {
		tracer = cfg.Tracer
	} else {
		tracer = sql.NoopTracer
	}

	sm := NewSessionManager(sb, tracer, e.Analyzer.Catalog.Database, e.MemoryManager, e.ProcessList, cfg.Address)
	h := &Handler{
		e:                 e,
		sm:                sm,
		readTimeout:       cfg.ConnReadTimeout,
		disableMultiStmts: cfg.DisableClientMultiStatements,
		maxLoggedQueryLen: cfg.MaxLoggedQueryLen,
		encodeLoggedQuery: cfg.EncodeLoggedQuery,
		sel:               listener,
	}

	handler, err := wrapper(h)
	if err != nil {
		return nil, err
	}

	return newServerFromHandler(cfg, e, sm, handler)
}

func portInUse(hostPort string) bool {
	timeout := time.Second
	conn, _ := net.DialTimeout("tcp", hostPort, timeout)
	if conn != nil {
		defer conn.Close()
		return true
	}
	return false
}

func newServerFromHandler(cfg Config, e *sqle.Engine, sm *SessionManager, handler mysql.Handler) (*Server, error) {
	for _, option := range cfg.Options {
		option(e, sm, handler)
	}

	if cfg.ConnReadTimeout < 0 {
		cfg.ConnReadTimeout = 0
	}
	if cfg.ConnWriteTimeout < 0 {
		cfg.ConnWriteTimeout = 0
	}
	if cfg.MaxConnections < 0 {
		cfg.MaxConnections = 0
	}

	return &Server{
		handler:    handler,
		sessionMgr: sm,
		Engine:     e,
	}, nil
}

// SessionManager returns the session manager for this server.
func (s *Server) SessionManager() *SessionManager {
	return s.sessionMgr
}
