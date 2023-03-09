package doraemon

import (
	"math/rand"
	"os"
	"time"
)

// GenerateRandomSeed 生成随机种子
func GenerateRandomSeed() int64 {
	// n^5 = maxInt64 , n = 6208
	var max = 6000

	var mod = int64(rand.Intn(max) + 1)
	//去掉后两位00
	now := time.Now().UnixNano() / 100
	factor1 := now%mod + 1
	// fmt.Println("now: ", now)
	// fmt.Println("mod: ", mod)
	// fmt.Println("factor1: ", factor1)

	mod = int64(rand.Intn(max) + 1)
	factor2 := int64(os.Getpid())%mod + 1

	now = time.Now().UnixNano() / 100
	mod = int64(rand.Intn(max) + 1)
	factor3 := now%mod + 1

	// Sleep防止太多的时间戳相同
	time.Sleep(time.Duration(factor3))

	now = time.Now().UnixNano() / 100
	mod = int64(rand.Intn(max) + 1)
	factor4 := now%mod + 1

	now = time.Now().UnixNano() / 100
	mod = int64(rand.Intn(max) + 1)
	factor5 := now%mod + 1

	return factor1 * factor2 * factor3 * factor4 * factor5
}
