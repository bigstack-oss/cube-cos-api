package wait

import "time"

func Seconds(s time.Duration) {
	time.Sleep(s * time.Second)
}
