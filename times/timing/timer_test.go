package timing_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/foredata/nova/times/timing"
)

func TestTimer(t *testing.T) {
	fmt.Printf("start %+v\n", time.Now())

	timing.NewDelayer(time.Second*2, func(data interface{}) {
		fmt.Printf("delay %+v, 2\n", time.Now())
	}, nil)

	timing.NewDelayer(time.Second*4, func(data interface{}) {
		fmt.Printf("delay %+v, 4\n", time.Now())
	}, nil)

	i := 0
	timing.NewTicker(time.Second, func(data interface{}) {
		i++
		fmt.Printf("tick  %+v, %+v\n", time.Now(), i)
	}, nil)

	time.Sleep(time.Second * 10)
}
