package core

import (
	"net/http"

	"github.com/gorilla/sessions"
	"golang.org/x/xerrors"
)

type SessionRouter struct {
	store sessions.Store
}

func NewSessionRouter(
	store sessions.Store,
) *SessionRouter {
	return &SessionRouter{
		store: store,
	}
}

func (s *SessionRouter) Get(r *http.Request, key string) (*Routes, error) {
	session, err := s.store.Get(r, key)
	if err != nil {
		return nil, xerrors.Errorf("failed to get session: %w", err)
	}

	iv := session.Values["identifier"]
	hv := session.Values["host"]
	if iv != nil && hv != nil {
		return &Routes{
			Identifier: iv.(string),
			Host:       hv.(string),
		}, nil
	}
	return nil, nil
}

func (s *SessionRouter) Save(r *http.Request, w http.ResponseWriter, key string, info *Routes) error {
	session, err := s.store.Get(r, key)
	if err != nil {
		return xerrors.Errorf("failed to get session: %w", err)
	}
	session.Values["identifier"] = info.Identifier
	session.Values["host"] = info.Host
	if err := session.Save(r, w); err != nil {
		return xerrors.Errorf("failed to save session: %w", err)
	}
	return nil
}
