package data

import (
	"context"
	"fmt"
	"github.com/go-kratos/kratos-layout/internal/conf"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis/extra/redisotel"
	"github.com/go-redis/redis/v8"
)

func NewRedis(conf *conf.Data, l log.Logger) (rdb *redis.Client, cleanup func(), err error) {
	cleanup = func() {
		l.Log(log.LevelInfo, "closing the redis resources")
	}

	rdb = redis.NewClient(&redis.Options{
		Addr:         conf.Redis.Addr,
		Password:     conf.Redis.Password,
		DB:           int(conf.Redis.Db),
		DialTimeout:  conf.Redis.DialTimeout.AsDuration(),
		WriteTimeout: conf.Redis.WriteTimeout.AsDuration(),
		ReadTimeout:  conf.Redis.ReadTimeout.AsDuration(),
	})
	rdb.AddHook(redisotel.TracingHook{})
	return
}

func likeKey(id int64) string {
	return fmt.Sprintf("like:%d", id)
}

func (r *greeterRepo) GetArticleLike(ctx context.Context, id int64) (rv int64, err error) {
	get := r.data.rdb.Get(ctx, likeKey(id))
	rv, err = get.Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return
}

func (r *greeterRepo) IncArticleLike(ctx context.Context, id int64) error {
	_, err := r.data.rdb.Incr(ctx, likeKey(id)).Result()
	return err
}
