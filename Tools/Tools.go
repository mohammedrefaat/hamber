package tools

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"

	// _ "time/tzdata"

	timestamp "github.com/golang/protobuf/ptypes/timestamp"
	"github.com/jinzhu/copier"
	"github.com/shopspring/decimal"
)

const layout = "02/01/2006"

func RoundTo(n *float64, decimals uint32) *float64 {
	if n == nil {
		return nil
	}
	f := math.Round((*n)*math.Pow(10, float64(decimals))) / math.Pow(10, float64(decimals))
	return &f
}
func RoundToD(n *float64, decimals uint32) *float64 {
	if n == nil {
		return ToFloatPointer(0)
	}
	f := math.Round((*n)*math.Pow(10, float64(decimals))) / math.Pow(10, float64(decimals))
	return &f
}

func ToFloatPointer(fl float64) *float64 {
	return &fl
}

func FloatPtrToFloat(fl *float64) float64 {
	if fl == nil {
		return 0
	}
	return *fl
}

func FloatPtrToInt32Ptr(fl *float64) *int32 {
	if fl == nil {
		return nil
	}
	i := FloatPtrToInt32(fl)
	return &i
}

func FloatPtrToInt32(fl *float64) int32 {
	if fl == nil {
		return 0
	}
	return int32(*fl)
}

func PtrInt64ToFloatPtr(i *int64) *float64 {
	if i == nil {
		return nil
	}

	fl := ToFloatPointer(float64(int32(*i)))

	return fl
}

func Int32PtrPtrToInt32(fl *int32) int32 {
	if fl == nil {
		return 0
	}
	return *fl
}

func Int32PtrToInt64Ptr(fl *int32) *int64 {
	if fl == nil {
		return nil
	}
	return ToIntPointer(int64(*fl))
}

func Int64PtrToInt32Ptr(fl *int64) *int32 {
	if fl == nil {
		return nil
	}
	return Int32ToInt32Ptr(int32(*fl))
}

func Int32ToInt32Ptr(fl int32) *int32 {
	return &fl
}

func IntPtrPtrToInt(fl *int64) int64 {
	if fl == nil {
		return 0
	}
	return *fl
}

func ToIntPointer(int642 int64) *int64 {
	return &int642
}

func ToStringPointer(s string) *string {
	return &s
}

func ToBoolPointer(s bool) *bool {
	return &s
}

func Float64ToString(f *float64) *string {
	if f == nil {
		return nil
	}
	str := strconv.FormatFloat(*f, 'f', -1, 64)
	return &str
}

func Float32ToString(f *float32) *string {
	if f == nil {
		return nil
	}
	f64 := float64(*f)
	str := strconv.FormatFloat(f64, 'f', -1, 32)
	return &str
}

func Int64ToString(f *int64) *string {
	if f == nil {
		return nil
	}
	str := strconv.FormatInt(*f, 10)
	return &str
}

func Int32ToString(f *int32) *string {
	if f == nil {
		return nil
	}
	i := int64(*f)
	str := strconv.FormatInt(i, 10)
	return &str
}

func StringToInt32(f *string) *int32 {
	if f == nil {
		return nil
	}
	i, err := strconv.Atoi(*f)
	if err != nil {
		return nil
	}
	r := int32(i)
	return &r
}

func StringToInt322(f string) int32 {
	return *StringToInt32(&f)
}

func StringToInt64(f *string) *int64 {
	if f == nil {
		return nil
	}
	i, err := strconv.Atoi(*f)
	if err != nil {
		return nil
	}
	r := int64(i)
	return &r
}

func StringCompare(s1 string, s2 string) bool {
	return strings.TrimSpace(s1) == strings.TrimSpace(s2)
}

func StringComparePointer(s1 *string, s2 *string) bool {
	if s1 == nil || s2 == nil {
		return false
	}
	return StringCompare(*s1, *s2)
}

func StringComparePointer2(s1 string, s2 *string) bool {
	if s2 == nil {
		return false
	}
	return StringCompare(s1, *s2)
}

func BoolToString(b *bool) *string {
	if b == nil {
		return nil
	}
	one := "1"
	zero := "0"
	if *b {
		return &one
	}
	return &zero
}

func Sum1(floats ...*float64) float64 {
	var sm float64 = 0
	for id := range floats {
		if floats[id] != nil {
			sm = sm + *floats[id]
		}
	}
	return sm
}

func Max1(floats ...*float64) *float64 {
	var mx *float64 = nil
	for id := range floats {
		if floats[id] != nil {
			if mx == nil || *mx < *floats[id] {
				mx = floats[id]
			}
		}
	}
	return mx
}

func Min1(floats ...*float64) *float64 {
	var mx *float64 = nil
	for id := range floats {
		if floats[id] != nil {
			if mx == nil || *mx > *floats[id] {
				mx = floats[id]
			}
		}
	}
	return mx
}

func Divide1(m *float64, n *float64) *float64 {
	if m == nil || n == nil {
		return nil
	}
	if *n == 0 {
		return nil
	}
	dv := (*m) / (*n)
	return &dv
}

func Multiply1(m *float64, n *float64) *float64 {
	if m == nil || n == nil {
		return nil
	}
	rs := (*m) * (*n)
	return &rs
}

func DefaultF(m *float64, n float64) float64 {
	if m == nil {
		return n
	}
	return *m
}

func DefaultB(m *bool, n bool) bool {
	if m == nil {
		return n
	}
	return *m
}

func DefaultI(m *int64, n int64) int64 {
	if m == nil {
		return n
	}
	return *m
}
func DefaultInt(m *int32, n int32) int32 {
	if m == nil {
		return n
	}
	return *m
}

func DefaultInt64(m *int64, n int64) int64 {
	if m == nil {
		return n
	}
	return *m
}

func DefaultString(m *string, n string) string {
	if m == nil {
		return n
	}
	return *m
}

func DefaultTime(m *time.Time, n time.Time) time.Time {
	if m == nil {
		return n
	}
	return *m
}

func DefaultTimeStamp(m *timestamp.Timestamp, n time.Time) time.Time {
	if m == nil {
		return n
	}
	return m.AsTime()
}

func ToTimePrt(m time.Time) *time.Time {
	return &m
}

func cleanString(inp *string, defult *string, toUpper *bool, toLower *bool) {
	if inp == nil {
		return
	}
	toUpperv := false
	toLowerv := false
	if toUpper != nil {
		toUpperv = *toUpper
	}
	if toLower != nil {
		toLowerv = *toLower
	}
	if *inp == "''" {
		inp = defult
	}
	inp = ToStringPointer(strings.TrimSpace(*inp))
	if strings.ToLower(*inp) == "undefined" || strings.ToLower(*inp) == "null" {
		inp = defult
	}
	inp = ToStringPointer(strings.TrimLeft(*inp, "'"))
	inp = ToStringPointer(strings.TrimRight(*inp, "'"))
	inp = ToStringPointer(strings.Replace(*inp, "\"", " ", -1))
	inp = ToStringPointer(strings.Replace(*inp, "'", " ", -1))
	inp = ToStringPointer(strings.TrimSpace(*inp))
	if toUpperv {
		inp = ToStringPointer(strings.ToUpper(*inp))
	}
	if toLowerv {
		inp = ToStringPointer(strings.ToLower(*inp))
	}
}

type filterf func(interface{}) bool

func filterFirst(in interface{}, fn filterf) interface{} {
	val := reflect.ValueOf(in)
	out := make([]interface{}, 0, val.Len())

	for i := 0; i < val.Len(); i++ {
		current := val.Index(i).Interface()

		if fn(current) {
			out = append(out, current)
			break
		}
	}

	return out
}

// Filter For Arrays
func Filter(in interface{}, fn filterf) interface{} {
	val := reflect.ValueOf(in)
	out := make([]interface{}, 0, val.Len())

	for i := 0; i < val.Len(); i++ {
		current := val.Index(i).Interface()

		if fn(current) {
			out = append(out, current)
		}
	}

	return out
}

// Exists For Arrays
func Exists(in interface{}, fn filterf) bool {
	val := reflect.ValueOf(in)

	for i := 0; i < val.Len(); i++ {
		current := val.Index(i).Interface()

		if fn(current) {
			return true
		}
	}

	return false
}

func IsEnglish(str string) (bool, string) {
	if strings.TrimSpace(str) == "" {
		return false, ""
	}
	strr := strings.TrimSpace(str)
	strew := strings.ToLower(strr)
	for idx := range strew {
		s := string(strew[idx])
		if !(s == "a" || s == "b" || s == "c" || s == "d" || s == "e" || s == "f" || s == "g" || s == "h" || s == "i" || s == "j" || s == "k" || s == "l" || s == "m" || s == "n" || s == "o" || s == "p" || s == "q" || s == "r" || s == "s" || s == "t" || s == "u" || s == "v" || s == "w" || s == "x" || s == "y" || s == "z" || s == "_" || s == " ") {
			return false, ""
		}
	}
	return true, strr
}

func ArrayToString(delim string, values ...interface{}) string {
	return strings.Trim(strings.Replace(fmt.Sprint(values...), " ", delim, -1), "[]")
}

var (
	pencilPrint   = color(pencilColor_PrintColor)
	errorPrint    = color(errorColor_PrintColor)
	successPrint  = color(successColor_PrintColor)
	warningPrint  = color(warningColor_PrintColor)
	noticePrint   = color(noticeColor_PrintColor)
	infoPrint     = color(infoColor_PrintColor)
	questionPrint = color(questionColor_PrintColor)
	defaultPrint  = color(defaultColor_PrintColor)
)

type PrintColor string

const (
	pencilColor_PrintColor   PrintColor = "\033[1;30m%s\033[0m"
	errorColor_PrintColor    PrintColor = "\033[1;31m%s\033[0m"
	successColor_PrintColor  PrintColor = "\033[1;32m%s\033[0m"
	warningColor_PrintColor  PrintColor = "\033[1;33m%s\033[0m"
	noticeColor_PrintColor   PrintColor = "\033[1;34m%s\033[0m"
	infoColor_PrintColor     PrintColor = "\033[1;35m%s\033[0m"
	questionColor_PrintColor PrintColor = "\033[1;36m%s\033[0m"
	defaultColor_PrintColor  PrintColor = "\033[1;37m%s\033[0m"
)

func color(colorString PrintColor) func(...interface{}) string {
	sprint := func(args ...interface{}) string {
		return fmt.Sprintf(string(colorString),
			fmt.Sprint(args...))
	}
	return sprint
}

func PrintWarning(doneReg *bool, trace *bool, a ...interface{}) {
	if trace != nil && *trace {
		if doneReg != nil && *doneReg {
			fmt.Println(warningPrint(a...))
		} else {
			fmt.Println(defaultPrint(a...))
		}
	}
}
func PrintError(doneReg *bool, trace *bool, onlyError *bool, a ...interface{}) {
	if (trace != nil && *trace) || (onlyError != nil && *onlyError) {
		if doneReg != nil && *doneReg {
			fmt.Println(errorPrint(a...))
		} else {
			fmt.Println(defaultPrint(a...))
		}
	}
}
func PrintInfo(doneReg *bool, trace *bool, a ...interface{}) {
	if trace != nil && *trace {
		if doneReg != nil && *doneReg {
			fmt.Println(infoPrint(a...))
		} else {
			fmt.Println(defaultPrint(a...))
		}
	}
}
func PrintQuestion(doneReg *bool, trace *bool, a ...interface{}) {
	if trace != nil && *trace {
		if doneReg != nil && *doneReg {
			fmt.Println(questionPrint(a...))
		} else {
			fmt.Println(defaultPrint(a...))
		}
	}
}
func PrintNotice(doneReg *bool, trace *bool, a ...interface{}) {
	if trace != nil && *trace {
		if doneReg != nil && *doneReg {
			fmt.Println(noticePrint(a...))
		} else {
			fmt.Println(defaultPrint(a...))
		}
	}
}
func PrintSuccess(doneReg *bool, trace *bool, a ...interface{}) {
	if trace != nil && *trace {
		if doneReg != nil && *doneReg {
			fmt.Println(successPrint(a...))
		} else {
			fmt.Println(defaultPrint(a...))
		}
	}
}
func PrintPencil(doneReg *bool, trace *bool, a ...interface{}) {
	if trace != nil && *trace {
		if doneReg != nil && *doneReg {
			fmt.Println(pencilPrint(a...))
		} else {
			fmt.Println(defaultPrint(a...))
		}
	}
}

func PrintCustom(doneReg *bool, trace *bool, onlyError *bool, customColor PrintColor, a ...interface{}) {
	if (trace != nil && *trace) || (onlyError != nil && *onlyError) {
		if doneReg != nil && *doneReg {
			var customPrint = color(customColor)
			fmt.Println(customPrint(a...))
		} else {
			fmt.Println(defaultPrint(a...))
		}
	}
}

func HasDuplicatePtrString(arr ...*string) bool {
	visited := make(map[string]bool, 0)
	for i := 0; i < len(arr); i++ {
		if visited[*arr[i]] == true {
			return true
		} else {
			visited[*arr[i]] = true
		}
	}
	return false
}

func HasDuplicateString(arr ...string) bool {
	visited := make(map[string]bool, 0)
	for i := 0; i < len(arr); i++ {
		if visited[arr[i]] == true {
			return true
		} else {
			visited[arr[i]] = true
		}
	}
	return false
}
func MaxTime(t ...*time.Time) *time.Time {
	var mx *time.Time = nil
	for id := range t {
		if t[id] != nil {
			if mx == nil || mx.Before(*t[id]) {
				mx = t[id]
			}
		}
	}
	return mx
}

func MinTime(t ...*time.Time) *time.Time {
	var mx *time.Time = nil
	for id := range t {
		if t[id] != nil {
			if mx == nil || mx.After(*t[id]) {
				mx = t[id]
			}
		}
	}
	return mx
}

func DistinctPtrInt32(intSlice ...*int32) []*int32 {
	keys := make(map[int32]bool)
	list := []*int32{}
	for _, entry := range intSlice {
		if _, value := keys[*entry]; !value {
			keys[*entry] = true
			list = append(list, entry)
		}
	}
	return list
}
func DistinctPtrInt64(intSlice ...*int64) []*int64 {
	keys := make(map[int64]bool)
	list := []*int64{}
	for _, entry := range intSlice {
		if _, value := keys[*entry]; !value {
			keys[*entry] = true
			list = append(list, entry)
		}
	}
	return list
}
func DistinctInt32(intSlice ...int32) []int32 {
	keys := make(map[int32]bool)
	list := []int32{}
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
func DistinctInt64(intSlice ...int64) []int64 {
	keys := make(map[int64]bool)
	list := []int64{}
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
func DistinctPtrString(intSlice ...*string) []*string {
	keys := make(map[string]bool)
	list := []*string{}
	for _, entry := range intSlice {
		if _, value := keys[*entry]; !value {
			keys[*entry] = true
			list = append(list, entry)
		}
	}
	return list
}
func DistinctString(intSlice ...string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
func MinInt32(ints ...*int32) *int32 {
	var mx *int32 = nil
	for id := range ints {
		if ints[id] != nil {
			if mx == nil || *mx > *ints[id] {
				mx = ints[id]
			}
		}
	}
	return mx
}

func MinInt64(ints ...*int64) *int64 {
	var mx *int64 = nil
	for id := range ints {
		if ints[id] != nil {
			if mx == nil || *mx > *ints[id] {
				mx = ints[id]
			}
		}
	}
	return mx
}
func MaxInt32(ints ...*int32) *int32 {
	var mx *int32 = nil
	for id := range ints {
		if ints[id] != nil {
			if mx == nil || *mx < *ints[id] {
				mx = ints[id]
			}
		}
	}
	return mx
}

func MaxInt64(ints ...*int64) *int64 {
	var mx *int64 = nil
	for id := range ints {
		if ints[id] != nil {
			if mx == nil || *mx < *ints[id] {
				mx = ints[id]
			}
		}
	}
	return mx
}
func StringToInt32WithDefaultValue(str *string, val int32) *int32 {
	if str == nil {
		return &val
	}
	i, err := strconv.Atoi(*str)
	if err != nil {
		return &val
	}
	r := int32(i)
	return &r
}
func StringToInt64WithDefaultValue(str *string, val int64) *int64 {
	if str == nil {
		return &val
	}
	i, err := strconv.Atoi(*str)
	if err != nil {
		return &val
	}
	r := int64(i)
	return &r
}
func Copy(toVal interface{}, fromVal interface{}) error {
	if fromVal == nil {
		return errors.New("القيمة المرسلة غير صحيحة")
	}
	return copier.Copy(toVal, fromVal)
}

func GetMaxMonthDays(year int, month int) int {
	d := time.Date(year, time.Month(month+1), 0, 0, 0, 0, 0, time.UTC)
	return d.Day()
}

func SumFloats(f ...float64) float64 {
	sum := decimal.Decimal{}
	for i := range f {
		n := f[i]
		num := decimal.NewFromFloat(n)
		sum = sum.Add(num)
	}
	final, _ := sum.Float64()
	return final
}
func SumFloat(f []float64) float64 {
	sum := decimal.Decimal{}
	for i := range f {
		n := f[i]
		num := decimal.NewFromFloat(n)
		sum = sum.Add(num)
	}
	final, _ := sum.Float64()
	return final
}
func SubstractFloat(f []float64) float64 {
	sub := decimal.Decimal{}
	for i := range f {
		n := f[i]
		num := decimal.NewFromFloat(n)
		if i == 0 {
			sub = num
			continue
		}
		sub = sub.Sub(num)
	}
	final, _ := sub.Float64()
	return final
}

func MultiplyFloat(f []float64) float64 {
	multi := decimal.Decimal{}
	for i := range f {
		n := f[i]
		num := decimal.NewFromFloat(n)
		if i == 0 {
			multi = num
			continue
		}
		multi = multi.Mul(num)
	}
	final, _ := multi.Float64()
	return final
}
func DivFloat(f []float64) float64 {
	div := decimal.Decimal{}
	for i := range f {
		n := f[i]
		num := decimal.NewFromFloat(n)
		if i == 0 {
			div = num
			continue
		}
		div = div.Div(num)
	}
	final, _ := div.Float64()
	return final
}
