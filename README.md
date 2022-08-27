# Overview
- WUID is a globally unique number generator, while it is NOT a UUID implementation.
- WUID is **10-135** times faster than UUID and **4600** times faster than generating unique numbers with Redis.
- Each WUID instance generates unique 64-bit integers in sequence. The high 28 bits are loaded from a data store. By now, Redis, MySQL, MongoDB and Callback are supported.

# Benchmarks
```
BenchmarkWUID       187500764            6.38 ns/op        0 B/op          0 allocs/op
BenchmarkRand       97180698            12.2 ns/op         0 B/op          0 allocs/op
BenchmarkTimestamp  17126514            67.8 ns/op         0 B/op          0 allocs/op
BenchmarkUUID_V1    11986558            99.6 ns/op         0 B/op          0 allocs/op
BenchmarkUUID_V2    12017754           101 ns/op           0 B/op          0 allocs/op
BenchmarkUUID_V3     4925020           242 ns/op         144 B/op          4 allocs/op
BenchmarkUUID_V4    14184271            84.1 ns/op        16 B/op          1 allocs/op
BenchmarkUUID_V5     4277338           274 ns/op         176 B/op          4 allocs/op
BenchmarkRedis         35462         35646 ns/op         176 B/op          5 allocs/op
BenchmarkSnowflake   4931476           244 ns/op           0 B/op          0 allocs/op
BenchmarkULID        8410358           141 ns/op          16 B/op          1 allocs/op
BenchmarkXID        15000969            79.2 ns/op         0 B/op          0 allocs/op
BenchmarkShortID     1738039           698.9 ns/op       311 B/op         11 allocs/op
```

# Features
- Extremely fast
- Thread-safe
- Being unique across time
- Being unique within a data center
- Being unique globally if all data centers share a same data store, or they use different section IDs
- Being capable of generating 100M unique numbers in a single second with each WUID instance
- Auto-renew when the low 36 bits are about to run out

# Install
``` bash
go get -u github.com/r2b89/wuid
```

# Usage examples
### Redis
``` go
import "github.com/r2b89/wuid/redis/wuid"

newClient := func() (redis.Cmdable, bool, error) {
    var client redis.Cmdable
    // ...
    return client, true, nil
}

// Setup
g := NewWUID("default", nil)
_ = g.LoadH28FromRedis(newClient, "wuid")

// Generate
for i := 0; i < 10; i++ {
    fmt.Printf("%#016x\n", g.Next())
}
```

### MySQL
``` go
import "github.com/r2b89/wuid/mysql/wuid"

newDB := func() (*sql.DB, bool, error) {
    var db *sql.DB
    // ...
    return db, true, nil
}

// Setup
g := NewWUID("default", nil)
_ = g.LoadH28FromMysql(newDB, "wuid")

// Generate
for i := 0; i < 10; i++ {
    fmt.Printf("%#016x\n", g.Next())
}
```

### MongoDB
``` go
import "github.com/r2b89/wuid/mongo/wuid"

newClient := func() (*mongo.Client, bool, error) {
    var client *mongo.Client
    // ...
    return client, true, nil
}

// Setup
g := NewWUID("default", nil)
_ = g.LoadH28FromMongo(newClient, "test", "wuid", "default")

// Generate
for i := 0; i < 10; i++ {
    fmt.Printf("%#016x\n", g.Next())
}
```

### Callback
``` go
import "github.com/r2b89/wuid/callback/wuid"

// Setup
g := NewWUID("default", nil)
_ = g.LoadH28WithCallback(func() (int64, func(), error) {
    resp, err := http.Get("https://stackoverflow.com/")
    if resp != nil {
        defer func() {
            _ = resp.Body.Close()
        }()
    }
    if err != nil {
        return 0, nil, err
    }

    bytes, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return 0, nil, err
    }

    fmt.Printf("Page size: %d (%#06x)\n\n", len(bytes), len(bytes))
    return int64(len(bytes)), nil, nil
})

// Generate
for i := 0; i < 10; i++ {
    fmt.Printf("%#016x\n", g.Next())
}
```

# Mysql table creation
``` sql
CREATE TABLE IF NOT EXISTS `wuid` (
    `h` int(10) NOT NULL AUTO_INCREMENT,
    `x` tinyint(4) NOT NULL DEFAULT '0',
    PRIMARY KEY (`x`),
    UNIQUE KEY `h` (`h`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;
```

# Section ID
You can specify a custom section ID for the generated numbers with `wuid.WithSection` when you call `wuid.NewWUID`. The section ID must be in between `[0, 7]`.

# Step
You can customize the step value of `Next()` with `wuid.WithStep`.

# Special thanks
- [dustinfog](https://github.com/dustinfog)
