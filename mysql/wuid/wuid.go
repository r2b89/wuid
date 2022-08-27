package wuid

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/r2b89/wuid/v2"
)

// WUID is an extremely fast unique number generator.
type WUID struct {
	w *v2.WUID
}

// NewWUID creates a new WUID instance.
func NewWUID(name string, opts ...Option) *WUID {
	return &WUID{w: v2.NewWUID(name, opts...)}
}

// Next returns the next unique number.
func (this *WUID) Next() int64 {
	return this.w.Next()
}

type NewDB func() (client *sql.DB, autoDisconnect bool, err error)

// LoadH28FromMysql adds 1 to a specific number in your MySQL, fetches its new value, and then
// sets that as the high 28 bits of the unique numbers that Next generates.
func (this *WUID) LoadH28FromMysql(newDB NewDB, table string) error {
	if len(table) == 0 {
		return errors.New("table cannot be empty. name: " + this.w.Name)
	}

	db, autoDisconnect, err := newDB()
	if err != nil {
		return err
	}
	if autoDisconnect {
		defer func() {
			_ = db.Close()
		}()
	}

	result, err := db.Exec(fmt.Sprintf("REPLACE INTO %s (x) VALUES (0)", table))
	if err != nil {
		return err
	}
	h28, err := result.LastInsertId()
	if err != nil {
		return err
	}
	if err = this.w.VerifyH28(h28); err != nil {
		return err
	}

	this.w.Reset(h28 << 36)

	this.w.Lock()
	defer this.w.Unlock()

	if this.w.Renew != nil {
		return nil
	}
	this.w.Renew = func() error {
		return this.LoadH28FromMysql(newDB, table)
	}

	return nil
}

// RenewNow reacquires the high 28 bits from your data store immediately
func (this *WUID) RenewNow() error {
	return this.w.RenewNow()
}

type Option = v2.Option

// WithSection adds a section ID to the generated numbers. The section ID must be in between [0, 7].
func WithSection(section int8) Option {
	return v2.WithSection(section)
}

// WithH28Verifier sets your own h28 verifier
func WithH28Verifier(cb func(h28 int64) error) Option {
	return v2.WithH28Verifier(cb)
}

// WithStep sets the step and floor of Next()
func WithStep(step int64, floor int64) Option {
	return v2.WithStep(step, floor)
}
