package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

func main() {
	// tokenBucketRateLimiter()
	// fixedWindowRateLimiter()
	// dynamicWindowRateLimiter()
	adaptiveRateLimiter()
}

// 令牌桶模式，可以应对突发流量
func tokenBucketRateLimiter() {
	// 使用 rate.NewLimiter 创建了一个限制器，其速率限制为每秒产生10个令牌，最大桶容量为5，可以应对突发流量
	limiter := rate.NewLimiter(rate.Limit(10), 5)
	var wg sync.WaitGroup
	for i := 0; i < 10000; i++ {
		i := i
		wg.Add(1)
		go func() {
			// 每个请求都会调用Wait方法，该方法会阻塞直到有令牌可用，
			// 如果水桶是空的，没有可用的令牌，请求就会被拒绝。
			err := limiter.Wait(context.Background())
			if err != nil {
				log.Printf("rate limit exceeded,request rejected error:%v", err)
				return
			}

			defer wg.Done()
			// 获得令牌后处理请求
			log.Printf("wait limiter success:%d at %s", i, time.Now().Format("2006-01-02 15:04:05"))
			// 模拟耗时
			// time.Sleep(1 * time.Second)
		}()
	}

	// 等待协程执行完毕
	wg.Wait()
	log.Println("wait limiter success")
}

// 固定窗口模式
/*
2025/04/04 11:32:01 rate limit exceeded,request rejected
2025/04/04 11:32:01 rate limit exceeded,request rejected
2025/04/04 11:32:01 rate limit exceeded,request rejected
2025/04/04 11:32:01 rate limit exceeded,request rejected
2025/04/04 11:32:01 allow limiter success:8 at 2025-04-04 11:32:01
2025/04/04 11:32:01 allow limiter success:11 at 2025-04-04 11:32:01
2025/04/04 11:32:01 allow limiter success:27 at 2025-04-04 11:32:01
2025/04/04 11:32:01 allow limiter success:65 at 2025-04-04 11:32:01
*/
func fixedWindowRateLimiter() {
	// 创建一个每秒300个请求,桶的大小为100
	limiter := rate.NewLimiter(rate.Limit(300), 100)
	var wg sync.WaitGroup
	for i := 0; i < 200; i++ {
		// 对每个请求调用limiter.Allow()方法，如果允许请求，则返回true;如果超过速率限制，则返回false。
		// 如果超过速率限制，请求将被拒绝。
		if !limiter.Allow() {
			log.Println("rate limit exceeded,request rejected")
			continue
		}

		wg.Add(1)
		i := i
		go func() {
			defer wg.Done()

			log.Printf("allow limiter success:%d at %s", i, time.Now().Format("2006-01-02 15:04:05"))
			// 模拟耗时
			time.Sleep(500 * time.Millisecond)
		}()
	}

	wg.Wait()
	log.Println("wait limiter success")
}

// 动态限制速度
func dynamicWindowRateLimiter() {
	limiter := rate.NewLimiter(rate.Limit(10), 100) // 每秒10个请求，桶大小为100
	// 每2s调整一次频率
	go func() {
		timer := time.NewTimer(3 * time.Second)
		defer timer.Stop()

		for {
			select {
			case <-timer.C:
				log.Println("dynamic adjust limiter")
				// 设置limit大小和桶容量大小
				// limiter.SetLimit(rate.Limit(200)) // 将频率调整为每秒200个请求
				// limiter.SetBurst(200)              // 设置桶的大小

				// 动态调整
				// limit := float64(time.Now().Second() / 10 + 100) // 模拟动态调整limit
				limit := math.Ceil(float64(randInt(10, 60))/10 + 100) // 模拟动态调整limit
				newRate := rate.Limit(limit)
				log.Println("rate updated to:", newRate)
				log.Println("current limit:", limit)
				burst := int(newRate)
				log.Println("current burst:", burst)
				// 设置limit大小和桶容量大小
				limiter.SetLimit(newRate)
				limiter.SetBurst(burst)
			default:
			}
		}
	}()

	// 模拟请求
	for i := 0; i < 30*10000; i++ {
		// 如果超过速率限制，请求将被拒绝。
		if !limiter.Allow() {
			log.Println("rate limit exceeded,request rejected")
			continue
		}

		process(i)
	}
}

// 自适应速率限制
func adaptiveRateLimiter() {
	limiter := rate.NewLimiter(rate.Limit(100), 1) // 允许每秒100次

	// 自适应调整
	timer := time.NewTimer(3 * time.Second)
	go func() {
		for {
			select {
			case <-timer.C:
				responseTime := measureResponseTime()
				// 使用 limiter.SetLimit 设置不同的值来动态调整速率限制，
				// 这样，系统就能根据观察到的响应时间调整其速率限制策略。
				if responseTime > 500*time.Millisecond {
					fmt.Println("rate limit adjust to:50")
					limiter.SetLimit(rate.Limit(50))
				} else {
					fmt.Println("rate limit adjust to:100")
					limiter.SetLimit(rate.Limit(100))
				}
			default:
			}
		}
	}()

	for i := 0; i < 3000; i++ {
		if !limiter.Allow() {
			fmt.Println("rate limit exceeded. Request rejected.")
			time.Sleep(time.Millisecond * 100)
			continue
		}

		process(i)
	}
}

// 测量以前请求的响应时间
// 执行自己的逻辑来测量响应时间
func measureResponseTime() time.Duration {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	if r.Intn(3) < 2 {
		return time.Millisecond * 600
	}

	return time.Millisecond * 100
}

func process(i int) {
	log.Printf("allow limiter success:%d at %s", i, time.Now().Format("2006-01-02 15:04:05"))
	log.Println("request processed successfully.")
	// 模拟耗时
	time.Sleep(100 * time.Millisecond)
}

func randInt(min int, max int) int {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	if min >= max || min == 0 || max == 0 {
		return max
	}

	return r.Intn(max-min) + min
}
