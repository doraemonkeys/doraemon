package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/cenkalti/backoff/v5"
)

// 1. 定义我们的结果类型 T
type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// 模拟一个不稳定的API
var requestCount = 0

func fetchUser(userID int) (User, error) {
	requestCount++
	fmt.Printf("Attempt %d: fetching user %d...\n", requestCount, userID)

	// 模拟前两次失败
	if requestCount <= 2 {
		return User{}, errors.New("simulated network flake")
	}

	// 模拟用户不存在 (永久性错误)
	if userID == 404 {
		return User{}, backoff.Permanent(fmt.Errorf("user with id %d not found", userID))
	}

	// 模拟成功
	fmt.Println("...success!")
	// 实际项目中这里会是 http.Get, json.Decode 等
	return User{ID: userID, Name: "John Doe"}, nil
}

func main() {
	// --- 场景1: 重试后成功 ---
	fmt.Println("--- Running Scenario 1: Success after retries ---")
	requestCount = 0 // 重置计数器

	// 定义操作，T 是 User
	operation := func() (User, error) {
		return fetchUser(123)
	}

	// 配置退避策略
	expBackoff := backoff.NewExponentialBackOff()
	expBackoff.InitialInterval = 100 * time.Millisecond

	// 使用泛型 Retry
	user, err := backoff.Retry(
		context.Background(),
		operation,
		backoff.WithBackOff(expBackoff),
	)

	if err != nil {
		log.Fatalf("Failed to fetch user: %v", err)
	}
	fmt.Printf("Successfully fetched user: %+v\n\n", user)

	// --- 场景2: 遇到永久性错误 ---
	fmt.Println("--- Running Scenario 2: Permanent error ---")
	requestCount = 0 // 重置计数器

	operationPermanent := func() (User, error) {
		fmt.Println("operationPermanent")
		return fetchUser(404) // 这个用户ID会触发永久性错误
	}

	// backoff.Retry 被设计为在遇到 backoff.Permanent 包装的错误时立即停止，因为它认为这种错误是不可恢复的，重试没有意义。
	user, err = backoff.Retry(context.Background(), operationPermanent, backoff.WithBackOff(backoff.NewExponentialBackOff()))
	if err != nil {
		fmt.Printf("Failed as expected with permanent error: %v requestCount: %d\n", err, requestCount)
		return
	}
	fmt.Println("This should not happen", user)

}

/*
指数退避（Exponential Backoff），并带有随机抖动（Jitter）

当你创建一个新的 `ExponentialBackOff` 实例时，它会使用以下默认值：

| 参数 | 默认值 | 解释 |
| :--- | :--- | :--- |
| **`InitialInterval`** | `500` 毫秒 | **初始等待时间**。第一次失败后，会等待大约半秒再重试。 |
| **`Multiplier`** | `1.5` | **乘数**。每次重试后，下一次的等待时间会是上一次的 1.5 倍。 |
| **`RandomizationFactor`** | `0.5` | **随机化因子**。实际等待时间会在计算出的间隔上下浮动最多 50%，以避免多个客户端同时重试。 |
| **`MaxInterval`** | `60` 秒 | **最长等待间隔**。即使按乘数计算出的等待时间超过了 1 分钟，实际等待也最多是 1 分钟。 |
| **`MaxElapsedTime`** | `15` 分钟 | **总耗时上限**。从第一次尝试开始，整个重试过程（包括所有等待）的总时长不会超过 15 分钟。这是防止无限重试的“安全阀”。 |

### 总结一句话

> 该库的默认策略是：**从 500ms 开始，每次重试的等待时间增加 50% 并带有随机抖动，最长单次等待不超过 1 分钟，总耗时不超过 15 分钟。**



*/

/*


### 1. 配置选项 (`RetryOption`)

`Retry` 函数通过 "Functional Options" 模式进行配置，这是一种灵活且可读性强的设计。你传递的 `opts ...RetryOption` 就是一系列函数，它们会修改 `Retry` 的内部配置。

虽然源码中没有直接展示，但根据 `args` 结构体，我们可以推断出主要的配置选项：

*   **`backoff.WithBackOff(b BackOff)`**: 这是**最重要的配置**。它告诉 `Retry` 使用哪一个退避策略算法。
    *   **默认值**: 如果不提供，它会使用一个默认的 `ExponentialBackOff`。
    *   **用法**: `backoff.WithBackOff(backoff.NewExponentialBackOff())`

*   **`backoff.WithMaxRetries(n uint)`**: 限制最大重试次数。
    *   **作用**: 即使总时间没有超时，只要重试次数达到了 `n` 次，就会停止。注意，这包括第一次尝试，所以 `WithMaxRetries(3)` 表示最多执行3次操作（1次初次尝试 + 2次重试）。
    *   **默认值**: 0 (表示不限制次数，只受时间限制)。

*   **`backoff.WithNotify(notifier Notify)`**: 提供一个在每次重试前调用的回调函数。
    *   **`Notify` 的类型**: `func(err error, d time.Duration)`。
    *   **作用**: 非常适合用于**日志记录**或**监控**。你可以在这里记录每次失败的原因（`err`）以及计划等待多长时间（`d`）再进行下一次尝试。
    *   **示例**: `backoff.WithNotify(func(err error, d time.Duration) { log.Printf("Operation failed: %v. Retrying in %s", err, d) })`

*   **`backoff.WithMaxElapsedTime(d time.Duration)`**: (这是一个v5新增的顶级配置) 设置一个总的最大耗时。
    *   **作用**: 覆盖 `BackOff` 策略本身可能带有的 `MaxElapsedTime`。它为整个 `Retry` 调用设定了一个硬性的时间上限。
    *   **默认值**: 15 分钟 (`backoff.DefaultMaxElapsedTime`)

---


```go
type ExponentialBackOff struct {
	InitialInterval     time.Duration // 初始间隔
	RandomizationFactor float64       // 随机化因子
	Multiplier          float64       // 乘数
	MaxInterval         time.Duration // 最大间隔(超过这个时间后，会使用最大间隔)

	currentInterval time.Duration     // 内部状态：当前间隔
}
```

它的 `NextBackOff()` 方法大致逻辑如下：

1.  **获取当前间隔**: 返回 `b.currentInterval` 作为本次的等待时间。
2.  **计算下一次间隔**: `b.currentInterval = b.currentInterval * b.Multiplier`。这就是“指数”增长的部分。
3.  **应用最大间隔限制**: `if b.currentInterval > b.MaxInterval { b.currentInterval = b.MaxInterval }`。等待时间不会无限增长，它有一个上限。
4.  **增加随机性 (Jitter)**:
    *   计算一个随机偏移量：`randomDelta = b.RandomizationFactor * b.currentInterval`。
    *   在当前等待时间上增加或减去一个随机值：`finalWait = b.currentInterval ± random(randomDelta)`。
    *   **为什么需要随机化？** 防止“惊群效应”（Thundering Herd）。如果多个客户端在同一时间失败，没有随机化，它们会在完全相同的时间点一起重试，可能会再次压垮刚刚恢复的服务。随机化可以将它们的重试时间点分散开。


*/
