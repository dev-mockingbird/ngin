package redis

import (
	"context"
	"time"

	"github.com/dev-mockingbird/logf"
	"github.com/dev-mockingbird/ngin"
	"github.com/redis/go-redis/v9"
)

var redis_cli *redis.Client

func Init(ctx *ngin.Context) {
	ctx.BindFunc("config-redis", ConfigRedis)
	ctx.BindFunc("redis-set", RedisSet)
	ctx.BindValuedFunc("redis-get", RedisGet)
}

func ConfigRedis(ctx *ngin.Context, args ...ngin.Value) (bool, error) {
	opt := &redis.Options{}
	if len(args) >= 1 {
		opt.Addr = args[0].String()
	}
	if len(args) >= 2 {
		opt.DB = int(args[1].Int())
	}
	if len(args) >= 3 {
		opt.Username = args[2].String()
	}
	if len(args) >= 4 {
		opt.Password = args[3].String()
	}
	redis_cli = redis.NewClient(opt)
	return true, nil
}

func RedisSet(ctx *ngin.Context, args ...ngin.Value) (bool, error) {
	if redis_cli == nil {
		ctx.Logger().Logf(logf.Error, "you should call redis_cfg before use it")
		return false, nil
	}
	if len(args) < 2 {
		ctx.Logger().Logf(logf.Error, "you should provide at least 2 argments for redis_set")
		return false, nil
	}
	key := args[0].String()
	val := args[1].String()
	expire := time.Duration(0)
	if len(args) >= 3 {
		expire = time.Second * time.Duration(args[2].Int())
	}
	err := redis_cli.Set(context.Background(), key, val, expire).Err()
	if err != nil {
		ctx.Logger().Logf(logf.Error, "redis_set: %s", err.Error())
		return false, nil
	}
	return true, nil
}

func RedisGet(ctx *ngin.Context, args ...ngin.Value) ngin.Value {
	if redis_cli == nil {
		ctx.Logger().Logf(logf.Error, "you should call redis_cfg before use it")
		return ngin.Null{}
	}
	if len(args) == 0 {
		ctx.Logger().Logf(logf.Error, "you should provide the key which you wanna fetch")
		return ngin.Null{}
	}
	res, err := redis_cli.Get(context.Background(), args[0].String()).Result()
	if err != nil {
		ctx.Logger().Logf(logf.Error, "redis get: %s", err.Error())
		return ngin.Null{}
	}
	return ngin.String(res)
}
