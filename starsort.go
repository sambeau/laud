// starsort.go

package main

import(
	// "fmt"
	"math"
)

// 	http://www.evanmiller.org/ranking-items-with-star-ratings.html

func sum(ns []int) int {
	var total int
	for _,n := range ns {
		total += n
	}
	return total
}

func f(s []int, ns []int)float64{
	N := sum(ns)
	K := len(ns)
	ks := make([]int,K)
	for i:=0; i<5; i++ {
		ks[i] = s[i] * (ns[i]+1)
	}
	return float64(sum(ks)) / float64(N+K)
}

func starSort(ns []int) float64 {
	N := sum(ns)
	K := len(ns)
	s := []int{5, 4, 3, 2, 1}
	s2 := []int{25, 16, 9, 4, 1}
	z := float64(1.65)
	fsns := f(s, ns)
	return fsns - z * math.Sqrt(((f(s2, ns) - (fsns*fsns)) / float64(N+K+1)))
}

// func main(){
// 	fmt.Println(starSort([]int{60, 80, 75, 20, 25}))
// 	fmt.Println(starSort([]int{1000, 0, 0, 0, 0}))
// 	fmt.Println(starSort([]int{1, 0, 0, 0, 0}))
// 	fmt.Println(starSort([]int{10, 0, 0, 0, 0}))
// }
