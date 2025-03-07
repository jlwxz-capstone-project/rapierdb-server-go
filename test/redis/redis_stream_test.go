package main

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestRedisStream(t *testing.T) {
	ctx := context.Background()

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer rdb.Close()

	// 清理测试前可能存在的流数据
	streamKey := "test-stream"
	rdb.Del(ctx, streamKey)

	// 配置参数
	numGroups := 2                  // 消费者组数量
	numConsumersPerGroup := 2       // 每个组的消费者数量
	numPublishers := 2              // 生产者数量
	testDuration := 5 * time.Second // 测试持续时间

	// 创建消费者组
	for i := 1; i <= numGroups; i++ {
		groupName := fmt.Sprintf("group%d", i)

		// 创建消费者组，使用MKSTREAM选项自动创建流
		err := rdb.XGroupCreateMkStream(ctx, streamKey, groupName, "0").Err()
		if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
			t.Fatalf("创建消费者组%d失败: %v", i, err)
		}
	}

	// 用于等待所有goroutine完成
	var wg sync.WaitGroup

	// 创建停止信号通道
	stopCh := make(chan struct{})

	// 启动生产者
	wg.Add(numPublishers)
	for i := 1; i <= numPublishers; i++ {
		publisherID := fmt.Sprintf("publisher%d", i)
		go func(id string) {
			defer wg.Done()
			publishMessages(ctx, rdb, streamKey, id, stopCh, t)
		}(publisherID)
	}

	// 启动消费者组和消费者
	for i := 1; i <= numGroups; i++ {
		groupName := fmt.Sprintf("group%d", i)

		// 为每个组启动指定数量的消费者
		wg.Add(numConsumersPerGroup)
		for j := 1; j <= numConsumersPerGroup; j++ {
			consumerID := fmt.Sprintf("consumer%d-%d", i, j)
			go func(group, consumer string) {
				defer wg.Done()
				consumeMessages(ctx, rdb, streamKey, group, consumer, stopCh, t)
			}(groupName, consumerID)
		}
	}

	// 运行指定时间后停止
	time.Sleep(testDuration)
	close(stopCh)

	// 等待所有goroutine完成
	wg.Wait()

	// 清理测试数据
	rdb.Del(ctx, streamKey)
}

// 发布消息到Redis Stream
func publishMessages(ctx context.Context, rdb *redis.Client, streamKey, publisherID string, stopCh <-chan struct{}, t *testing.T) {
	counter := 0
	for {
		select {
		case <-stopCh:
			return
		default:
			counter++
			message := fmt.Sprintf("消息 #%d 来自 %s", counter, publisherID)

			// 添加消息到流
			id, err := rdb.XAdd(ctx, &redis.XAddArgs{
				Stream: streamKey,
				ID:     "*", // 自动生成ID
				Values: map[string]interface{}{
					"publisher": publisherID,
					"content":   message,
					"timestamp": time.Now().UnixNano(),
				},
			}).Result()

			if err != nil {
				t.Logf("发布者 %s 发布消息失败: %v", publisherID, err)
			} else {
				t.Logf("发布者 %s 发布消息成功, ID: %s, 内容: %s", publisherID, id, message)
			}

			// 短暂休眠，避免消息过多
			time.Sleep(500 * time.Millisecond)
		}
	}
}

// 从Redis Stream消费消息
func consumeMessages(ctx context.Context, rdb *redis.Client, streamKey, groupID, consumerID string, stopCh <-chan struct{}, t *testing.T) {
	for {
		select {
		case <-stopCh:
			return
		default:
			// 从流中读取新消息
			streams, err := rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
				Group:    groupID,
				Consumer: consumerID,
				Streams:  []string{streamKey, ">"}, // ">" 表示只读取未被组内消费的消息
				Count:    2,                        // 每次最多读取2条
				Block:    time.Second,              // 如果没有消息，最多阻塞1秒
			}).Result()

			if err != nil {
				if err != redis.Nil {
					t.Logf("消费者 %s-%s 读取消息失败: %v", groupID, consumerID, err)
				}
				time.Sleep(100 * time.Millisecond)
				continue
			}

			// 处理读取到的消息
			for _, stream := range streams {
				for _, message := range stream.Messages {
					t.Logf("消费者 %s-%s 收到消息: ID=%s, 内容=%v",
						groupID, consumerID, message.ID, message.Values)

					// 确认消息已处理
					err := rdb.XAck(ctx, streamKey, groupID, message.ID).Err()
					if err != nil {
						t.Logf("消费者 %s-%s 确认消息 %s 失败: %v",
							groupID, consumerID, message.ID, err)
					}
				}
			}

			// 短暂休眠，避免CPU使用率过高
			time.Sleep(100 * time.Millisecond)
		}
	}
}
