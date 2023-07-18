package main

import "fmt"

type MyInface interface {
	TThree(count int) (result int, err error)
}

type Mytest interface {
	TOne(count int) (result int, err error)
	TTwo(name string) (result string, err error)
}

type MyObject struct {
	client MyInface
}

func (m *MyObject) TOne(count int) (result int, err error) {
	return 1, nil
}

func (m *MyObject) TTwo(name string) (result string, err error) {
	return "1", nil
}

func (m *MyObject) TThree(count int) (result int, err error) {
	return 100, nil
}

var _ Mytest = (*MyObject)(nil)
var _ MyInface = (*MyObject)(nil)

func main() {
	var m = &MyObject{}
	fmt.Println(m.TOne(1))
	fmt.Println(m.TTwo("1"))
	fmt.Println(m.TThree(0))
}
