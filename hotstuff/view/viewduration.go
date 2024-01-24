package view

import (
	"math"
	"time"
)

// ViewDuration 确定视图的持续时间。
// 视图同步器使用此接口设置其超时时间。
type ViewDuration interface {
	Duration() time.Duration // Duration 返回下一个视图应持续的时间。
	ViewStarted()            // ViewStarted 在启动新视图时由同步器调用。
	ViewSucceeded()          // ViewSucceeded 在视图成功结束时由同步器调用。
	ViewTimeout()            // ViewTimeout 在视图超时时由同步器调用。
}

// NewViewDuration 返回一个ViewDuration，它基于先前视图的持续时间来近似视图持续时间。
// sampleSize 确定应考虑的先前视图的数量。
// startTimeout 确定前几个视图的视图持续时间。
// 当发生超时时，下一个视图持续时间将乘以乘数。
func NewViewDuration(sampleSize uint64, startTimeout, maxTimeout, multiplier float64) ViewDuration {
	return &viewDuration{
		limit: sampleSize,
		mean:  startTimeout,
		max:   maxTimeout,
		mul:   multiplier,
	}
}

// viewDuration 使用先前视图的统计数据来猜测视图持续时间的合适值。
// 它只考虑有限数量的测量。
type viewDuration struct {
	mul       float64   // 在失败的视图上，将当前平均值乘以此数（应大于1）
	limit     uint64    // 应包括在平均值中的测量数量
	count     uint64    // 总测量数量
	startTime time.Time // 当前测量的开始时间
	mean      float64   // 平均视图持续时间
	m2        float64   // 与平均值的差的平方和
	prevM2    float64   // 上一个周期计算的m2
	max       float64   // 视图超时的上限
}

// ViewSucceeded 计算视图的持续时间
// 并更新用于计算平均值和方差的内部值。
func (v *viewDuration) ViewSucceeded() {
	if v.startTime.IsZero() {
		return
	}

	duration := float64(time.Since(v.startTime)) / float64(time.Millisecond)
	v.count++

	// 偶尔重置m2，以便更快地检测方差的变化。
	// 我们将m2存储到prevM2中，在计算方差时将使用它。
	// 这确保至少'limit'个测量对近似方差做出了贡献。
	if v.count%v.limit == 0 {
		v.prevM2 = v.m2
		v.m2 = 0
	}

	var c float64
	if v.count > v.limit {
		c = float64(v.limit)
		// 丢弃一个测量值
		v.mean -= v.mean / float64(c)
	} else {
		c = float64(v.count)
	}

	// Welford算法
	d1 := duration - v.mean
	v.mean += d1 / c
	d2 := duration - v.mean
	v.m2 += d1 * d2
}

// ViewTimeout 在视图超时时应调用。它将当前平均值乘以'mul'。
func (v *viewDuration) ViewTimeout() {
	v.mean *= v.mul
}

// ViewStarted 记录视图的开始时间。
func (v *viewDuration) ViewStarted() {
	v.startTime = time.Now()
}

// Duration 返回平均视图持续时间的95%置信区间的上限。
func (v *viewDuration) Duration() time.Duration {
	conf := 1.96 // 95%置信度
	dev := float64(0)
	if v.count > 1 {
		c := float64(v.count)
		m2 := v.m2
		// 标准差是从prevM2和m2的总和计算的。
		if v.count >= v.limit {
			c = float64(v.limit) + float64(v.count%v.limit)
			m2 += v.prevM2
		}
		dev = math.Sqrt(m2 / c)
	}

	duration := v.mean + dev*conf
	if v.max > 0 && duration > v.max {
		duration = v.max
	}

	return time.Duration(duration * float64(time.Millisecond))
}
