package redisc

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/garyburd/redigo/redis"
)

const (
	maxIdleConnections = 3
	idleTimeout        = 240 * time.Second
	cmdLpush           = "LPUSH"
	cmLpop             = "LPOP"
)

// Client provides executes commands on redis
type Client interface {
	//Push pushes to a redis list
	Push(listName string, payload []byte) error
	//Pop pops from a redis list
	Pop(listName string) (payload []byte, err error)
	//Close closes the connection
	Close() error
}

//RedisClient is a thread-safe client to redis.
//each method will get a connection from the pool
type RedisClient struct {
	URLPort string
	pool    *redis.Pool
}

// NewRedisClient creates a pool without checking connectivity
func NewRedisClient(urlPort string) (*RedisClient, error) {
	client := &RedisClient{
		URLPort: urlPort,
		pool:    newPool(urlPort),
	}
	client.cleanupHook()
	return client, nil
}

//Close closes the connection
func (client *RedisClient) Close() error {
	return client.pool.Close()
}

func newPool(urlPort string) *redis.Pool {
	return &redis.Pool{

		MaxIdle:     maxIdleConnections,
		IdleTimeout: idleTimeout,

		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", urlPort)
		},

		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

func (client *RedisClient) getConn() (redis.Conn, error) {
	conn := client.pool.Get()
	err := conn.Err()
	if err != nil {
		return nil, err
	}
	return conn, err
}

func (client *RedisClient) cleanupHook() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	signal.Notify(c, syscall.SIGINT)
	go func() {
		<-c
		client.Close()
	}()
}

//Push pushes to a redis list
func (client *RedisClient) Push(listName string, payload []byte) error {
	conn, err := client.getConn()
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Do(cmdLpush, listName, payload)
	return err
}

//Pop pops from a redis list
func (client *RedisClient) Pop(listName string) ([]byte, error) {
	conn, err := client.getConn()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	reply, err := conn.Do(cmLpop, listName)
	if err != nil {
		return nil, err
	}
	if reply == nil {
		return nil, fmt.Errorf("List is empty")
	}
	payload, isbytes := reply.([]byte)
	if !isbytes {
		return nil, fmt.Errorf("Received non []byte %v", payload)
	}
	return payload, err
}
