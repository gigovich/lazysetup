[![Go Report Card](https://goreportcard.com/badge/github.com/gigovich/lazysetup)](https://goreportcard.com/report/github.com/gigovich/lazysetup)
# Lazy setup
Golang lazy setup settings package. For most situations package constructors `func init()` are enought,
but sometimes you can't do some steps in constructors, also this package can be alternative
for dependency injection pattern.

## How to use
Just import package, and add init callbacks to default settings.

We define logger and it initialization in `log.go`:
```go
package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/gigovich/lazysetup"
)

// Log instance
var Log *logrus.Logger

func init() {
	lazysetup.OnInit(func() error {
		Log = logrus.New()
		Log.Level = logrus.DebugLevel
		return nil
	}, "InitLogger")
}
```

In `conf.go` we read config but use logger to log errors, so we depend on logger initialization:
```go
package main

import (
	"encoding/json"
	"github.com/gigovich/lazysetup"
	"os"
)

// Config settings
var Config = struct {
	Name string
	User string
	Pass string
	Host string
	Port string
}{}

func init() {
	lazysetup.OnInit(func() error {
		fh, err := os.Open("./config.json")
		if err != nil {
			Log.Debugf("open default config file: %v", err)
			return err
		}
		return json.NewDecoder(fh).Decode(&Config)
	}, "InitConfig", "InitLog") // we depend on initialized logger, so add them as last arg
}
```

Next one, database, depends on config and logger, also we need close connection to database, so we define close callback:
```go
package main

import (
	"database/sql"
	"fmt"

	"github.com/gigovich/lazysetup"
)

// DB instance
var DB *sql.DB

func init() {
	lazysetup.OnInit(func() (err error) {
		qs := fmt.Sprintf("user=%v dbname=%v password=%v", Config.User, Config.Name, Config.Pass)
		Log.Infof("connect to DB: %v", qs)
		DB, err = sql.Open("postgresql", qs)
		return err
	}, "InitDB", "InitConfig", "InitLog")

	lazysetup.OnClose(func() {
		if err := DB.Close(); err != nil {
			Log.Errorf("can't close DB: %v", err)
		}
	}, "CloseDB")
}
```

And at last, main function where we init everything:
```go
package main

import (
	"fmt"

	"github.com/gigovich/lazysetup"
)

func main() {
	if err := lazysetup.Init(); err != nil {
		fmt.Println("init setup", err)
		return
	}
	defer lazysetup.Close()

	if _, err := DB.Query("SELECT 1"); err != nil {
		Log.Errorf("exec query: %v", err)
	}
}
```

## Init order
The `OnInit` function is defined as:
`func OnInit(setupFunc func() error, name string, after ...string)`

First argument is callback function which will be called only when all other callbacks listed in `after` arguments will be executed.
Second argument is `name` of this initialization and can be used in `after` arguments list in other initializations.
