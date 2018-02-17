package rady

import (
	"github.com/tidwall/gjson"
	"os"
	"reflect"
	"time"
)

type (
	/*
		CtrlBean contains value and tag of a controller
	*/
	CtrlBean struct {
		Name  string
		Value reflect.Value
		Tag   reflect.StructTag
	}

	/*
		MdWareBean contains value and tag of a middleware
	*/
	MdWareBean struct {
		Name  string
		Value reflect.Value
		Tag   reflect.StructTag
	}

	// Method contains value, param list and name of a 'BeanMethod'
	Method struct {
		Value    reflect.Value
		Ins      []reflect.Type
		Name     string
		OutValue reflect.Value
		InValues []reflect.Value
	}

	/*
		ValueBean contains value from config file parsed by 'gjson'

		ValueMap is different types the value converted to

		ParamSlice is the param list contain this value
	*/
	ValueBean struct {
		Value     gjson.Result
		ValueMap  map[reflect.Type]reflect.Value
		MethodSet map[*Method]bool
		Key       string
		Default   gjson.Result
	}

	// Bean contains the value and tag of a type
	Bean struct {
		Tag   reflect.StructTag
		Value reflect.Value
	}
)

func (m *Method) LoadIns(app *Application) {
	for _, inType := range m.Ins {
		if ConfirmSameTypeInMap(app.BeanMap, inType) {
			if len(app.BeanMap[inType]) > 1 {
				app.Logger.Critical("There are more than one %s, please named it.", inType)
				os.Exit(1)
			}
			for _, bean := range app.BeanMap[inType] {
				m.InValues = append(m.InValues, bean.Value)
			}
		} else {
			newValue := reflect.New(inType.Elem()).Elem()
			app.load(inType, newValue, GetTagFromName(""))
			m.InValues = append(m.InValues, newValue)
		}
	}
}

func (m *Method) Call(app *Application) {
	params := make([]reflect.Value, 0)
	for _, value := range m.InValues {
		params = append(params, value.Addr())
	}
	result := m.Value.Call(params)
	if len(result) != 1 {
		app.Logger.Error("Result of %s is not a Component!!!", m.Name)
		os.Exit(1)
	}
	app.Logger.Debug("Result of %s set %s", m.Name, result[0].Elem())
	m.OutValue.Set(result[0].Elem())
}

func (v *ValueBean) Reload(a *Application) {
	newResult := gjson.Get(a.ConfigFile, v.Key)
	if newResult != v.Value {
		if !newResult.Exists() {
			a.Logger.Info("Key %s doesn't exist, use default value %s", v.Key, v.Default.String())
			newResult = v.Default
		}
		v.Value = newResult
		a.Logger.Debug("Reset Value '%s' to %s", v.Key, v.Value.String())
		v.resetValue()
		v.recallFactory(a)
	}
}

func (v *ValueBean) resetValue() {
	for Type, Value := range v.ValueMap {
		switch Type {
		case IntType:
			Value.SetInt(v.Value.Int())
		case UintType:
			Value.SetUint(v.Value.Uint())
		case FloatType:
			Value.SetFloat(v.Value.Float())
		case StringType:
			Value.SetString(v.Value.String())
		case BoolType:
			Value.SetBool(v.Value.Bool())
		case TimeType:
			Value.Set(reflect.ValueOf(v.Value.Time()))
		case ArrayType:
			Value.Set(reflect.ValueOf(v.Value.Array()))
		case MapType:
			Value.Set(reflect.ValueOf(v.Value.Map()))
		case ArrayIntType:
			result := v.Value.Array()
			length := len(result)
			realResult := make([]int64, length)
			for i := 0; i < length; i++ {
				realResult[i] = result[i].Int()
			}
			Value.Set(reflect.ValueOf(realResult))
		case ArrayUintType:
			result := v.Value.Array()
			length := len(result)
			realResult := make([]uint64, length)
			for i := 0; i < length; i++ {
				realResult[i] = result[i].Uint()
			}
			Value.Set(reflect.ValueOf(realResult))
		case ArrayFloatType:
			result := v.Value.Array()
			length := len(result)
			realResult := make([]float64, length)
			for i := 0; i < length; i++ {
				realResult[i] = result[i].Float()
			}
			Value.Set(reflect.ValueOf(realResult))
		case ArrayBoolType:
			result := v.Value.Array()
			length := len(result)
			realResult := make([]bool, length)
			for i := 0; i < length; i++ {
				realResult[i] = result[i].Bool()
			}
			Value.Set(reflect.ValueOf(realResult))
		case ArrayStringType:
			result := v.Value.Array()
			length := len(result)
			realResult := make([]string, length)
			for i := 0; i < length; i++ {
				realResult[i] = result[i].String()
			}
			Value.Set(reflect.ValueOf(realResult))
		case ArrayTimeType:
			result := v.Value.Array()
			length := len(result)
			realResult := make([]time.Time, length)
			for i := 0; i < length; i++ {
				realResult[i] = result[i].Time()
			}
			Value.Set(reflect.ValueOf(realResult))
		}
	}
}

func (v *ValueBean) recallFactory(a *Application) {
	for Method := range v.MethodSet {
		if _, ok := a.FactoryToRecall[Method]; !ok {
			a.FactoryToRecall[Method] = true
		}
	}
}

func (v *ValueBean) setValue(IsPtr bool, value reflect.Value, Type reflect.Type) bool {
	confValue, ok := v.ValueMap[Type]
	if ok {
		if IsPtr {
			value.Set(confValue.Addr())
		} else {
			value.Set(confValue)
		}
		return true
	}
	return false
}

func (v *ValueBean) SetValue(value reflect.Value, Type reflect.Type) bool {
	IsPtr := Type.Kind() == reflect.Ptr
	if IsPtr {
		Type = Type.Elem()
	}

	if v.setValue(IsPtr, value, Type) {
		return true
	}
	switch Type {
	case IntType:
		result := v.Value.Int()
		v.ValueMap[Type] = reflect.ValueOf(&result).Elem()
	case UintType:
		result := v.Value.Uint()
		v.ValueMap[Type] = reflect.ValueOf(&result).Elem()
	case FloatType:
		result := v.Value.Float()
		v.ValueMap[Type] = reflect.ValueOf(&result).Elem()
	case StringType:
		result := v.Value.String()
		v.ValueMap[Type] = reflect.ValueOf(&result).Elem()
	case BoolType:
		result := v.Value.Bool()
		v.ValueMap[Type] = reflect.ValueOf(&result).Elem()
	case TimeType:
		result := v.Value.Time()
		v.ValueMap[Type] = reflect.ValueOf(&result).Elem()
	case ArrayType:
		result := v.Value.Array()
		v.ValueMap[Type] = reflect.ValueOf(&result).Elem()
	case MapType:
		result := v.Value.Map()
		v.ValueMap[Type] = reflect.ValueOf(&result).Elem()
	case ArrayIntType:
		result := v.Value.Array()
		length := len(result)
		realResult := make([]int64, length)
		for i := 0; i < length; i++ {
			realResult[i] = result[i].Int()
		}
		v.ValueMap[Type] = reflect.ValueOf(&realResult).Elem()
	case ArrayUintType:
		result := v.Value.Array()
		length := len(result)
		realResult := make([]uint64, length)
		for i := 0; i < length; i++ {
			realResult[i] = result[i].Uint()
		}
		v.ValueMap[Type] = reflect.ValueOf(&realResult).Elem()
	case ArrayFloatType:
		result := v.Value.Array()
		length := len(result)
		realResult := make([]float64, length)
		for i := 0; i < length; i++ {
			realResult[i] = result[i].Float()
		}
		v.ValueMap[Type] = reflect.ValueOf(&realResult).Elem()
	case ArrayBoolType:
		result := v.Value.Array()
		length := len(result)
		realResult := make([]bool, length)
		for i := 0; i < length; i++ {
			realResult[i] = result[i].Bool()
		}
		v.ValueMap[Type] = reflect.ValueOf(&realResult).Elem()
	case ArrayStringType:
		result := v.Value.Array()
		length := len(result)
		realResult := make([]string, length)
		for i := 0; i < length; i++ {
			realResult[i] = result[i].String()
		}
		v.ValueMap[Type] = reflect.ValueOf(&realResult).Elem()
	case ArrayTimeType:
		result := v.Value.Array()
		length := len(result)
		realResult := make([]time.Time, length)
		for i := 0; i < length; i++ {
			realResult[i] = result[i].Time()
		}
		v.ValueMap[Type] = reflect.ValueOf(&realResult).Elem()
	}
	return v.setValue(IsPtr, value, Type)
}

/*
NewBean is factory function of Bean
*/
func NewBean(Value reflect.Value, Tag reflect.StructTag) *Bean {
	return &Bean{
		Tag:   Tag,
		Value: Value,
	}
}

/*
NewBeanMethod is factory function of Method
*/
func NewBeanMethod(Value reflect.Value, Name string) *Method {
	return &Method{
		Value:    Value,
		Ins:      make([]reflect.Type, 0),
		InValues: make([]reflect.Value, 0),
		Name:     Name,
	}
}

///*
//NewParamBean is factory function of ParamBean
// */
//func NewParamBean(Value reflect.Value, MethodBean *Method) *ParamBean {
//	return &ParamBean{
//		Value:      Value,
//		MethodBean: MethodBean,
//	}
//}

/*
NewValueBean is factory function of ValueBean
*/
func NewValueBean(Value gjson.Result, key string, defaultValue gjson.Result) *ValueBean {
	return &ValueBean{
		Value:     Value,
		ValueMap:  make(map[reflect.Type]reflect.Value),
		MethodSet: make(map[*Method]bool),
		Key:       key,
		Default:   defaultValue,
	}
}

/*
NewCtrlBean is factory function of CtrlBean
*/
func NewCtrlBean(Value reflect.Value, Tag reflect.StructTag, Name string) *CtrlBean {
	return &CtrlBean{
		Name:  Name,
		Tag:   Tag,
		Value: Value,
	}
}

/*
NewMdWareBean is factory function of MdwareBean
*/
func NewMdWareBean(Value reflect.Value, Tag reflect.StructTag, Name string) *MdWareBean {
	return &MdWareBean{
		Name:  Name,
		Tag:   Tag,
		Value: Value,
	}
}
