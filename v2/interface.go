package v2

type WUID interface {
	Next() int64
}
