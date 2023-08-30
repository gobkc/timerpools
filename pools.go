package timerpools

import (
	"fmt"
	"strconv"
	"sync"
	"time"
)

const poolId = 10000

var (
	maxPoolId    = 10000
	maxGoroutine = 1000
	goroutineNum = 0
	chLimit      chan struct{}
	once         sync.Once
	pool         *TimerPool
)

type poolIdType = int
type timerIdType = int64

type TimerPool struct {
	cb         map[poolIdType][]*TimerPoolRuntime
	l          sync.Mutex // global lock
	pl         sync.Mutex // pool lock
	state      sync.Map   //map[poolIdType]struct{}
	timerIdMap map[timerIdType]*TimerPoolRuntime
	poolIds    []poolMeta
}

func NewTimerPool() *TimerPool {
	once.Do(func() {
		pool = &TimerPool{}
		pool.initCb()
	})
	return pool
}

type TimerPoolRuntime struct {
	fc      func(ctx *TimeContext)
	poolId  int
	timerId int64
}

type poolMeta struct {
	poolId poolIdType
	delay  time.Duration
}

type TimeContext struct {
	pointer *TimerPool
	//context.Context todo Future features
	Pid     poolIdType
	TimerId timerIdType
	Delay   time.Duration
	Now     *time.Time
}

func (c *TimeContext) Stop() {
	c.pointer.Stop(c.TimerId)
}

func SetMaxGoroutine(num int) {
	maxGoroutine = num
	chLimit = make(chan struct{}, num)
}

func (p *TimerPool) startPool(delay time.Duration, pid poolIdType) {
	<-chLimit
	p.pl.Lock()
	if _, ok := p.state.Load(pid); !ok {
		ticker := time.NewTicker(delay)
		p.state.Store(pid, struct{}{})
		p.pl.Unlock()
		for now := range ticker.C {
			p.l.Lock()
			if poolRuntime, queryOk := p.cb[pid]; queryOk {
				for _, rt := range poolRuntime {
					go rt.fc(&TimeContext{
						pointer: p,
						Pid:     rt.poolId,
						TimerId: rt.timerId,
						Delay:   delay,
						Now:     &now,
					})
				}
			}
			p.l.Unlock()
		}
	} else {
		p.pl.Unlock()
	}
}

func (p *TimerPool) Stop(timerId int64) {
	p.l.Lock()
	defer p.l.Unlock()
	if timerId < 0 {
		return
	}
	queryTimer, ok := p.timerIdMap[timerId]
	if ok {
		delete(p.timerIdMap, timerId)
		if queryPools, queryOk := p.cb[queryTimer.poolId]; queryOk {
			for i, datum := range queryPools {
				if datum.timerId == queryTimer.timerId {
					copy(p.cb[queryTimer.poolId][i:], p.cb[queryTimer.poolId][i+1:])
					p.cb[queryTimer.poolId] = p.cb[queryTimer.poolId][:len(p.cb[queryTimer.poolId])-1]
					goroutineNum--
					return
				}
			}
		}
	}
}

func (p *TimerPool) initCb() {
	p.cb = make(map[poolIdType][]*TimerPoolRuntime)
	p.timerIdMap = make(map[timerIdType]*TimerPoolRuntime)
	p.poolIds = []poolMeta{{
		poolId: poolId,
		delay:  time.Second,
	}}
	chLimit = make(chan struct{}, maxGoroutine)
}

func (p *TimerPool) Subscribe(delay time.Duration, callback func(ctx *TimeContext)) {
	p.l.Lock()

	if p.cb == nil {
		p.initCb()
	}
	// append subscription to pool
	var queryPoolId int
	for i := poolId; i <= maxPoolId; i++ {
		query, ok := p.cb[i]
		if ok {
			if len(query) < maxGoroutine {
				for _, pid := range p.poolIds {
					if pid.delay == delay {
						queryPoolId = i
					}
				}
			}
		}
	}
	if queryPoolId == 0 {
		maxPoolId++
		queryPoolId = maxPoolId + 1
		p.poolIds = append(p.poolIds, poolMeta{
			poolId: queryPoolId,
			delay:  delay,
		})
	}
	goroutineNum++
	idLen := len(fmt.Sprintf(`%v`, goroutineNum)) + 1
	timberId := fmt.Sprintf(`%d%.`+fmt.Sprintf(`%v`, idLen)+`d`, queryPoolId, goroutineNum)
	timerId64, _ := strconv.ParseInt(timberId, 10, 64)
	timerRuntime := &TimerPoolRuntime{
		fc:      callback,
		poolId:  queryPoolId,
		timerId: timerId64,
	}
	p.cb[queryPoolId] = append(p.cb[queryPoolId], timerRuntime)
	p.timerIdMap[timerId64] = timerRuntime
	p.l.Unlock()
	chLimit <- struct{}{}
	go p.startPool(delay, queryPoolId)
}
