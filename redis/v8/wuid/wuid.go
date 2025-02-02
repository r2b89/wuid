package wuid

import (
	"context"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
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

type NewClient func() (client redis.UniversalClient, autoDisconnect bool, err error)

// LoadH28FromRedis adds 1 to a specific number in your Redis, fetches its new value, and then
// sets that as the high 28 bits of the unique numbers that Next generates.
func (this *WUID) LoadH28FromRedis(newClient NewClient, key string) error {
	if len(key) == 0 {
		return errors.New("key cannot be empty. name: " + this.w.Name)
	}

	client, autoDisconnect, err := newClient()
	if err != nil {
		return err
	}
	if autoDisconnect {
		defer func() {
			_ = client.Close()
		}()
	}

	ctx1, cancel1 := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel1()
	h28, err := client.Incr(ctx1, key).Result()
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
		return this.LoadH28FromRedis(newClient, key)
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
