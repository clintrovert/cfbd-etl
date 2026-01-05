package utils

import (
   "strconv"
   "time"
)

func YearsFrom2005ToNow() []int32 {
   currentYear := time.Now().Year()
   years := make([]int32, 0, currentYear-2005+1)

   for y := 2005; y <= currentYear; y++ {
      years = append(years, int32(y))
   }

   return years
}

func Int32ToString(val int32) string {
   return strconv.FormatInt(int64(val), 10)
}
