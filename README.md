# settings - Imperative settings management for Go

## Key points

1. This module allows you to manage the application settings *imperative* way. Being not declarative is the
whole point of this project. Imperative code allows you to be more flexible and to have more control over the settings.
2. Performance is not a concern. Settings are usually loaded only once at the beginning of the application runtime.
That's why we can afford to do a lot of unnecessary work in exchange for more flexibility.

If you don't share my unhappiness with the declarative style of managing settings, you probably don't need this module.

## Features

- Read configuration from YAML, env variables, and command line arguments, and combine them.
- No reflection and no struct tags, so you can parse into a single variable, unexported field of a struct or even
a field of a struct declared in another Go module (which means, no more intermediate structures!), anything would work.
- Mark required settings.
- Set default values.

## Examples

### Direct usage on a struct from another module

```go
import (
	"github.com/redis/go-redis/v9"
	"github.com/tomr-ninja/go-settings"
)

func setupRedis() *redis.Client {
	redisOptions := new(redis.Options)
	redisConfig := settings.NewParser(settings.WithEnvPrefix("REDIS"))
	redisConfig.Add(&redisOptions.Addr).Env("ADDR").Default("localhost:6379")
	redisConfig.MustParse()	

	return redis.NewClient(redisOptions)
}
```

### Multiple settings sources

```go
import "github.com/tomr-ninja/go-settings"

func main() {
    var config struct {
        Addr string
        Port int
    }

	// priority decreases from the left to the right
	settings.Add(&config.Addr).Flag("addr").Env("ADDR").Default("localhost")
	settings.Add(&config.Port).Env("PORT").Flag("port").Default(8080)
	settings.MustParse()

    fmt.Printf("Addr: %s, Port: %d\n", config.Addr, config.Port)
}
```
