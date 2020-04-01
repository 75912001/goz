package xrTime_test

import (
	"fmt"
	"testing"

	"github.com/75912001/goz/xrTime"
)

func TestGenYYYYMMDD(t *testing.T) {
	fmt.Println(xrTime.GenYYYYMMDD(1585199729))
}
