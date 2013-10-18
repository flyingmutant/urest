package handlers

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/sporttech/urest"
	"net/http"
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
	dbtxs = map[*http.Request]*dbtx{}
)

func (h WithTxHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open(driverName, dbUrl)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to establish database connection: %v", err), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	dbtxs[r] = &dbtx{db, nil}
	defer delete(dbtxs, r)

	defer func() {
		if tx := dbtxs[r].tx; tx != nil {
			tx.Rollback()
		}
	}()

	tw := &TransparentResponseWriter{w, 0, 0}
	h.Handler.ServeHTTP(tw, r)

	if tw.Success() && !urest.IsSafeRequest(r) {
		if tx := dbtxs[r].tx; tx != nil {
			tx.Commit()
		}
	}
}

func Tx(r *http.Request) *sql.Tx {
	c := dbtxs[r]
	if c.tx == nil {
		if tx, err := c.db.Begin(); err != nil {
			panic(err)
		} else {
			c.tx = tx
		}
	}
	return c.tx
}
