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
	"time"

	sqle "github.com/dolthub/go-mysql-server"
	"github.com/dolthub/go-mysql-server/server/golden"
	"github.com/dolthub/go-mysql-server/sql"
	"go.opentelemetry.io/otel/trace"
)

type ServerEventListener interface {
	ClientConnected()
	ClientDisconnected()
	QueryStarted()
	QueryCompleted(success bool, duration time.Duration)
}

// NewDefaultServer creates a Server with the default session builder.
func NewDefaultServer(cfg Config, e *sqle.Engine) (*Server, error) {
	return NewServer(cfg, e, DefaultSessionBuilder, nil)
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
	if cfg.ConnReadTimeout < 0 {
		cfg.ConnReadTimeout = 0
	}
	if cfg.ConnWriteTimeout < 0 {
		cfg.ConnWriteTimeout = 0
	}
	if cfg.MaxConnections < 0 {
		cfg.MaxConnections = 0
	}

	le := buildDefaultLogger(cfg.Logger)

	sessionMgr := NewSessionManager(
		sb,
		tracer,
		e.Analyzer.Catalog.Database,
		e.MemoryManager,
		e.ProcessList,
		cfg.Address,
	)
	handler := NewHandler(
		le,
		e,
		sessionMgr,
		cfg.ConnReadTimeout,
		cfg.DisableClientMultiStatements,
		cfg.MaxLoggedQueryLen,
		cfg.EncodeLoggedQuery,
		listener,
	)

	return &Server{handler: handler, Engine: e, sessionMgr: sessionMgr, le: le}, nil
}

// NewValidatingServer creates a Server that validates its query results using a MySQL connection
// as a source of golden-value query result sets.
func NewValidatingServer(
	cfg Config,
	e *sqle.Engine,
	sb SessionBuilder,
	listener ServerEventListener,
	mySqlConn string,
) (*Server, error) {
	server, err := NewServer(cfg, e, sb, listener)
	if err != nil {
		return nil, err
	}

	handler, err := golden.NewValidatingHandler(server.handler, mySqlConn, server.le)
	if err != nil {
		return nil, err
	}
	server.handler = handler

	return server, nil
}

func (s *Server) WarnIfLoadFileInsecure() {
	_, v, ok := sql.SystemVariables.GetGlobal("secure_file_priv")
	if ok {
		if v == "" {
			s.le.Warn("secure_file_priv is set to \"\", which is insecure. Any user with GRANT FILE privileges will be able to read any file which the sql-server process can read. Please consider restarting the server with secure_file_priv set to a safe (or non-existant) directory.")
		}
	}
}

// SessionManager returns the session manager for this server.
func (s *Server) SessionManager() *SessionManager {
	return s.sessionMgr
}
