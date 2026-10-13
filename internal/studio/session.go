// SPDX-License-Identifier: EUPL-1.2
package studio

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

type Options struct {
	Input, Hugo string
	Port        int
}

// Session owns only its root, source handles, listener and publishing children.
type Session struct {
	mu             sync.RWMutex
	state          State
	files          map[string][]byte
	retained       map[string]retained
	subscribers    map[uint64]chan struct{}
	nextSubscriber uint64
	selection      *Selection
	root, tool     string
	gate           *gate
	listener       net.Listener
	server         *http.Server
	jobs           chan revision
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
	closeOnce      sync.Once
	closeErr       error
	done           chan struct{}
	fatal          chan error
}

func Start(ctx context.Context, options Options) (session *Session, err error) {
	if options.Port < 0 || options.Port > 65535 {
		return nil, fmt.Errorf("port must be between 0 and 65535")
	}
	selected, err := OpenSelection(options.Input)
	if err != nil {
		return nil, err
	}
	root, err := os.MkdirTemp("", "sheetmill-studio-")
	if err != nil {
		selected.Close()
		return nil, err
	}
	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		selected.Close()
		os.RemoveAll(root)
		return nil, err
	}
	s := &Session{selection: selected, root: root, jobs: make(chan revision, 1), subscribers: map[uint64]chan struct{}{}, retained: map[string]retained{}, files: map[string][]byte{}, done: make(chan struct{}), fatal: make(chan error, 1)}
	s.ctx, s.cancel = context.WithCancel(ctx)
	defer func() {
		if err != nil {
			s.Close()
		}
	}()
	tool := options.Hugo
	if tool == "" {
		tool, err = exec.LookPath("hugo")
		if err != nil {
			return nil, fmt.Errorf("install Hugo 0.167.0 or select --hugo")
		}
	} else if !filepath.IsAbs(tool) {
		return nil, fmt.Errorf("--hugo requires an absolute installed executable")
	}
	s.tool, err = filepath.Abs(tool)
	if err != nil {
		return nil, err
	}
	if err = runTool(s.ctx, s.tool, []string{"version"}, root, true, maxGenerated); err != nil {
		return nil, err
	}
	s.listener, err = net.Listen("tcp4", fmt.Sprintf("127.0.0.1:%d", options.Port))
	if err != nil {
		return nil, fmt.Errorf("cannot bind the selected loopback port: %w", err)
	}
	s.gate, err = newGate(s.listener.Addr().String())
	if err != nil {
		return nil, err
	}
	initial := selected.Read()
	s.begin(1, initial)
	built := compile(s.ctx, s.tool, root, 1, initial, nil, 0)
	s.publish(1, built)
	if built.Error != "" {
		return nil, fmt.Errorf("cannot start preview: %s", built.Error)
	}
	s.server = &http.Server{Handler: s.gate.wrap(s.handler()), ReadHeaderTimeout: 3 * time.Second, IdleTimeout: 30 * time.Second, MaxHeaderBytes: 16 << 10}
	s.wg.Add(3)
	go func() { defer s.wg.Done(); s.watch(s.ctx, initial) }()
	go func() {
		defer s.wg.Done()
		for {
			select {
			case <-s.ctx.Done():
				return
			case r := <-s.jobs:
				s.mu.RLock()
				previous := make(map[string]retained, len(s.retained))
				for k, v := range s.retained {
					previous[k] = v
				}
				bytes := 0
				for _, entry := range previous {
					data, _ := entry.Document.JSON()
					bytes += len(data)
				}
				for _, v := range s.files {
					bytes += len(v)
				}
				s.mu.RUnlock()
				built := compile(s.ctx, s.tool, s.root, r.seq, r.snapshot, previous, bytes)
				s.publish(r.seq, built)
			}
		}
	}()
	go func() {
		defer s.wg.Done()
		err := s.server.Serve(s.listener)
		if err != nil && err != http.ErrServerClosed {
			select {
			case s.fatal <- fmt.Errorf("preview listener failed"):
			default:
			}
			s.cancel()
		}
	}()
	return s, nil
}
func (s *Session) BootstrapURL() string {
	return s.gate.origin + "/studio/bootstrap?token=" + s.gate.secret
}
func (s *Session) Scope() string { return s.selection.Scope }
func (s *Session) Wait() error {
	select {
	case err := <-s.fatal:
		return err
	case <-s.ctx.Done():
		select {
		case err := <-s.fatal:
			return err
		default:
			return nil
		}
	}
}
func (s *Session) Close() error {
	s.closeOnce.Do(func() {
		if s.cancel != nil {
			s.cancel()
		}
		if s.server != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			s.closeErr = s.server.Shutdown(ctx)
			cancel()
			if s.closeErr != nil {
				s.server.Close()
			}
		} else if s.listener != nil {
			s.listener.Close()
		}
		s.wg.Wait()
		if s.selection != nil {
			s.selection.Close()
		}
		if s.root != "" {
			if err := os.RemoveAll(s.root); s.closeErr == nil {
				s.closeErr = err
			}
		}
		if s.done != nil {
			close(s.done)
		}
	})
	return s.closeErr
}
