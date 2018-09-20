package core

import "regexp"

// Validation context manages data validation and error message.
type Validation struct {
}

func (v *Validation) apply(chk Validator, obj interface{}) bool {
	if chk.IsSatisfied(obj) == false {
		return false
	} else {
		return true
	}
}

// Required tests that the argument is non-nil and non-empty (if string or list)
func (v *Validation) Required(obj interface{}) bool {
	return v.apply(Required{}, obj)
}

func (v *Validation) Min(n int, min int) bool {
	return v.MinFloat(float64(n), float64(min))
}

func (v *Validation) MinFloat(n float64, min float64) bool {
	return v.apply(Min{min}, n)
}

func (v *Validation) Max(n int, max int) bool {
	return v.MaxFloat(float64(n), float64(max))
}

func (v *Validation) MaxFloat(n float64, max float64) bool {
	return v.apply(Max{max}, n)
}

func (v *Validation) Range(n, min, max int) bool {
	return v.RangeFloat(float64(n), float64(min), float64(max))
}

func (v *Validation) Range64(n, min, max int64) bool {
	return v.RangeFloat(float64(n), float64(min), float64(max))
}

func (v *Validation) RangeFloat(n, min, max float64) bool {
	return v.apply(Range{Min{min}, Max{max}}, n)
}

func (v *Validation) MinSize(obj interface{}, min int) bool {
	return v.apply(MinSize{min}, obj)
}

func (v *Validation) MaxSize(obj interface{}, max int) bool {
	return v.apply(MaxSize{max}, obj)
}

func (v *Validation) Length(obj interface{}, n int) bool {
	return v.apply(Length{n}, obj)
}

func (v *Validation) Match(str string, regex *regexp.Regexp) bool {
	return v.apply(Match{regex}, str)
}

func (v *Validation) Email(str string) bool {
	return v.apply(Email{Match{emailPattern}}, str)
}

func (v *Validation) IPAddr(str string, cktype ...int) bool {
	return v.apply(IPAddr{cktype}, str)
}

func (v *Validation) MacAddr(str string) bool {
	return v.apply(IPAddr{}, str)
}

func (v *Validation) Domain(str string) bool {
	return v.apply(Domain{}, str)
}

func (v *Validation) URL(str string) bool {
	return v.apply(URL{}, str)
}

func (v *Validation) PureText(str string, m int) bool {
	return v.apply(PureText{m}, str)
}

func (v *Validation) FilePath(str string, m int) bool {
	return v.apply(FilePath{m}, str)
}
