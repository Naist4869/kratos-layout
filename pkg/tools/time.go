package tools

import (
	"bytes"
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

var (
	location *time.Location
)

const (
	defaultLayout = "2006-01-02 15:04:05"
)

var (
	monthLayout  = defaultLayout[:7]
	dayLayout    = defaultLayout[:10]
	hourLayout   = defaultLayout[:13]
	minuteLayout = defaultLayout[:16]
	fullLayout   = defaultLayout
)

// Duration be used toml unmarshal string time, like 1s, 500ms.
type Duration time.Duration

// UnmarshalText unmarshal text to duration.
func (d *Duration) UnmarshalText(text []byte) error {
	tmp, err := time.ParseDuration(string(text))
	if err == nil {
		*d = Duration(tmp)
	}
	return err
}

// Shrink will decrease the duration by comparing with context's timeout duration
// and return new timeout\context\CancelFunc.
func (d Duration) Shrink(c context.Context) (Duration, context.Context, context.CancelFunc) {
	if deadline, ok := c.Deadline(); ok {
		if ctimeout := time.Until(deadline); ctimeout < time.Duration(d) {
			// deliver small timeout
			return Duration(ctimeout), c, func() {}
		}
	}
	ctx, cancel := context.WithTimeout(c, time.Duration(d))
	return d, ctx, cancel
}

// Time 封装的时间,主要是使用utc+8 时间
type Time time.Time

func (t *Time) Scan(value interface{}) (err error) {
	nullTime := &sql.NullTime{}
	err = nullTime.Scan(value)
	*t = Time(nullTime.Time)
	return
}

func (t Time) Value() (driver.Value, error) {
	y, m, d := time.Time(t).Date()
	if t.IsZero() {
		return time.Date(y, m, d, 0, 0, 0, 0, time.Time(t).Location()), nil
	}
	return time.Date(y, m, d, 0, 0, 0, 0, location), nil
}

// GormDataType gorm common data type
func (t Time) GormDataType() string {
	return "date"
}

func (t Time) GobEncode() ([]byte, error) {
	return time.Time(t).GobEncode()
}

func (t *Time) GobDecode(b []byte) error {
	return (*time.Time)(t).GobDecode(b)
}

/*Now 返回当前时间
参数:
返回值:
*	Time	Time
*/
func Now() Time {
	return Time(time.Now())
}

/*WeekStart 返回本周开始时间，即周一00:00:00
参数:
*	t	Time
返回值:
*	Time	Time
*/
func (t Time) WeekStart() Time {
	stdTime := time.Time(t)
	weekDay := stdTime.Weekday()
	day := stdTime.Day()
	switch weekDay {
	case time.Sunday: // Sunday 值为0,
		day -= 6
	default:
		day -= int(weekDay) - 1
	}
	return Time(time.Date(stdTime.Year(), stdTime.Month(), day, 0, 0, 0, 0, location))
}

/*WeekEnd 返回本周结束时间，即周日23:59:59
参数:
*	t	Time
返回值:
*	Time	Time
*/
func (t Time) WeekEnd() Time {
	stdTime := time.Time(t)
	weekDay := stdTime.Weekday()
	day := stdTime.Day()
	switch weekDay {
	case time.Sunday: // 周日不用处理

	default: // 其天时间，比如周1 需要加 7-1==6 天¬
		day += 7 - int(weekDay)
	}
	return Time(time.Date(stdTime.Year(), stdTime.Month(), day, 23, 59, 59, 0, location))
}

/*WeekRange 返回当前所在的周一00:00和周日23:59:59
参数:
返回值:
*	start	Time    开始
*	end  	Time    结束
*/
func (t Time) WeekRange() (start, end Time) {
	return t.WeekStart(), t.WeekEnd()
}

/*WeekStart 返回本月开始时间，即1号00:00:00
参数:
*	t	Time
返回值:
*	Time	Time
*/
func (t Time) MonthStart() Time {
	stdTime := time.Time(t)

	return Time(time.Date(stdTime.Year(), stdTime.Month(), 1, 0, 0, 0, 0, location))
}

/*MonthEnd 返回本月结束时间，即最后一天23:59:59
参数:
返回值:
*	Time	Time
*/
func (t Time) MonthEnd() Time {
	stdTime := time.Time(t)
	month := stdTime.Month()
	day := 31
	switch month {
	case time.February: // 2 月
		if year := stdTime.Year(); year%400 == 0 || (year%100 != 0 && year%4 == 0) { // 闰年29天
			day = 29
		} else { // 非闰年28天
			day = 28
		}

	case 4, 6, 9, 11: // 小月 30天
		day = 30
	}
	return Time(time.Date(stdTime.Year(), stdTime.Month(), day, 23, 59, 59, 0, location))
}

/*MonthRange 返回本月开始结束
参数:
返回值:
*	Time	Time    开始
*	Time	Time    结束
*/
func (t Time) MonthRange() (Time, Time) {
	return t.MonthStart(), t.MonthEnd()
}

func (t Time) DayStart() Time {
	stdTime := time.Time(t)
	return Time(time.Date(stdTime.Year(), stdTime.Month(), stdTime.Day(), 0, 0, 0, 0, location))
}

func (t Time) DayEnd() Time {
	stdTime := time.Time(t)
	return Time(time.Date(stdTime.Year(), stdTime.Month(), stdTime.Day(), 23, 59, 59, 0, location))
}

func (t Time) DayRange() (Time, Time) {
	return t.DayStart(), t.DayEnd()
}

func EachDay(from, to Time) []Time {
	if to.After(from) {
		year := time.Time(from).Year()
		fromDay := time.Time(from).YearDay()
		endDay := time.Time(to).YearDay()
		result := make([]Time, 0, endDay-fromDay+1)
		result = append(result, from)
		if endDay != fromDay {
			for i := fromDay + 1; i < endDay; i++ {
				result = append(result, Time(time.Date(year, 1, i, 0, 0, 0, 0, location)))
			}
			result = append(result, to)
		}
		return result
	}
	return nil
}
func (t Time) SameDay(another Time) bool {
	return time.Time(t).Year() == time.Time(another).Year() && time.Time(t).YearDay() == time.Time(another).YearDay()
}

// DeepCopyInto creates a deep-copy of the Time value.  The underlying time.Time
// type is effectively immutable in the time API, so it is safe to
// copy-by-assign, despite the presence of (unexported) Pointer fields.
func (t *Time) DeepCopyInto(out *Time) {
	*out = *t
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Time.
func (t *Time) DeepCopy() *Time {
	if t == nil {
		return nil
	}
	out := new(Time)
	t.DeepCopyInto(out)
	return out
}

func (t Time) MarshalCSV() ([]byte, error) {
	return []byte(time.Time(t).Format(defaultLayout)), nil
}

func (t *Time) UnmarshalCSV(data []byte) error {
	data = bytes.Trim(data, `""`)
	data = bytes.TrimPrefix(data, []byte("`"))
	s := string(data)
	if s == "--" || s == "" {
		return nil
	}
	if len(data) == 10 {
		parseInt, err := strconv.ParseInt(s, 10, 64)
		if err == nil {
			*t = Time(time.Unix(parseInt, 0))
			return nil
		}
	}
	parsedTime, err := ParseTimeInLength(s)
	if err != nil {
		return err
	}
	*t = parsedTime
	return nil
}

func (t *Time) UnmarshalJSON(data []byte) error {
	data = bytes.Trim(data, `""`)
	s := string(data)
	if s == "--" || s == "" {
		return nil
	}
	if len(data) == 10 {
		parseInt, err := strconv.ParseInt(s, 10, 64)
		if err == nil {
			*t = Time(time.Unix(parseInt, 0))
			return nil
		}
	}
	parsedTime, err := ParseTimeInLength(s)
	if err != nil {
		return err
	}
	*t = parsedTime
	return nil
}

func (t *Time) UnmarshalText(data []byte) error {
	data = bytes.Trim(data, `""`)
	if string(data) == "" {
		return nil
	} else {
		parsedTime, err := ParseTimeInLength(string(data))
		if err != nil {
			return err
		} else {
			*t = Time(parsedTime)
			return nil
		}
	}
}
func (t Time) IsZero() bool {
	return time.Time(t).In(location).IsZero()
}
func (t Time) MarshalText() ([]byte, error) {
	if t.IsZero() {
		return nil, nil
	}
	return []byte(t.String()), nil
}

func (t Time) String() string {
	return t.Format(defaultLayout)
}

func (t Time) Format(layout string) string {
	if t.IsZero() {
		return ""
	}
	return time.Time(t).In(location).Format(layout)
}
func (t Time) StringDay() string {
	return t.Format(dayLayout)
}
func (t Time) StringSecond() string {
	return t.Format(fullLayout)
}
func (t Time) StringHour() string {
	return t.Format(hourLayout)
}
func (t Time) StringMinute() string {
	return t.Format(minuteLayout)
}
func (t Time) StringMonth() string {
	return t.Format(monthLayout)
}
func (t *Time) UnmarshalBSONValue(bt bsontype.Type, v []byte) error {
	value := bsoncore.Value{
		Type: bt,
		Data: v,
	}
	if value, ok := value.TimeOK(); !ok {
		return fmt.Errorf("错误的类型[%s],不是Time", t.String())
	} else {
		*t = Time(value)
		return nil
	}
}
func (t Time) MarshalBSONValue() (bsontype.Type, []byte, error) {
	return bsontype.DateTime, bsoncore.AppendTime(nil, time.Time(t)), nil
}

func ParseTimeInLength(data string) (Time, error) {
	if value, err := time.ParseInLocation(defaultLayout[:len(data)], data, location); err != nil {
		return Time{}, err
	} else {
		return Time(value), nil
	}
}
func NewTimeFromCBCDeal(date string) (Time, error) {
	if value, err := time.ParseInLocation("20060102150405", date, location); err != nil {
		return Time{}, err
	} else {
		return Time(value), nil
	}
}
func (t Time) Before(another Time) bool {
	return time.Time(t).Before(time.Time(another))
}
func (t Time) After(another Time) bool {
	return time.Time(t).After(time.Time(another))
}

func (t Time) Sub(another Time) time.Duration {
	return time.Time(t).Sub(time.Time(another))
}

func (t Time) Add(duration time.Duration) Time {
	return Time(time.Time(t).Add(duration))
}
func (t Time) Unix() int64 {
	return time.Time(t).Unix()
}

func ConvertTime(data string) (interface{}, error) {
	d := &Time{}
	err := d.UnmarshalText([]byte(data))
	return *d, err
}
