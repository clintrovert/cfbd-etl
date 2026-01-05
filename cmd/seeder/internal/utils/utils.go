//nolint:revive,lll // Package name is acceptable for utility functions
package utils

import (
	"strconv"
	"strings"
	"time"

	"github.com/lib/pq"
)

func YearsFrom2005ToNow() []int32 {
	currentYear := time.Now().Year()
	years := make([]int32, 0, currentYear-2005+1)

	for y := 2005; y <= currentYear; y++ {
		//nolint:gosec // Year values are always within int32 range
		years = append(years, int32(y))
	}

	return years
}

func Int32ToString(val int32) string {
	return strconv.FormatInt(int64(val), 10)
}

func Int32SliceToInt64Array(xs []int32) pq.Int64Array {
	if len(xs) == 0 {
		return nil
	}
	out := make(pq.Int64Array, 0, len(xs))
	for _, v := range xs {
		out = append(out, int64(v))
	}
	return out
}

func ToStringArray(in []string) pq.StringArray {
	if len(in) == 0 {
		// store empty array rather than NULL
		return pq.StringArray{}
	}
	out := make([]string, 0, len(in))
	for _, s := range in {
		v := strings.TrimSpace(s)
		if v == "" {
			continue
		}
		out = append(out, v)
	}
	return pq.StringArray(out)
}
