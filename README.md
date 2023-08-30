# timerpools
timerpools for timer reuse in game projects

### Features:

1. Automatic expansion, when the number of instances exceeds 1000 (can be modified), a new timer is automatically added for expansion
2. The instance can be stopped flexibly, which is suitable for the countdown in the game
3. Strong performance, 1 million timer usage scenarios, no more than 100 timers are actually created

### Quick Start:

````
package main

import (
	"github.com/gobkc/timerpools"
	"fmt"
	"time"
)

func main() {
	timerpools.SetMaxGoroutine(50000)
	t := timerpools.NewTimerTools()
    endSecond := 5
	t.Subscribe(1*time.Second, func(ctx *TimeContext) {
        if endSecond >= 0 {
            fmt.Println(`Countdown:`, endSecond)
            fmt.Println(`pool:`, ctx.Pid, "timerId:", ctx.TimerId, " now:", ctx.Now, " delay:", ctx.Delay)
            endSecond--
        } else {
            ctx.Stop()
            fmt.Println(`game over`)
        }
	})
}
````


### License
© Gobkc, 2023~time.Now

Released under the Apache License