package snowflake

import (
	"sync"
	"time"
)

// 雪花算法位分配（共64位int64）：
// 1位符号位（固定0） + 41位时间戳（毫秒） + 10位机器ID + 12位序列号
const (
	timestampBits = 41 // 时间戳占用位数（可支持约69年）
	machineIDBits = 10 // 机器ID占用位数（支持最多1024个节点）
	sequenceBits  = 12 // 序列号占用位数（每毫秒最多生成4096个ID）

	maxMachineID = (1 << machineIDBits) - 1 // 机器ID最大值（0~1023）
	maxSequence  = (1 << sequenceBits) - 1  // 序列号最大值（0~4095）
)

// Config 雪花算法初始化配置
type Config struct {
	MachineID int64  // 机器ID，范围 0~1023（必填）
	StartTime string // 起始时间（字符串格式："2023-01-01"，可选，默认2023-01-01 00:00:00 UTC+8）
}

// Snowflake 雪花算法实例
type Snowflake struct {
	mu            sync.Mutex // 并发安全锁
	startTime     int64      // 起始时间戳（毫秒，内部存储仍用int64）
	machineID     int64      // 机器ID（0~1023）
	lastTimestamp int64      // 上一次生成ID的时间戳（毫秒）
	sequence      int64      // 当前毫秒内的序列号
}

var instance *Snowflake // 全局单例实例

// 默认起始时间：2023-01-01 00:00:00 UTC+8（转成毫秒时间戳）
const defaultStartTime = 1672502400000

// 时间字符串解析格式（兼容"20XX-XX-XX"）
const timeLayout = "2006-01-02"

// Init 初始化雪花算法（必须先调用）
// config: 初始化配置，MachineID必填（0~1023），StartTime可选（不传则用默认值）
func Init(config Config) {
	// 1. 校验机器ID合法性
	if config.MachineID < 0 || config.MachineID > maxMachineID {
		panic("snowflake: MachineID超出范围（必须是0~1023）")
	}

	// 2. 处理起始时间字符串，解析为毫秒时间戳
	var startTime int64
	if config.StartTime == "" {
		// 无配置时用默认时间
		startTime = defaultStartTime
	} else {
		// 指定时区（UTC+8，避免本地时区干扰）
		loc, err := time.LoadLocation("Asia/Shanghai")
		if err != nil {
			panic("snowflake: 加载时区失败（Asia/Shanghai）：" + err.Error())
		}
		// 解析时间字符串（格式：2023-01-01）
		t, err := time.ParseInLocation(timeLayout, config.StartTime, loc)
		if err != nil {
			panic("snowflake: StartTime格式错误（需为20XX-XX-XX）：" + err.Error())
		}
		// 转成毫秒时间戳
		startTime = t.UnixMilli()
	}

	// 3. 校验起始时间不能是未来时间
	now := time.Now().UnixMilli()
	if startTime > now {
		panic("snowflake: StartTime不能晚于当前时间")
	}

	// 4. 初始化单例
	instance = &Snowflake{
		startTime:     startTime,
		machineID:     config.MachineID,
		lastTimestamp: -1, // 初始化为-1，确保首次生成时重置序列号
		sequence:      0,
	}
}

// GetID 生成唯一ID（需先调用Init初始化）
func GetID() int64 {
	if instance == nil {
		panic("snowflake: 未初始化，请先调用Init(config)")
	}
	return instance.nextID()
}

// nextID 生成单个ID（内部方法，已加锁）
func (s *Snowflake) nextID() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UnixMilli() // 当前时间戳（毫秒）

	// 1. 处理时钟回拨（若当前时间早于上次生成时间，阻塞等待到下一个毫秒）
	if now < s.lastTimestamp {
		// 生产环境可调整为阻塞等待，而非panic
		for now <= s.lastTimestamp {
			now = time.Now().UnixMilli()
		}
	}

	// 2. 处理同一毫秒内的序列号
	if now == s.lastTimestamp {
		s.sequence++
		// 序列号超出最大值，阻塞到下一个毫秒
		if s.sequence > maxSequence {
			for now <= s.lastTimestamp {
				now = time.Now().UnixMilli()
			}
			s.sequence = 0 // 重置序列号
		}
	} else {
		// 3. 新的毫秒，重置序列号
		s.sequence = 0
	}

	// 更新上次生成ID的时间戳
	s.lastTimestamp = now

	// 4. 组合ID：时间戳偏移 + 机器ID + 序列号
	return (now-s.startTime)<<(machineIDBits+sequenceBits) | // 时间戳部分（基于自定义起始时间的偏移）
		s.machineID<<sequenceBits | // 机器ID部分
		s.sequence // 序列号部分
}
