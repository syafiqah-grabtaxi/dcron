package driver_test

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/require"
	"github.com/syafiqah-mr/dcron/dlog"
	"github.com/syafiqah-mr/dcron/driver"
)

func testFuncNewRedisZSetDriver(addr string) driver.DriverV2 {
	redisCli := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: []string{addr},
	})
	return driver.NewRedisZSetDriver(redisCli)
}

func TestRedisZSetDriver_GetNodes(t *testing.T) {
	rds := miniredis.RunT(t)
	drvs := make([]driver.DriverV2, 0)
	N := 10
	for i := 0; i < N; i++ {
		drv := testFuncNewRedisZSetDriver(rds.Addr())
		drv.Init(
			t.Name(),
			driver.NewTimeoutOption(5*time.Second),
			driver.NewLoggerOption(dlog.NewLoggerForTest(t)))
		err := drv.Start(context.Background())
		require.Nil(t, err)
		drvs = append(drvs, drv)
	}

	for _, v := range drvs {
		nodes, err := v.GetNodes(context.Background())
		require.Nil(t, err)
		require.Equal(t, N, len(nodes))
	}

	for _, v := range drvs {
		v.Stop(context.Background())
	}
}

func TestRedisZSetDriver_Stop(t *testing.T) {
	var err error
	var nodes []string
	rds := miniredis.RunT(t)
	drv1 := testFuncNewRedisZSetDriver(rds.Addr())
	drv1.Init(t.Name(),
		driver.NewTimeoutOption(5*time.Second),
		driver.NewLoggerOption(dlog.NewLoggerForTest(t)))

	drv2 := testFuncNewRedisZSetDriver(rds.Addr())
	drv2.Init(t.Name(),
		driver.NewTimeoutOption(5*time.Second),
		driver.NewLoggerOption(dlog.NewLoggerForTest(t)))
	err = drv2.Start(context.Background())
	require.Nil(t, err)

	err = drv1.Start(context.Background())
	require.Nil(t, err)

	nodes, err = drv1.GetNodes(context.Background())
	require.Nil(t, err)
	require.Len(t, nodes, 2)

	nodes, err = drv2.GetNodes(context.Background())
	require.Nil(t, err)
	require.Len(t, nodes, 2)

	drv1.Stop(context.Background())

	<-time.After(6 * time.Second)
	nodes, err = drv2.GetNodes(context.Background())
	require.Nil(t, err)
	require.Len(t, nodes, 1)

	err = drv1.Start(context.Background())
	require.Nil(t, err)
	<-time.After(5 * time.Second)
	nodes, err = drv2.GetNodes(context.Background())
	require.Nil(t, err)
	require.Len(t, nodes, 2)

	drv2.Stop(context.Background())
}
