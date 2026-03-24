package assembly

import (
	"cnb.cool/mliev/dwz/dwz-server/pkg/interfaces"
	"cnb.cool/mliev/dwz/dwz-server/pkg/service/redis/config"
	"cnb.cool/mliev/dwz/dwz-server/pkg/service/redis/impl"
)

type Redis struct {
	Helper interfaces.HelperInterface
}

func (receiver *Redis) Assembly() error {
	redisConfig := config.NewRedis(receiver.Helper.GetConfig())
	redis, err := impl.NewRedis(receiver.Helper, redisConfig.Host, redisConfig.Port, redisConfig.DB, redisConfig.Password)
	if err != nil {
		return err
	}
	receiver.Helper.SetRedis(redis)
	return err
}
