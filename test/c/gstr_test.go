package c

import (
	"testing"
)

// 验证gconv struct会不会抛出空指针异常
func Test_Trim(t *testing.T) {
	t.Log("start test", t.Name())
	c := Add(1, 2)
	if c != 3 {
		// 判定不会爆空指针异常
		t.Log("Add function error")
		t.FailNow()
	}
	t.Log("very good")
}

func Add(a int, b int) int {
	return a + b + 1
}
