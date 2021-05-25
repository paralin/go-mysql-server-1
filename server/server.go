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

	"github.com/opentracing/opentracing-go"

	sqle "github.com/dolthub/go-mysql-server"
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
	var tracer opentracing.Tracer
	if cfg.Tracer != nil {
		tracer = cfg.Tracer
	} else {
		tracer = opentracing.NoopTracer{}
	}

	le := buildDefaultLogger(cfg.Logger)

	handler := NewHandler(
		le,
		e,
		NewSessionManager(
			sb,
			tracer,
			e.Analyzer.Catalog.HasDB,
			e.MemoryManager,
			e.ProcessList,
			cfg.Address),
		cfg.ConnReadTimeout,
		cfg.DisableClientMultiStatements,
		listener,
	)

	return &Server{h: handler}, nil
}

// SessionManager returns the session manager for this server.
func (s *Server) SessionManager() *SessionManager {
	return s.h.sm
}
