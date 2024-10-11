package public

import (
	"gatewat_web/common/lib"
	"github.com/garyburd/redigo/redis"
)

// 该函数提供了一种方便的方式来执行多个 Redis 操作，使用了管道（pipeline）的概念。
// 通过传递多个函数，每个函数执行一个或多个 Redis 操作，可以在单个连接上批量执行这些操作。
// 通过使用管道，可以减少网络延迟，提高 Redis 操作的效率，因为这些操作会一次性发送到服务器。
func RedisConfPipline(pip ...func(c redis.Conn)) error {
	c, err := lib.RedisConnFactory("default")
	if err != nil {
		return err
	}
	defer c.Close()
	for _, f := range pip {
		f(c)
	}
	c.Flush()
	return nil
}

// 提供一些简化的接口，用于执行 Redis 操作
func RedisConfDo(commandName string, args string) (interface{}, error) {
	c, err := lib.RedisConnFactory("default")
	if err != nil {
		return nil, err
	}
	defer c.Close()
	return c.Do(commandName, args)
}
