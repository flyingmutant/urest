package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"sync"

	_ "github.com/lib/pq"
	"github.com/sporttech/urest"
)

type (
	WithTxHandler struct {
		http.Handler
	}

	dbtx struct {
		db *sql.DB
		tx *sql.Tx
	}
)

const (
	driverName = "postgres"
	dbUrl      = "postgres://"
)

var (
	dbtxs      = map[*http.Request]*dbtx{}
	dbtxsMutex sync.Mutex
)

func (h WithTxHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open(driverName, dbUrl)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to establish database connection: %v", err), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	c := &dbtx{db, nil}

	dbtxsMutex.Lock()
	dbtxs[r] = c
	dbtxsMutex.Unlock()

	tw := &TransparentResponseWriter{w, 0, 0}
	h.Handler.ServeHTTP(tw, r)

	if c.tx != nil {
		if tw.Success() && !urest.IsSafeRequest(r) {
			c.tx.Commit()
		} else {
			c.tx.Rollback()
		}
	}

	dbtxsMutex.Lock()
	delete(dbtxs, r)
	dbtxsMutex.Unlock()
}

func Tx(r *http.Request) *sql.Tx {
	dbtxsMutex.Lock()
	c := dbtxs[r]
	dbtxsMutex.Unlock()

	if c.tx == nil {
		if tx, err := c.db.Begin(); err != nil {
			panic(err)
		} else {
			c.tx = tx
		}
	}
	return c.tx
}
