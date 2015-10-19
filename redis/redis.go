// Copyright 2015 Greg Osuri. All rights reserved.
// Use of this source code is governed by the Apache License, Version 2.0
// that can be found in the LICENSE file.

package redis

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gosuri/go-store/_vendor/src/code.google.com/p/go-uuid/uuid"
	driver "github.com/gosuri/go-store/_vendor/src/github.com/garyburd/redigo/redis"
	"github.com/gosuri/go-store/store"
)

const (
	// MaxItems specifies the max number of items to fetch form redis on each call
	MaxItems = 1024
)

var (
	// DefaultRedisURLEnv specifies the name of environment variable
	// that contains the Redis connection url. It expects the format
	// redis://:password@hostname:port/db_number
	DefaultRedisURLEnv = "REDIS_URL"

	// Default connection url to connect to redis
	DefaultRedisUrl = "redis://@127.0.0.1:6379"
)

// item represent the data structure used to store values in redis.
type item struct {
	prefix string
	key    string
	data   map[string]interface{}
}

// Key returns the redis key used to store a redis item by prefix the item type.
func (i *item) Key() string {
	return i.prefix + ":" + i.key
}

// Config stores the configuration values used for establishing a
// connection with Redis server.
type Config struct {
	Host, Port, Pass string
	Db               int
	// Namespace for redis
	Namespace string
}

// redis implements represents the Store methods implemention for Redis.
type Redis struct {
	pool      *driver.Pool
	namespace string
}

func New(config *Config) (r *Redis, err error) {
	if config == nil {
		config, err = NewConfig(DefaultRedisUrl)
		if err != nil {
			return nil, err
		}
	}
	return &Redis{pool: NewPool(config), namespace: config.Namespace}, nil
}

// NewStore returns an instance of Store. It parses the connection information from the connUrl provided
// and expects the format redis://:password@hostname:port/db_number. If connUrl is empty it reads from
// the environment variable DefaultRedisURLEnv or defaults to redis://127.0.0.0:6837
func NewStore(connUrl, namespace string) (store.Store, error) {
	config, err := NewConfig(connUrl)
	if err != nil {
		return &Redis{}, err
	}
	return &Redis{pool: NewPool(config), namespace: namespace}, nil
}

// NewConfig returns a default redis config. It parses the connection information from the connUrl provided
// and expects the format redis://:password@hostname:port/db_number. If connUrl is empty it reads from
// the environment variable DefaultRedisURLEnv or defaults to redis://127.0.0.0:6837
func NewConfig(connUrl string) (*Config, error) {
	// read from DefaultRedisURLEnv var if url is missing
	config := &Config{}
	if len(connUrl) == 0 {
		connUrl = os.Getenv(DefaultRedisURLEnv)
	}

	// default to redis://@127.0.0.0:6837 if no environment variable is present
	if len(connUrl) == 0 {
		connUrl = DefaultRedisUrl
	}

	rUrl, err := url.Parse(connUrl)
	if err != nil {
		return config, err
	}

	// Parse redis host and port
	rHost := strings.Split(rUrl.Host, ":")
	config.Host = rHost[0]
	config.Port = rHost[1]

	// Set password if exits

	rPass, passOk := rUrl.User.Password()
	if passOk {
		config.Pass = rPass
	}

	// Set redis db number if exists
	rDb := strings.TrimPrefix(rUrl.Path, "/")
	if len(rDb) > 0 {
		if config.Db, err = strconv.Atoi(rDb); err != nil {
			return config, err
		}
	}
	return config, nil
}

// NewPool returns a default redis pool with default configuration.
func NewPool(config *Config) *driver.Pool {
	return &driver.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (driver.Conn, error) {
			c, err := driver.Dial("tcp", config.Host+":"+config.Port)
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
		TestOnBorrow: func(c driver.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

// Pool returns a redis pool in use with the store.
// It returns a new pool otherwise
func (r *Redis) Pool() *driver.Pool {
	if r.pool == nil {
		r.pool = NewPool(nil)
	}
	return r.pool
}

// Read reads the item from redis store and copies the values to item
// It Returns store.ErrKeyNotFound when no values are found for the key provided
// and store.ErrKeyMissing when key is not provided. Unmarshalling id done using
// driver provided redis.ScanStruct
func (s *Redis) Read(i store.Item) error {
	c := s.pool.Get()
	defer c.Close()

	value := reflect.ValueOf(i).Elem()
	if len(i.Key()) == 0 {
		return store.ErrEmptyKey
	}
	ri := &item{
		key:    i.Key(),
		prefix: s.nameInNamespace(value.Type().Name()),
	}
	reply, err := driver.Values(c.Do("HGETALL", ri.Key()))
	if err != nil {
		return err
	}
	if len(reply) == 0 {
		return store.ErrKeyNotFound
	}
	if err := driver.ScanStruct(reply, i); err != nil {
		return err
	}
	return nil
}

// ReadMultiple gets the values from redis in a single call by pipelining
func (s *Redis) ReadMultiple(i interface{}) error {
	v := reflect.ValueOf(i)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Slice {
		return errors.New("store: value must be a a slice")
	}

	c := s.pool.Get()
	defer c.Close()

	var key string
	var err error
	prefix := s.typeName(v) + ":"

	// Using transactions to execute HGETALL in a pipeline.
	// Mark the start of a transaction block.
	// Subsequent commands will be queued for atomic execution.
	c.Send("MULTI")
	for y := 0; y < v.Len(); y++ {
		if key = v.Index(y).Addr().MethodByName("Key").Call(nil)[0].String(); len(key) == 0 {
			return store.ErrEmptyKey
		}
		// Send writes the command to the connection's output buffer.
		if err = c.Send("HGETALL", prefix+key); err != nil {
			return err
		}
	}
	// Flush flushes the connection's output buffer to the server
	if err = c.Flush(); err != nil {
		return err
	}
	// Execute all previously queued commands
	// in a transaction and restores the connection
	// state to normal
	reply, err := c.Do("EXEC")
	if err != nil {
		return err
	}

	replyValue := reflect.ValueOf(reply)
	var values []interface{}
	// Reply is a two dimentional array of interfaces. Iterate over the first
	// dimension and scan each slice into destination interface type
	for y := 0; y < replyValue.Len(); y++ {
		itemPtrV := reflect.New(v.Type().Elem())
		if values, err = driver.Values(replyValue.Index(y).Interface(), nil); err != nil {
			return err
		}
		driver.ScanStruct(values, itemPtrV.Interface())
		v.Index(y).Set(itemPtrV.Elem())
	}
	return nil
}

// WriteMultiple writes multiple items i to the store.
func (s *Redis) WriteMultiple(i []store.Item) error {
	return errors.New("Implementation pending")
}

// Write writes the item to the store. It constructs the key using the i.Key()
// and prefixes it with the type of struct. When the key is empty, it assigns
// a unique universal id(UUID) using the SetKey method of the Item
func (s *Redis) Write(i store.Item) error {
	c := s.pool.Get()
	defer c.Close()

	value := reflect.ValueOf(i).Elem()

	ri := &item{
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
	if err := marshall(value, ri); err != nil {
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

// DeleteMultiple deletes multiple items i from the store. It returns the count
// of items successfully deleted. It returns an error if any of the items do
// not exist or can't be deleted. It will delete the other items, in that case.
func (s *Redis) DeleteMultiple(items []store.Item) (int, error) {
	c := s.pool.Get()
	defer c.Close()

	keys := make([]interface{}, len(items))
	for i, item := range items {
		value := reflect.ValueOf(item).Elem()
		if len(item.Key()) > 0 {
			keys[i] = fmt.Sprintf("%s:%s", s.typeName(value), item.Key())
		}
	}

	count, err := driver.Int(c.Do("DEL", keys...))
	if err != nil {
		return 0, err
	}
	if count != len(keys) {
		return count, store.ErrKeyNotFound
	}

	return count, nil
}

// Delete deletes the item from the store. It constructs the key using i.Key().
// When the key is empty, it returns a store.ErrEmptyKey error. When the key
// does not exist, it returns a store.ErrKeyNotFound error.
func (s *Redis) Delete(i store.Item) error {
	c := s.pool.Get()
	defer c.Close()

	value := reflect.ValueOf(i).Elem()

	ri := &item{
		prefix: s.typeName(value),
	}

	ri.key = i.Key()
	if len(ri.key) == 0 {
		return store.ErrEmptyKey
	}

	count, err := driver.Int(c.Do("DEL", ri.Key()))
	if err != nil {
		return err
	}
	if count == 0 {
		return store.ErrKeyNotFound
	}

	return nil
}

// List populates the slice with ids of the slice element type.
func (s *Redis) List(i interface{}) error {
	v := reflect.ValueOf(i)
	// Get the elements of the interface if its a pointer
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Slice {
		return errors.New("store: value must be a a slice")
	}

	c := s.pool.Get()
	defer c.Close()

	typeName := s.typeName(v)
	var cursor int64
	var keys []string

	// Ideally, want to fetch in a go routine
	for cursor >= 0 {
		// SCAN return value is an array of two values: the first value
		// is the new cursor to use in the next call, the second value
		// is an array of elements.
		reply, err := c.Do("SCAN", cursor, "MATCH", typeName+":*", "COUNT", MaxItems)
		if err != nil {
			return err
		}
		// Read the cursor bits, the driver provides them as
		// an array of unsigned 8-bit integers
		cursorBytes := reflect.ValueOf(reply).Index(0).Interface().([]uint8)

		// Converting the []uint8 to int by converting to a string first, there
		// is perhaps an optimal way but I could not figure out in go's constructs
		if cursor, err = strconv.ParseInt(fmt.Sprintf("%s", cursorBytes), 10, 64); err != nil {
			return err
		}
		valueBytes := reflect.ValueOf(reply).Index(1).Interface().([]interface{})
		values, _ := driver.Strings(valueBytes, nil)
		keys = append(keys, values...)
		// Break the loop when the no more records left to read (cursor is 0)
		if cursor == 0 {
			break
		}
	}

	// Format and copy the keys to interface and ensure the interface
	// has the required length.
	ensureSliceLen(v, len(keys))
	for index, key := range keys {
		// Remove the type of item from the key and just return the id
		id := strings.TrimPrefix(key, typeName+":")
		// value representing a pointer to a new zero value for the slice
		// element type. Basically, initialize a new item struct
		itemPtrV := reflect.New(v.Type().Elem())

		// function value corresponding to the SetKey function of the Struct
		setKeyFuncV := itemPtrV.MethodByName("SetKey")

		// array of values representing string ids to pass to the SetKey function
		setKeyFuncArgsV := []reflect.Value{reflect.ValueOf(id)}

		// call the SetKey function on the struct to store the key
		setKeyFuncV.Call(setKeyFuncArgsV)
		v.Index(index).Set(itemPtrV.Elem())
	}
	return nil
}

// ensureSliceLen is a helper function to ensure the length of the slice is n
func ensureSliceLen(d reflect.Value, n int) {
	if n > d.Cap() {
		d.Set(reflect.MakeSlice(d.Type(), n, n))
	} else {
		d.SetLen(n)
	}
}

// typeName is a helper function to return the name of the type.
func (s *Redis) typeName(value reflect.Value) string {
	if value.Kind() == reflect.Slice {
		return s.nameInNamespace(value.Type().Elem().Name())
	}
	return s.nameInNamespace(value.Type().Name())
}

// marshall is a helper function that copies the value to redis.item and converts
// the struct field types to driver supported types
func marshall(value reflect.Value, rItem *item) error {
	// Ideally use the driver default marshalling redis.ConverAssignBytes
	for i := 0; i < value.NumField(); i++ {
		// key for data map
		k := value.Type().Field(i).Name
		field := value.Field(i)
		// ignore unexported fields
		if !field.CanSet() {
			continue
		}
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
			return fmt.Errorf("store: cannot convert %s (type: %s)", k, field.Kind())
		}
	}
	return nil
}

// nameInNamespace returns the item names with namespace prefixed
func (s *Redis) nameInNamespace(name string) string {
	if len(s.namespace) != 0 {
		return s.namespace + ":" + name
	}
	return name
}
