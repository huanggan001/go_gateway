package public

import (
	"fmt"
	"testing"
)

func TestCheck(t *testing.T) {
	password := GenSaltPassword("admin", "123456")
	fmt.Println(password)
}
