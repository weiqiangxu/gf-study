// Copyright GoFrame Author(https://goframe.org). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gogf/gf.

package gvalid_test

import (
	"testing"
	"time"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/test/gtest"
)

var (
	ctx = gctx.New()
)

func Test_Check(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		rule := "abc:6,16"
		val1 := 0
		val2 := 7
		val3 := 20
		err1 := g.Validator().Data(val1).Rules(rule).Messages(nil).Run(ctx)
		err2 := g.Validator().Data(val2).Rules(rule).Messages(nil).Run(ctx)
		err3 := g.Validator().Data(val3).Rules(rule).Messages(nil).Run(ctx)
		t.Assert(err1, "InvalidRules: abc:6,16")
		t.Assert(err2, "InvalidRules: abc:6,16")
		t.Assert(err3, "InvalidRules: abc:6,16")
	})
}

func Test_Required(t *testing.T) {
	if m := g.Validator().Data("1").Rules("required").Messages(nil).Run(ctx); m != nil {
		t.Error(m)
	}
	if m := g.Validator().Data("").Rules("required").Messages(nil).Run(ctx); m == nil {
		t.Error(m)
	}
	if m := g.Validator().Data("").Assoc(map[string]interface{}{"id": 1, "age": 19}).Rules("required-if: id,1,age,18").Messages(nil).Run(ctx); m == nil {
		t.Error("Required校验失败")
	}
	if m := g.Validator().Data("").Assoc(map[string]interface{}{"id": 2, "age": 19}).Rules("required-if: id,1,age,18").Messages(nil).Run(ctx); m != nil {
		t.Error("Required校验失败")
	}
}

func Test_RequiredIf(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		rule := "required-if:id,1,age,18"
		t.AssertNE(g.Validator().Data("").Assoc(g.Map{"id": 1}).Rules(rule).Messages(nil).Run(ctx), nil)
		t.Assert(g.Validator().Data("").Assoc(g.Map{"id": 0}).Rules(rule).Messages(nil).Run(ctx), nil)
		t.AssertNE(g.Validator().Data("").Assoc(g.Map{"age": 18}).Rules(rule).Messages(nil).Run(ctx), nil)
		t.Assert(g.Validator().Data("").Assoc(g.Map{"age": 20}).Rules(rule).Messages(nil).Run(ctx), nil)
	})
}

func Test_RequiredUnless(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		rule := "required-unless:id,1,age,18"
		t.Assert(g.Validator().Data("").Assoc(g.Map{"id": 1}).Rules(rule).Messages(nil).Run(ctx), nil)
		t.AssertNE(g.Validator().Data("").Assoc(g.Map{"id": 0}).Rules(rule).Messages(nil).Run(ctx), nil)
		t.Assert(g.Validator().Data("").Assoc(g.Map{"age": 18}).Rules(rule).Messages(nil).Run(ctx), nil)
		t.AssertNE(g.Validator().Data("").Assoc(g.Map{"age": 20}).Rules(rule).Messages(nil).Run(ctx), nil)
	})
}

func Test_RequiredWith(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		rule := "required-with:id,name"
		val1 := ""
		params1 := g.Map{
			"age": 18,
		}
		params2 := g.Map{
			"id": 100,
		}
		params3 := g.Map{
			"id":   100,
			"name": "john",
		}
		err1 := g.Validator().Data(val1).Assoc(params1).Rules(rule).Messages(nil).Run(ctx)
		err2 := g.Validator().Data(val1).Assoc(params2).Rules(rule).Messages(nil).Run(ctx)
		err3 := g.Validator().Data(val1).Assoc(params3).Rules(rule).Messages(nil).Run(ctx)
		t.Assert(err1, nil)
		t.AssertNE(err2, nil)
		t.AssertNE(err3, nil)
	})
	// time.Time
	gtest.C(t, func(t *gtest.T) {
		rule := "required-with:id,time"
		val1 := ""
		params1 := g.Map{
			"age": 18,
		}
		params2 := g.Map{
			"id": 100,
		}
		params3 := g.Map{
			"time": time.Time{},
		}
		err1 := g.Validator().Data(val1).Assoc(params1).Rules(rule).Messages(nil).Run(ctx)
		err2 := g.Validator().Data(val1).Assoc(params2).Rules(rule).Messages(nil).Run(ctx)
		err3 := g.Validator().Data(val1).Assoc(params3).Rules(rule).Messages(nil).Run(ctx)
		t.Assert(err1, nil)
		t.AssertNE(err2, nil)
		t.Assert(err3, nil)
	})
	gtest.C(t, func(t *gtest.T) {
		rule := "required-with:id,time"
		val1 := ""
		params1 := g.Map{
			"age": 18,
		}
		params2 := g.Map{
			"id": 100,
		}
		params3 := g.Map{
			"time": time.Now(),
		}
		err1 := g.Validator().Data(val1).Assoc(params1).Rules(rule).Messages(nil).Run(ctx)
		err2 := g.Validator().Data(val1).Assoc(params2).Rules(rule).Messages(nil).Run(ctx)
		err3 := g.Validator().Data(val1).Assoc(params3).Rules(rule).Messages(nil).Run(ctx)
		t.Assert(err1, nil)
		t.AssertNE(err2, nil)
		t.AssertNE(err3, nil)
	})
	// gtime.Time
	gtest.C(t, func(t *gtest.T) {
		type UserApiSearch struct {
			Uid       int64       `json:"uid"`
			Nickname  string      `json:"nickname" v:"required-with:Uid"`
			StartTime *gtime.Time `json:"start_time" v:"required-with:EndTime"`
			EndTime   *gtime.Time `json:"end_time" v:"required-with:StartTime"`
		}
		data := UserApiSearch{
			StartTime: nil,
			EndTime:   nil,
		}
		t.Assert(g.Validator().Data(data).Run(ctx), nil)
	})
	gtest.C(t, func(t *gtest.T) {
		type UserApiSearch struct {
			Uid       int64       `json:"uid"`
			Nickname  string      `json:"nickname" v:"required-with:Uid"`
			StartTime *gtime.Time `json:"start_time" v:"required-with:EndTime"`
			EndTime   *gtime.Time `json:"end_time" v:"required-with:StartTime"`
		}
		data := UserApiSearch{
			StartTime: nil,
			EndTime:   gtime.Now(),
		}
		t.AssertNE(g.Validator().Data(data).Run(ctx), nil)
	})
}

func Test_RequiredWithAll(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		rule := "required-with-all:id,name"
		val1 := ""
		params1 := g.Map{
			"age": 18,
		}
		params2 := g.Map{
			"id": 100,
		}
		params3 := g.Map{
			"id":   100,
			"name": "john",
		}
		err1 := g.Validator().Data(val1).Assoc(params1).Rules(rule).Messages(nil).Run(ctx)
		err2 := g.Validator().Data(val1).Assoc(params2).Rules(rule).Messages(nil).Run(ctx)
		err3 := g.Validator().Data(val1).Assoc(params3).Rules(rule).Messages(nil).Run(ctx)
		t.Assert(err1, nil)
		t.Assert(err2, nil)
		t.AssertNE(err3, nil)
	})
}

func Test_RequiredWithOut(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		rule := "required-without:id,name"
		val1 := ""
		params1 := g.Map{
			"age": 18,
		}
		params2 := g.Map{
			"id": 100,
		}
		params3 := g.Map{
			"id":   100,
			"name": "john",
		}
		err1 := g.Validator().Data(val1).Assoc(params1).Rules(rule).Messages(nil).Run(ctx)
		err2 := g.Validator().Data(val1).Assoc(params2).Rules(rule).Messages(nil).Run(ctx)
		err3 := g.Validator().Data(val1).Assoc(params3).Rules(rule).Messages(nil).Run(ctx)
		t.AssertNE(err1, nil)
		t.AssertNE(err2, nil)
		t.Assert(err3, nil)
	})
}

func Test_RequiredWithOutAll(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		rule := "required-without-all:id,name"
		val1 := ""
		params1 := g.Map{
			"age": 18,
		}
		params2 := g.Map{
			"id": 100,
		}
		params3 := g.Map{
			"id":   100,
			"name": "john",
		}
		err1 := g.Validator().Data(val1).Assoc(params1).Rules(rule).Messages(nil).Run(ctx)
		err2 := g.Validator().Data(val1).Assoc(params2).Rules(rule).Messages(nil).Run(ctx)
		err3 := g.Validator().Data(val1).Assoc(params3).Rules(rule).Messages(nil).Run(ctx)
		t.AssertNE(err1, nil)
		t.Assert(err2, nil)
		t.Assert(err3, nil)
	})
}

func Test_Date(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		rule := "date"
		val1 := "2010"
		val2 := "201011"
		val3 := "20101101"
		val4 := "2010-11-01"
		val5 := "2010.11.01"
		val6 := "2010/11/01"
		val7 := "2010=11=01"
		val8 := "123"
		err1 := g.Validator().Data(val1).Rules(rule).Messages(nil).Run(ctx)
		err2 := g.Validator().Data(val2).Rules(rule).Messages(nil).Run(ctx)
		err3 := g.Validator().Data(val3).Rules(rule).Messages(nil).Run(ctx)
		err4 := g.Validator().Data(val4).Rules(rule).Messages(nil).Run(ctx)
		err5 := g.Validator().Data(val5).Rules(rule).Messages(nil).Run(ctx)
		err6 := g.Validator().Data(val6).Rules(rule).Messages(nil).Run(ctx)
		err7 := g.Validator().Data(val7).Rules(rule).Messages(nil).Run(ctx)
		err8 := g.Validator().Data(val8).Rules(rule).Messages(nil).Run(ctx)
		t.AssertNE(err1, nil)
		t.AssertNE(err2, nil)
		t.Assert(err3, nil)
		t.Assert(err4, nil)
		t.Assert(err5, nil)
		t.Assert(err6, nil)
		t.AssertNE(err7, nil)
		t.AssertNE(err8, nil)
	})
}

func Test_Datetime(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		m := g.MapStrBool{
			"2010":                false,
			"2010.11":             false,
			"2010-11-01":          false,
			"2010-11-01 12:00":    false,
			"2010-11-01 12:00:00": true,
			"2010.11.01 12:00:00": false,
		}
		for k, v := range m {
			err := g.Validator().Rules(`datetime`).Data(k).Run(ctx)
			if v {
				t.AssertNil(err)
			} else {
				t.AssertNE(err, nil)
			}
		}
	})
}

func Test_DateFormat(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		val1 := "2010"
		val2 := "201011"
		val3 := "2010.11"
		val4 := "201011-01"
		val5 := "2010~11~01"
		val6 := "2010-11~01"
		err1 := g.Validator().Data(val1).Rules("date-format:Y").Messages(nil).Run(ctx)
		err2 := g.Validator().Data(val2).Rules("date-format:Ym").Messages(nil).Run(ctx)
		err3 := g.Validator().Data(val3).Rules("date-format:Y.m").Messages(nil).Run(ctx)
		err4 := g.Validator().Data(val4).Rules("date-format:Ym-d").Messages(nil).Run(ctx)
		err5 := g.Validator().Data(val5).Rules("date-format:Y~m~d").Messages(nil).Run(ctx)
		err6 := g.Validator().Data(val6).Rules("date-format:Y~m~d").Messages(nil).Run(ctx)
		t.Assert(err1, nil)
		t.Assert(err2, nil)
		t.Assert(err3, nil)
		t.Assert(err4, nil)
		t.Assert(err5, nil)
		t.AssertNE(err6, nil)
	})
	gtest.C(t, func(t *gtest.T) {
		t1 := gtime.Now()
		t2 := time.Time{}
		err1 := g.Validator().Data(t1).Rules("date-format:Y").Run(ctx)
		err2 := g.Validator().Data(t2).Rules("date-format:Y").Run(ctx)
		t.Assert(err1, nil)
		t.AssertNE(err2, nil)
	})
}

func Test_Email(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		rule := "email"
		value1 := "m@johngcn"
		value2 := "m@www@johngcn"
		value3 := "m-m_m@mail.johng.cn"
		value4 := "m.m-m@johng.cn"
		err1 := g.Validator().Data(value1).Rules(rule).Messages(nil).Run(ctx)
		err2 := g.Validator().Data(value2).Rules(rule).Messages(nil).Run(ctx)
		err3 := g.Validator().Data(value3).Rules(rule).Messages(nil).Run(ctx)
		err4 := g.Validator().Data(value4).Rules(rule).Messages(nil).Run(ctx)
		t.AssertNE(err1, nil)
		t.AssertNE(err2, nil)
		t.Assert(err3, nil)
		t.Assert(err4, nil)
	})
}

func Test_Phone(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		err1 := g.Validator().Data("1361990897").Rules("phone").Messages(nil).Run(ctx)
		err2 := g.Validator().Data("13619908979").Rules("phone").Messages(nil).Run(ctx)
		err3 := g.Validator().Data("16719908979").Rules("phone").Messages(nil).Run(ctx)
		err4 := g.Validator().Data("19719908989").Rules("phone").Messages(nil).Run(ctx)
		t.AssertNE(err1.String(), nil)
		t.Assert(err2, nil)
		t.Assert(err3, nil)
		t.Assert(err4, nil)
	})
}

func Test_PhoneLoose(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		err1 := g.Validator().Data("13333333333").Rules("phone-loose").Messages(nil).Run(ctx)
		err2 := g.Validator().Data("15555555555").Rules("phone-loose").Messages(nil).Run(ctx)
		err3 := g.Validator().Data("16666666666").Rules("phone-loose").Messages(nil).Run(ctx)
		err4 := g.Validator().Data("23333333333").Rules("phone-loose").Messages(nil).Run(ctx)
		err5 := g.Validator().Data("1333333333").Rules("phone-loose").Messages(nil).Run(ctx)
		err6 := g.Validator().Data("10333333333").Rules("phone-loose").Messages(nil).Run(ctx)
		t.Assert(err1, nil)
		t.Assert(err2, nil)
		t.Assert(err3, nil)
		t.AssertNE(err4, nil)
		t.AssertNE(err5, nil)
		t.AssertNE(err6, nil)
	})
}

func Test_Telephone(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		rule := "telephone"
		val1 := "869265"
		val2 := "028-869265"
		val3 := "86292651"
		val4 := "028-8692651"
		val5 := "0830-8692651"
		err1 := g.Validator().Data(val1).Rules(rule).Messages(nil).Run(ctx)
		err2 := g.Validator().Data(val2).Rules(rule).Messages(nil).Run(ctx)
		err3 := g.Validator().Data(val3).Rules(rule).Messages(nil).Run(ctx)
		err4 := g.Validator().Data(val4).Rules(rule).Messages(nil).Run(ctx)
		err5 := g.Validator().Data(val5).Rules(rule).Messages(nil).Run(ctx)
		t.AssertNE(err1, nil)
		t.AssertNE(err2, nil)
		t.Assert(err3, nil)
		t.Assert(err4, nil)
		t.Assert(err5, nil)
	})
}

func Test_Passport(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		rule := "passport"
		val1 := "123456"
		val2 := "a12345-6"
		val3 := "aaaaa"
		val4 := "aaaaaa"
		val5 := "a123_456"
		err1 := g.Validator().Data(val1).Rules(rule).Messages(nil).Run(ctx)
		err2 := g.Validator().Data(val2).Rules(rule).Messages(nil).Run(ctx)
		err3 := g.Validator().Data(val3).Rules(rule).Messages(nil).Run(ctx)
		err4 := g.Validator().Data(val4).Rules(rule).Messages(nil).Run(ctx)
		err5 := g.Validator().Data(val5).Rules(rule).Messages(nil).Run(ctx)
		t.AssertNE(err1, nil)
		t.AssertNE(err2, nil)
		t.AssertNE(err3, nil)
		t.Assert(err4, nil)
		t.Assert(err5, nil)
	})
}

func Test_Password(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		rule := "password"
		val1 := "12345"
		val2 := "aaaaa"
		val3 := "a12345-6"
		val4 := ">,/;'[09-"
		val5 := "a123_456"
		err1 := g.Validator().Data(val1).Rules(rule).Messages(nil).Run(ctx)
		err2 := g.Validator().Data(val2).Rules(rule).Messages(nil).Run(ctx)
		err3 := g.Validator().Data(val3).Rules(rule).Messages(nil).Run(ctx)
		err4 := g.Validator().Data(val4).Rules(rule).Messages(nil).Run(ctx)
		err5 := g.Validator().Data(val5).Rules(rule).Messages(nil).Run(ctx)
		t.AssertNE(err1, nil)
		t.AssertNE(err2, nil)
		t.Assert(err3, nil)
		t.Assert(err4, nil)
		t.Assert(err5, nil)
	})
}

func Test_Password2(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		rule := "password2"
		val1 := "12345"
		val2 := "Naaaa"
		val3 := "a12345-6"
		val4 := ">,/;'[09-"
		val5 := "a123_456"
		val6 := "Nant1986"
		val7 := "Nant1986!"
		err1 := g.Validator().Data(val1).Rules(rule).Messages(nil).Run(ctx)
		err2 := g.Validator().Data(val2).Rules(rule).Messages(nil).Run(ctx)
		err3 := g.Validator().Data(val3).Rules(rule).Messages(nil).Run(ctx)
		err4 := g.Validator().Data(val4).Rules(rule).Messages(nil).Run(ctx)
		err5 := g.Validator().Data(val5).Rules(rule).Messages(nil).Run(ctx)
		err6 := g.Validator().Data(val6).Rules(rule).Messages(nil).Run(ctx)
		err7 := g.Validator().Data(val7).Rules(rule).Messages(nil).Run(ctx)
		t.AssertNE(err1, nil)
		t.AssertNE(err2, nil)
		t.AssertNE(err3, nil)
		t.AssertNE(err4, nil)
		t.AssertNE(err5, nil)
		t.Assert(err6, nil)
		t.Assert(err7, nil)
	})
}

func Test_Password3(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		rule := "password3"
		val1 := "12345"
		val2 := "Naaaa"
		val3 := "a12345-6"
		val4 := ">,/;'[09-"
		val5 := "a123_456"
		val6 := "Nant1986"
		val7 := "Nant1986!"
		err1 := g.Validator().Data(val1).Rules(rule).Messages(nil).Run(ctx)
		err2 := g.Validator().Data(val2).Rules(rule).Messages(nil).Run(ctx)
		err3 := g.Validator().Data(val3).Rules(rule).Messages(nil).Run(ctx)
		err4 := g.Validator().Data(val4).Rules(rule).Messages(nil).Run(ctx)
		err5 := g.Validator().Data(val5).Rules(rule).Messages(nil).Run(ctx)
		err6 := g.Validator().Data(val6).Rules(rule).Messages(nil).Run(ctx)
		err7 := g.Validator().Data(val7).Rules(rule).Messages(nil).Run(ctx)
		t.AssertNE(err1, nil)
		t.AssertNE(err2, nil)
		t.AssertNE(err3, nil)
		t.AssertNE(err4, nil)
		t.AssertNE(err5, nil)
		t.AssertNE(err6, nil)
		t.Assert(err7, nil)
	})
}

func Test_Postcode(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		rule := "postcode"
		val1 := "12345"
		val2 := "610036"
		err1 := g.Validator().Data(val1).Rules(rule).Messages(nil).Run(ctx)
		err2 := g.Validator().Data(val2).Rules(rule).Messages(nil).Run(ctx)
		t.AssertNE(err1, nil)
		t.Assert(err2, nil)
	})
}

func Test_ResidentId(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		rule := "resident-id"
		val1 := "11111111111111"
		val2 := "1111111111111111"
		val3 := "311128500121201"
		val4 := "510521198607185367"
		val5 := "51052119860718536x"
		err1 := g.Validator().Data(val1).Rules(rule).Messages(nil).Run(ctx)
		err2 := g.Validator().Data(val2).Rules(rule).Messages(nil).Run(ctx)
		err3 := g.Validator().Data(val3).Rules(rule).Messages(nil).Run(ctx)
		err4 := g.Validator().Data(val4).Rules(rule).Messages(nil).Run(ctx)
		err5 := g.Validator().Data(val5).Rules(rule).Messages(nil).Run(ctx)
		t.AssertNE(err1, nil)
		t.AssertNE(err2, nil)
		t.AssertNE(err3, nil)
		t.AssertNE(err4, nil)
		t.Assert(err5, nil)
	})
}

func Test_BankCard(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		rule := "bank-card"
		val1 := "6230514630000424470"
		val2 := "6230514630000424473"
		err1 := g.Validator().Data(val1).Rules(rule).Messages(nil).Run(ctx)
		err2 := g.Validator().Data(val2).Rules(rule).Messages(nil).Run(ctx)
		t.AssertNE(err1, nil)
		t.Assert(err2, nil)
	})
}

func Test_QQ(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		rule := "qq"
		val1 := "100"
		val2 := "1"
		val3 := "10000"
		val4 := "38996181"
		val5 := "389961817"
		err1 := g.Validator().Data(val1).Rules(rule).Messages(nil).Run(ctx)
		err2 := g.Validator().Data(val2).Rules(rule).Messages(nil).Run(ctx)
		err3 := g.Validator().Data(val3).Rules(rule).Messages(nil).Run(ctx)
		err4 := g.Validator().Data(val4).Rules(rule).Messages(nil).Run(ctx)
		err5 := g.Validator().Data(val5).Rules(rule).Messages(nil).Run(ctx)
		t.AssertNE(err1, nil)
		t.AssertNE(err2, nil)
		t.Assert(err3, nil)
		t.Assert(err4, nil)
		t.Assert(err5, nil)
	})
}

func Test_Ip(t *testing.T) {
	if m := g.Validator().Data("10.0.0.1").Rules("ip").Messages(nil).Run(ctx); m != nil {
		t.Error(m)
	}
	if m := g.Validator().Data("10.0.0.1").Rules("ipv4").Messages(nil).Run(ctx); m != nil {
		t.Error(m)
	}
	if m := g.Validator().Data("0.0.0.0").Rules("ipv4").Messages(nil).Run(ctx); m != nil {
		t.Error(m)
	}
	if m := g.Validator().Data("1920.0.0.0").Rules("ipv4").Messages(nil).Run(ctx); m == nil {
		t.Error("ipv4校验失败")
	}
	if m := g.Validator().Data("1920.0.0.0").Rules("ip").Messages(nil).Run(ctx); m == nil {
		t.Error("ipv4校验失败")
	}
	if m := g.Validator().Data("fe80::5484:7aff:fefe:9799").Rules("ipv6").Messages(nil).Run(ctx); m != nil {
		t.Error(m)
	}
	if m := g.Validator().Data("fe80::5484:7aff:fefe:9799123").Rules("ipv6").Messages(nil).Run(ctx); m == nil {
		t.Error(m)
	}
	if m := g.Validator().Data("fe80::5484:7aff:fefe:9799").Rules("ip").Messages(nil).Run(ctx); m != nil {
		t.Error(m)
	}
	if m := g.Validator().Data("fe80::5484:7aff:fefe:9799123").Rules("ip").Messages(nil).Run(ctx); m == nil {
		t.Error(m)
	}
}

func Test_IPv4(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		rule := "ipv4"
		val1 := "0.0.0"
		val2 := "0.0.0.0"
		val3 := "1.1.1.1"
		val4 := "255.255.255.0"
		val5 := "127.0.0.1"
		err1 := g.Validator().Data(val1).Rules(rule).Messages(nil).Run(ctx)
		err2 := g.Validator().Data(val2).Rules(rule).Messages(nil).Run(ctx)
		err3 := g.Validator().Data(val3).Rules(rule).Messages(nil).Run(ctx)
		err4 := g.Validator().Data(val4).Rules(rule).Messages(nil).Run(ctx)
		err5 := g.Validator().Data(val5).Rules(rule).Messages(nil).Run(ctx)
		t.AssertNE(err1, nil)
		t.Assert(err2, nil)
		t.Assert(err3, nil)
		t.Assert(err4, nil)
		t.Assert(err5, nil)
	})
}

func Test_IPv6(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		rule := "ipv6"
		val1 := "192.168.1.1"
		val2 := "CDCD:910A:2222:5498:8475:1111:3900:2020"
		val3 := "1030::C9B4:FF12:48AA:1A2B"
		val4 := "2000:0:0:0:0:0:0:1"
		val5 := "0000:0000:0000:0000:0000:ffff:c0a8:5909"
		err1 := g.Validator().Data(val1).Rules(rule).Messages(nil).Run(ctx)
		err2 := g.Validator().Data(val2).Rules(rule).Messages(nil).Run(ctx)
		err3 := g.Validator().Data(val3).Rules(rule).Messages(nil).Run(ctx)
		err4 := g.Validator().Data(val4).Rules(rule).Messages(nil).Run(ctx)
		err5 := g.Validator().Data(val5).Rules(rule).Messages(nil).Run(ctx)
		t.AssertNE(err1, nil)
		t.Assert(err2, nil)
		t.Assert(err3, nil)
		t.Assert(err4, nil)
		t.Assert(err5, nil)
	})
}

func Test_MAC(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		rule := "mac"
		val1 := "192.168.1.1"
		val2 := "44-45-53-54-00-00"
		val3 := "01:00:5e:00:00:00"
		err1 := g.Validator().Data(val1).Rules(rule).Messages(nil).Run(ctx)
		err2 := g.Validator().Data(val2).Rules(rule).Messages(nil).Run(ctx)
		err3 := g.Validator().Data(val3).Rules(rule).Messages(nil).Run(ctx)
		t.AssertNE(err1, nil)
		t.Assert(err2, nil)
		t.Assert(err3, nil)
	})
}

func Test_URL(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		rule := "url"
		val1 := "127.0.0.1"
		val2 := "https://www.baidu.com"
		val3 := "http://127.0.0.1"
		val4 := "file:///tmp/test.txt"
		err1 := g.Validator().Data(val1).Rules(rule).Messages(nil).Run(ctx)
		err2 := g.Validator().Data(val2).Rules(rule).Messages(nil).Run(ctx)
		err3 := g.Validator().Data(val3).Rules(rule).Messages(nil).Run(ctx)
		err4 := g.Validator().Data(val4).Rules(rule).Messages(nil).Run(ctx)
		t.AssertNE(err1, nil)
		t.Assert(err2, nil)
		t.Assert(err3, nil)
		t.Assert(err4, nil)
	})
}

func Test_Domain(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		m := g.MapStrBool{
			"localhost":     false,
			"baidu.com":     true,
			"www.baidu.com": true,
			"jn.np":         true,
			"www.jn.np":     true,
			"w.www.jn.np":   true,
			"127.0.0.1":     false,
			"www.360.com":   true,
			"www.360":       false,
			"360":           false,
			"my-gf":         false,
			"my-gf.com":     true,
			"my-gf.360.com": true,
		}
		var err error
		for k, v := range m {
			err = g.Validator().Data(k).Rules("domain").Messages(nil).Run(ctx)
			if v {
				// fmt.Println(k)
				t.Assert(err, nil)
			} else {
				// fmt.Println(k)
				t.AssertNE(err, nil)
			}
		}
	})
}

func Test_Length(t *testing.T) {
	rule := "length:6,16"
	if m := g.Validator().Data("123456").Rules(rule).Messages(nil).Run(ctx); m != nil {
		t.Error(m)
	}
	if m := g.Validator().Data("12345").Rules(rule).Messages(nil).Run(ctx); m == nil {
		t.Error("长度校验失败")
	}
}

func Test_MinLength(t *testing.T) {
	rule := "min-length:6"
	msgs := map[string]string{
		"min-length": "地址长度至少为{min}位",
	}
	if m := g.Validator().Data("123456").Rules(rule).Messages(nil).Run(ctx); m != nil {
		t.Error(m)
	}
	if m := g.Validator().Data("12345").Rules(rule).Messages(nil).Run(ctx); m == nil {
		t.Error("长度校验失败")
	}
	if m := g.Validator().Data("12345").Rules(rule).Messages(msgs).Run(ctx); m == nil {
		t.Error("长度校验失败")
	}

	rule2 := "min-length:abc"
	if m := g.Validator().Data("123456").Rules(rule2).Messages(nil).Run(ctx); m == nil {
		t.Error("长度校验失败")
	}
}

func Test_MaxLength(t *testing.T) {
	rule := "max-length:6"
	msgs := map[string]string{
		"max-length": "地址长度至大为{max}位",
	}
	if m := g.Validator().Data("12345").Rules(rule).Messages(nil).Run(ctx); m != nil {
		t.Error(m)
	}
	if m := g.Validator().Data("1234567").Rules(rule).Messages(nil).Run(ctx); m == nil {
		t.Error("长度校验失败")
	}
	if m := g.Validator().Data("1234567").Rules(rule).Messages(msgs).Run(ctx); m == nil {
		t.Error("长度校验失败")
	}

	rule2 := "max-length:abc"
	if m := g.Validator().Data("123456").Rules(rule2).Messages(nil).Run(ctx); m == nil {
		t.Error("长度校验失败")
	}
}

func Test_Size(t *testing.T) {
	rule := "size:5"
	if m := g.Validator().Data("12345").Rules(rule).Messages(nil).Run(ctx); m != nil {
		t.Error(m)
	}
	if m := g.Validator().Data("123456").Rules(rule).Messages(nil).Run(ctx); m == nil {
		t.Error("长度校验失败")
	}
}

func Test_Between(t *testing.T) {
	rule := "between:6.01, 10.01"
	if m := g.Validator().Data(10).Rules(rule).Messages(nil).Run(ctx); m != nil {
		t.Error(m)
	}
	if m := g.Validator().Data(10.02).Rules(rule).Messages(nil).Run(ctx); m == nil {
		t.Error("大小范围校验失败")
	}
	if m := g.Validator().Data("a").Rules(rule).Messages(nil).Run(ctx); m == nil {
		t.Error("大小范围校验失败")
	}
}

func Test_Min(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		rule := "min:100"
		val1 := "1"
		val2 := "99"
		val3 := "100"
		val4 := "1000"
		val5 := "a"
		err1 := g.Validator().Data(val1).Rules(rule).Messages(nil).Run(ctx)
		err2 := g.Validator().Data(val2).Rules(rule).Messages(nil).Run(ctx)
		err3 := g.Validator().Data(val3).Rules(rule).Messages(nil).Run(ctx)
		err4 := g.Validator().Data(val4).Rules(rule).Messages(nil).Run(ctx)
		err5 := g.Validator().Data(val5).Rules(rule).Messages(nil).Run(ctx)
		t.AssertNE(err1, nil)
		t.AssertNE(err2, nil)
		t.Assert(err3, nil)
		t.Assert(err4, nil)
		t.AssertNE(err5, nil)

		rule2 := "min:a"
		err6 := g.Validator().Data(val1).Rules(rule2).Messages(nil).Run(ctx)
		t.AssertNE(err6, nil)
	})
}

func Test_Max(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		rule := "max:100"
		val1 := "1"
		val2 := "99"
		val3 := "100"
		val4 := "1000"
		val5 := "a"
		err1 := g.Validator().Data(val1).Rules(rule).Messages(nil).Run(ctx)
		err2 := g.Validator().Data(val2).Rules(rule).Messages(nil).Run(ctx)
		err3 := g.Validator().Data(val3).Rules(rule).Messages(nil).Run(ctx)
		err4 := g.Validator().Data(val4).Rules(rule).Messages(nil).Run(ctx)
		err5 := g.Validator().Data(val5).Rules(rule).Messages(nil).Run(ctx)
		t.Assert(err1, nil)
		t.Assert(err2, nil)
		t.Assert(err3, nil)
		t.AssertNE(err4, nil)
		t.AssertNE(err5, nil)

		rule2 := "max:a"
		err6 := g.Validator().Data(val1).Rules(rule2).Messages(nil).Run(ctx)
		t.AssertNE(err6, nil)
	})
}

func Test_Json(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		rule := "json"
		val1 := ""
		val2 := "."
		val3 := "{}"
		val4 := "[]"
		val5 := "[1,2,3,4]"
		val6 := `{"list":[1,2,3,4]}`
		err1 := g.Validator().Data(val1).Rules(rule).Messages(nil).Run(ctx)
		err2 := g.Validator().Data(val2).Rules(rule).Messages(nil).Run(ctx)
		err3 := g.Validator().Data(val3).Rules(rule).Messages(nil).Run(ctx)
		err4 := g.Validator().Data(val4).Rules(rule).Messages(nil).Run(ctx)
		err5 := g.Validator().Data(val5).Rules(rule).Messages(nil).Run(ctx)
		err6 := g.Validator().Data(val6).Rules(rule).Messages(nil).Run(ctx)
		t.AssertNE(err1, nil)
		t.AssertNE(err2, nil)
		t.Assert(err3, nil)
		t.Assert(err4, nil)
		t.Assert(err5, nil)
		t.Assert(err6, nil)
	})
}

func Test_Integer(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		rule := "integer"
		val1 := ""
		val2 := "1.0"
		val3 := "001"
		val4 := "1"
		val5 := "100"
		val6 := `999999999`
		err1 := g.Validator().Data(val1).Rules(rule).Messages(nil).Run(ctx)
		err2 := g.Validator().Data(val2).Rules(rule).Messages(nil).Run(ctx)
		err3 := g.Validator().Data(val3).Rules(rule).Messages(nil).Run(ctx)
		err4 := g.Validator().Data(val4).Rules(rule).Messages(nil).Run(ctx)
		err5 := g.Validator().Data(val5).Rules(rule).Messages(nil).Run(ctx)
		err6 := g.Validator().Data(val6).Rules(rule).Messages(nil).Run(ctx)
		t.AssertNE(err1, nil)
		t.AssertNE(err2, nil)
		t.Assert(err3, nil)
		t.Assert(err4, nil)
		t.Assert(err5, nil)
		t.Assert(err6, nil)
	})
}

func Test_Float(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		rule := "float"
		val1 := ""
		val2 := "a"
		val3 := "1"
		val4 := "1.0"
		val5 := "1.1"
		val6 := `0.1`
		err1 := g.Validator().Data(val1).Rules(rule).Messages(nil).Run(ctx)
		err2 := g.Validator().Data(val2).Rules(rule).Messages(nil).Run(ctx)
		err3 := g.Validator().Data(val3).Rules(rule).Messages(nil).Run(ctx)
		err4 := g.Validator().Data(val4).Rules(rule).Messages(nil).Run(ctx)
		err5 := g.Validator().Data(val5).Rules(rule).Messages(nil).Run(ctx)
		err6 := g.Validator().Data(val6).Rules(rule).Messages(nil).Run(ctx)
		t.AssertNE(err1, nil)
		t.AssertNE(err2, nil)
		t.Assert(err3, nil)
		t.Assert(err4, nil)
		t.Assert(err5, nil)
		t.Assert(err6, nil)
	})
}

func Test_Boolean(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		rule := "boolean"
		val1 := "a"
		val2 := "-"
		val3 := ""
		val4 := "1"
		val5 := "true"
		val6 := `off`
		err1 := g.Validator().Data(val1).Rules(rule).Messages(nil).Run(ctx)
		err2 := g.Validator().Data(val2).Rules(rule).Messages(nil).Run(ctx)
		err3 := g.Validator().Data(val3).Rules(rule).Messages(nil).Run(ctx)
		err4 := g.Validator().Data(val4).Rules(rule).Messages(nil).Run(ctx)
		err5 := g.Validator().Data(val5).Rules(rule).Messages(nil).Run(ctx)
		err6 := g.Validator().Data(val6).Rules(rule).Messages(nil).Run(ctx)
		t.AssertNE(err1, nil)
		t.AssertNE(err2, nil)
		t.Assert(err3, nil)
		t.Assert(err4, nil)
		t.Assert(err5, nil)
		t.Assert(err6, nil)
	})
}

func Test_Same(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		rule := "same:id"
		val1 := "100"
		params1 := g.Map{
			"age": 18,
		}
		params2 := g.Map{
			"id": 100,
		}
		params3 := g.Map{
			"id":   100,
			"name": "john",
		}
		err1 := g.Validator().Data(val1).Assoc(params1).Rules(rule).Messages(nil).Run(ctx)
		err2 := g.Validator().Data(val1).Assoc(params2).Rules(rule).Messages(nil).Run(ctx)
		err3 := g.Validator().Data(val1).Assoc(params3).Rules(rule).Messages(nil).Run(ctx)
		t.AssertNE(err1, nil)
		t.Assert(err2, nil)
		t.Assert(err3, nil)
	})
}

func Test_Different(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		rule := "different:id"
		val1 := "100"
		params1 := g.Map{
			"age": 18,
		}
		params2 := g.Map{
			"id": 100,
		}
		params3 := g.Map{
			"id":   100,
			"name": "john",
		}
		err1 := g.Validator().Data(val1).Assoc(params1).Rules(rule).Messages(nil).Run(ctx)
		err2 := g.Validator().Data(val1).Assoc(params2).Rules(rule).Messages(nil).Run(ctx)
		err3 := g.Validator().Data(val1).Assoc(params3).Rules(rule).Messages(nil).Run(ctx)
		t.Assert(err1, nil)
		t.AssertNE(err2, nil)
		t.AssertNE(err3, nil)
	})
}

func Test_In(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		rule := "in:100,200"
		val1 := ""
		val2 := "1"
		val3 := "100"
		val4 := "200"
		err1 := g.Validator().Data(val1).Rules(rule).Messages(nil).Run(ctx)
		err2 := g.Validator().Data(val2).Rules(rule).Messages(nil).Run(ctx)
		err3 := g.Validator().Data(val3).Rules(rule).Messages(nil).Run(ctx)
		err4 := g.Validator().Data(val4).Rules(rule).Messages(nil).Run(ctx)
		t.AssertNE(err1, nil)
		t.AssertNE(err2, nil)
		t.Assert(err3, nil)
		t.Assert(err4, nil)
	})
}

func Test_NotIn(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		rule := "not-in:100"
		val1 := ""
		val2 := "1"
		val3 := "100"
		val4 := "200"
		err1 := g.Validator().Data(val1).Rules(rule).Messages(nil).Run(ctx)
		err2 := g.Validator().Data(val2).Rules(rule).Messages(nil).Run(ctx)
		err3 := g.Validator().Data(val3).Rules(rule).Messages(nil).Run(ctx)
		err4 := g.Validator().Data(val4).Rules(rule).Messages(nil).Run(ctx)
		t.Assert(err1, nil)
		t.Assert(err2, nil)
		t.AssertNE(err3, nil)
		t.Assert(err4, nil)
	})
	gtest.C(t, func(t *gtest.T) {
		rule := "not-in:100,200"
		val1 := ""
		val2 := "1"
		val3 := "100"
		val4 := "200"
		err1 := g.Validator().Data(val1).Rules(rule).Messages(nil).Run(ctx)
		err2 := g.Validator().Data(val2).Rules(rule).Messages(nil).Run(ctx)
		err3 := g.Validator().Data(val3).Rules(rule).Messages(nil).Run(ctx)
		err4 := g.Validator().Data(val4).Rules(rule).Messages(nil).Run(ctx)
		t.Assert(err1, nil)
		t.Assert(err2, nil)
		t.AssertNE(err3, nil)
		t.AssertNE(err4, nil)
	})
}

func Test_Regex1(t *testing.T) {
	rule := `regex:\d{6}|\D{6}|length:6,16`
	if m := g.Validator().Data("123456").Rules(rule).Messages(nil).Run(ctx); m != nil {
		t.Error(m)
	}
	if m := g.Validator().Data("abcde6").Rules(rule).Messages(nil).Run(ctx); m == nil {
		t.Error("校验失败")
	}
}

func Test_Regex2(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		rule := `required|min-length:6|regex:^data:image\/(jpeg|png);base64,`
		str1 := ""
		str2 := "data"
		str3 := "data:image/jpeg;base64,/9jrbattq22r"
		err1 := g.Validator().Data(str1).Rules(rule).Messages(nil).Run(ctx)
		err2 := g.Validator().Data(str2).Rules(rule).Messages(nil).Run(ctx)
		err3 := g.Validator().Data(str3).Rules(rule).Messages(nil).Run(ctx)
		t.AssertNE(err1, nil)
		t.AssertNE(err2, nil)
		t.Assert(err3, nil)

		t.AssertNE(err1.Map()["required"], nil)
		t.AssertNE(err2.Map()["min-length"], nil)
	})
}

// issue: https://github.com/gogf/gf/issues/1077
func Test_InternalError_String(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		type a struct {
			Name string `v:"hh"`
		}
		aa := a{Name: "2"}
		err := g.Validator().Data(&aa).Rules(nil).Run(ctx)

		t.Assert(err.String(), "InvalidRules: hh")
		t.Assert(err.Strings(), g.Slice{"InvalidRules: hh"})
		t.Assert(err.FirstError(), "InvalidRules: hh")
		t.Assert(gerror.Current(err), "InvalidRules: hh")
	})
}

func Test_Code(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		err := g.Validator().Rules("required").Data("").Run(ctx)
		t.AssertNE(err, nil)
		t.Assert(gerror.Code(err), gcode.CodeValidationFailed)
	})

	gtest.C(t, func(t *gtest.T) {
		err := g.Validator().Rules("none-exist-rule").Data("").Run(ctx)
		t.AssertNE(err, nil)
		t.Assert(gerror.Code(err), gcode.CodeInternalError)
	})
}

func Test_Bail(t *testing.T) {
	// check value with no bail
	gtest.C(t, func(t *gtest.T) {
		err := g.Validator().
			Rules("required|min:1|between:1,100").
			Messages("|min number is 1|size is between 1 and 100").
			Data(-1).Run(ctx)
		t.AssertNE(err, nil)
		t.Assert(err.Error(), "min number is 1; size is between 1 and 100")
	})

	// check value with bail
	gtest.C(t, func(t *gtest.T) {
		err := g.Validator().
			Rules("bail|required|min:1|between:1,100").
			Messages("||min number is 1|size is between 1 and 100").
			Data(-1).Run(ctx)
		t.AssertNE(err, nil)
		t.Assert(err.Error(), "min number is 1")
	})

	// struct with no bail
	gtest.C(t, func(t *gtest.T) {
		type Params struct {
			Page int `v:"required|min:1"`
			Size int `v:"required|min:1|between:1,100 # |min number is 1|size is between 1 and 100"`
		}
		obj := &Params{
			Page: 1,
			Size: -1,
		}
		err := g.Validator().Data(obj).Run(ctx)
		t.AssertNE(err, nil)
		t.Assert(err.Error(), "min number is 1; size is between 1 and 100")
	})
	// struct with bail
	gtest.C(t, func(t *gtest.T) {
		type Params struct {
			Page int `v:"required|min:1"`
			Size int `v:"bail|required|min:1|between:1,100 # ||min number is 1|size is between 1 and 100"`
		}
		obj := &Params{
			Page: 1,
			Size: -1,
		}
		err := g.Validator().Data(obj).Run(ctx)
		t.AssertNE(err, nil)
		t.Assert(err.Error(), "min number is 1")
	})
}
