package main

type T3 interface {
	T2
	F3()
}

type T2 interface {
	T1
	F2()
}

type T1 interface {
	F1()
}

/*
type Test[K comparable, V any] interface {
	Set(K, V)
}

func Add[K comparable, V T1](m Test[K, V], k K, e V) V {
	m.Set(k, e)
	return e
}

func AddE(m Test[int, T1], k int, e T1) T1 {
	m.Set(k, e)
	return e
}

func main() {
	fmt.Printf("Hallo")

	var t1 Test[int, T1]
	var t2 Test[int, T2]
	var t3 Test[int, T3]
	var e3 T3
	var e1 T1

	AddE(t1, 1, e3)
	Add(t1, 1, e3)
	Add(t2, 1, e3)
	Add(t3, 1, e3)
	//e3 = Add(t2, 1, e3)
	//e3 = Add(t3, 1, e3)
	_ = e1

}

*/
