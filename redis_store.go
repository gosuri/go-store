package store

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"code.google.com/p/go-uuid/uuid"
	"github.com/garyburd/redigo/redis"
)

const (
	MAX_ITEMS = 1000
)

var (
	// The environment variable to lookup redis connection url with
	// the format redis://:password@hostname:port/db_number
	DefaultRedisUrlEnv = "REDIS_URL"
)

// RedisItem represent the data structure
// used to store values in redis
type RedisItem struct {
	prefix string
	key    string
	data   map[string]interface{}
}

// Key returns the redis key used to
// store a redis item
func (i *RedisItem) Key() string {
	return i.prefix + ":" + i.key
}

// RedisConfig stores the configuration values used for
// establishing a connection with Redis server
type RedisConfig struct {
	Host, Port, User, Pass string
	Db                     int
}

// RedisStore represents the Store Implmention for Redis
type RedisStore struct {
	pool *redis.Pool
}

// NewRedisStore returns a RedisStore with
// default configuration values
func NewRedisStore() Store {
	return &RedisStore{pool: NewRedisPool(NewRedisConfig())}
}

// NewRedisConfig returns a default redis config. It uses the environment
// variable $REDIS_URL with the format redis://:password@hostname:port/db_number
// when present or defaults to 127.0.0.1:6379
func NewRedisConfig() *RedisConfig {
	// Use local Redis instance by default
	c := &RedisConfig{
		Host: "127.0.0.1",
		Port: "6379",
	}

	// Override default configure when env var $REDIS_URL
	// is present with the format redis://user:pass@host:port
	if url := os.Getenv(DefaultRedisUrlEnv); len(url) > 0 {
		url = strings.TrimPrefix(url, "redis://")
		parts := strings.Split(url, "@")

		// username and/or password exists
		if len(parts) == 2 {
			auth := strings.Split(parts[0], ":")
			c.User, c.Pass = auth[0], auth[1]
		}

		// the last part of the url is the address
		if addr := parts[len(parts)-1]; len(addr) > 0 {
			addrparts := strings.Split(addr, ":")
			c.Host, c.Port = addrparts[0], addrparts[1]
		}
	}
	return c
}

// NewRedisPool returns a default redis pool with default configuration
func NewRedisPool(config *RedisConfig) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", config.Host+":"+config.Port)
			if err != nil {
				return nil, err
			}
			if len(config.Pass) > 0 {
				if _, err := c.Do("AUTH", config.Pass); err != nil {
					c.Close()
					return nil, err
				}
			}
			if config.Db > 0 {
				if _, err := c.Do("SELECT", config.Db); err != nil {
					c.Close()
					return nil, err
				}
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

// Read reads the item from redis store and copies the values to item
// It Returns ErrKeyNotFound when no values are found for the key provided
// and ErrKeyMissing when key is not provided. Unmarshalling id done using
// driver provided redis.ScanStruct
func (s *RedisStore) Read(i Item) error {
	c := s.pool.Get()
	defer c.Close()

	value := reflect.ValueOf(i).Elem()
	if len(i.Key()) == 0 {
		return ErrEmptyKey
	}
	ri := &RedisItem{
		key:    i.Key(),
		prefix: value.Type().Name(),
	}
	reply, err := redis.Values(c.Do("HGETALL", ri.Key()))
	if err != nil {
		return err
	}
	if len(reply) == 0 {
		return ErrKeyNotFound
	}
	if err := redis.ScanStruct(reply, i); err != nil {
		return err
	}
	return nil
}

// Write writes the item to the store. It constructs the key using the i.Key()
// and prefixes it with the type of struct. When the key is empty, it assigns
// a unique universal id(UUID) using the SetKey method of the Item
func (s *RedisStore) Write(i Item) error {
	c := s.pool.Get()
	defer c.Close()

	value := reflect.ValueOf(i).Elem()

	ri := &RedisItem{
		prefix: s.typeName(value),
		data:   make(map[string]interface{}),
	}

	// Use the Items id if set or generate
	// a new UUID
	ri.key = i.Key()
	if len(ri.key) == 0 {
		ri.key = uuid.New()
	}
	i.SetKey(ri.key)

	// convert the item to redis item
	if err := marshall(i, value, ri); err != nil {
		return err
	}

	for key, val := range ri.data {
		if err := c.Send("HSET", ri.Key(), key, val); err != nil {
			return err
		}
		c.Flush()
	}

	return nil
}

// List populates the slice with ids of the slice element type
func (s *RedisStore) List(i interface{}) error {
	v := reflect.ValueOf(i)
	if v.Kind() != reflect.Ptr && v.Elem().Kind() != reflect.Slice {
		return errors.New("store: value must be a pointer to a slice")
	}

	c := s.pool.Get()
	defer c.Close()

	typeName := s.typeName(v.Elem())
	reply, err := redis.Values(c.Do("SCAN", "0", "MATCH", typeName+":*", "COUNT", MAX_ITEMS))
	if err != nil {
		return err
	}

	keys, _ := redis.Strings(reply[1], nil)
	ensureSliceLen(v.Elem(), len(keys))
	for index, key := range keys {
		// Remove the type of item from the key and just return the id
		id := strings.TrimPrefix(key, typeName+":")
		// value representing a pointer to a new zero value for the slice
		// element type. Basically, initialize a new item struct
		itemPtrV := reflect.New(v.Type().Elem().Elem())

		// function value corresponding to the SetKey function of the Struct
		setKeyFuncV := itemPtrV.MethodByName("SetKey")

		// array of values representing string ids to pass to the SetKey function
		setKeyFuncArgsV := []reflect.Value{reflect.ValueOf(id)}

		// call the SetKey function on the struct to store the key
		setKeyFuncV.Call(setKeyFuncArgsV)
		v.Elem().Index(index).Set(itemPtrV.Elem())
	}
	return nil
}

// Helper function to ensure the length of the slice in n
func ensureSliceLen(d reflect.Value, n int) {
	if n > d.Cap() {
		d.Set(reflect.MakeSlice(d.Type(), n, n))
	} else {
		d.SetLen(n)
	}
}

// Helper method to return the types name
func (scope *RedisStore) typeName(value reflect.Value) string {
	if value.Kind() == reflect.Slice {
		return value.Type().Elem().Name()
	}
	return value.Type().Name()
}

// Copies the Item to RedisItem and converts the
// struct field types to driver supported types
func marshall(item Item, value reflect.Value, rItem *RedisItem) error {
	// Ideally use the driver default marshalling redis.ConverAssignBytes
	for i := 0; i < value.NumField(); i++ {
		// key for data map
		k := value.Type().Field(i).Name
		field := value.Field(i)
		switch field.Kind() {
		case reflect.String:
			rItem.data[k] = field.String()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			rItem.data[k] = strconv.FormatInt(field.Int(), 10)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			rItem.data[k] = field.Uint()
		case reflect.Float32, reflect.Float64:
			rItem.data[k] = field.Float()
		case reflect.Bool:
			if field.Bool() {
				rItem.data[k] = "1"
			} else {
				rItem.data[k] = "0"
			}
		default:
			return errors.New(fmt.Sprintf("store: cannot convert %s (type: %s)", k, field.Kind()))
		}
	}
	return nil
}
