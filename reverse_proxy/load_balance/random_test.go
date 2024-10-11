package load_balance

import (
	"fmt"
	"testing"
)

func TestRandomBalance(t *testing.T) {
	//rb := &RandomBalance{}
	rb := &WeightRoundRobinBalance{}
	rb.Add("127.0.0.1:2003", "4")
	rb.Add("127.0.0.1:2004", "3")
	rb.Add("127.0.0.1:2005", "2")

	fmt.Println(rb.Next())
	fmt.Println(rb.Next())
	fmt.Println(rb.Next())
	fmt.Println(rb.Next())
	fmt.Println(rb.Next())
	fmt.Println(rb.Next())
	fmt.Println(rb.Next())
	fmt.Println(rb.Next())
	fmt.Println(rb.Next())
	fmt.Println(rb.Next())
	fmt.Println(rb.Next())
	fmt.Println(rb.Next())

}
