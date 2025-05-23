package pipeline_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	fpparse "github.com/turbot/flowpipe/internal/parse"
)

func TestParamValidation(t *testing.T) {
	assert := assert.New(t)

	pipelines, _, err := fpparse.LoadPipelines(context.TODO(), "./pipelines/param_validation.fp")
	assert.Nil(err, "error found")

	validateMyParam := pipelines["local.pipeline.validate_my_param"]
	if validateMyParam == nil {
		assert.Fail("validate_my_param pipeline not found")
		return
	}

	stringValid := map[string]interface{}{
		"my_token": "abc",
	}

	assert.Equal(0, len(fpparse.ValidateParams(validateMyParam, stringValid, nil)))

	stringInvalid := map[string]interface{}{
		"my_token": 123,
	}

	errs := fpparse.ValidateParams(validateMyParam, stringInvalid, nil)
	assert.Equal(1, len(errs))
	assert.Equal("Bad Request: invalid data type for parameter 'my_token' wanted string but received int", errs[0].Error())

	invalidParam := map[string]interface{}{
		"invalid": "foo",
	}
	errs = fpparse.ValidateParams(validateMyParam, invalidParam, nil)
	assert.Equal(1, len(errs))
	assert.Equal("Bad Request: unknown parameter specified 'invalid'", errs[0].Error())

	allValid := map[string]interface{}{
		"my_token":      "123",
		"my_number":     123,
		"my_number_two": 123.45,
		"my_bool":       true,
	}

	errs = fpparse.ValidateParams(validateMyParam, allValid, nil)
	assert.Equal(0, len(errs))

	invalidNum := map[string]interface{}{
		"my_token":      "123",
		"my_number":     123,
		"my_number_two": "123.45",
		"my_bool":       true,
	}
	errs = fpparse.ValidateParams(validateMyParam, invalidNum, nil)
	assert.Equal(1, len(errs))
	assert.Equal("Bad Request: invalid data type for parameter 'my_number_two' wanted number but received string", errs[0].Error())

	moreThanOneInvalids := map[string]interface{}{
		"my_token":      "123",
		"my_number":     "a",
		"my_number_two": "123.45",
		"my_bool":       "true",
	}
	errs = fpparse.ValidateParams(validateMyParam, moreThanOneInvalids, nil)
	assert.Equal(3, len(errs))

	expectedErrors := []string{
		"Bad Request: invalid data type for parameter 'my_number' wanted number but received string",
		"Bad Request: invalid data type for parameter 'my_bool' wanted bool but received string",
		"Bad Request: invalid data type for parameter 'my_number_two' wanted number but received string",
	}

	actualErrors := []string{}
	for _, err := range errs {
		actualErrors = append(actualErrors, err.Error())
	}

	less := func(a, b string) bool { return a < b }
	equalIgnoreOrder := cmp.Equal(expectedErrors, actualErrors, cmpopts.SortSlices(less))
	assert.True(equalIgnoreOrder, "expected errors do not match")

	paramList := map[string]interface{}{
		"list_string":       []string{"foo", "bar"},
		"list_number":       []float64{1.23, 4.56},
		"list_number_two":   []float32{1.23, 4.56},
		"list_number_three": []int64{1, 4},
	}

	errs = fpparse.ValidateParams(validateMyParam, paramList, nil)
	assert.Equal(0, len(errs))

	paramListMoreNumberType := map[string]interface{}{
		"list_string":       []string{"foo", "bar"},
		"list_number":       []int{1, 4, 5, 6},
		"list_number_two":   []uint{1, 4, 5},
		"list_number_three": []int16{1, 4},
	}

	errs = fpparse.ValidateParams(validateMyParam, paramListMoreNumberType, nil)
	assert.Equal(0, len(errs))

	paramListAsInterface := map[string]interface{}{
		"list_string":       []interface{}{"foo", "bar"},
		"list_number":       []interface{}{1, 4, -4, 6},
		"list_number_two":   []interface{}{1, 4, 5.5}, // mixed float and int
		"list_number_three": []interface{}{1, 4},
	}

	errs = fpparse.ValidateParams(validateMyParam, paramListAsInterface, nil)
	assert.Equal(0, len(errs))

	paramNotList := map[string]interface{}{
		"list_string":     "foo",
		"list_number":     1,
		"list_number_two": 1.23,
	}

	errs = fpparse.ValidateParams(validateMyParam, paramNotList, nil)
	assert.Equal(3, len(errs))

	expectedErrors = []string{
		"Bad Request: invalid data type for parameter 'list_string' wanted list of string but received string",
		"Bad Request: invalid data type for parameter 'list_number' wanted list of number but received int",
		"Bad Request: invalid data type for parameter 'list_number_two' wanted list of number but received float64",
	}

	actualErrors = []string{}
	for _, err := range errs {
		actualErrors = append(actualErrors, err.Error())
	}

	equalIgnoreOrder = cmp.Equal(expectedErrors, actualErrors, cmpopts.SortSlices(less))
	assert.True(equalIgnoreOrder, "expected errors do not match")

	listStringInvalid := map[string]interface{}{
		"list_string": []interface{}{"foo", 1, "two"},
	}
	errs = fpparse.ValidateParams(validateMyParam, listStringInvalid, nil)
	assert.Equal(1, len(errs))
	assert.Equal("Bad Request: invalid data type for parameter 'list_string' wanted list of string but received []interface {}", errs[0].Error())

	listAny := map[string]interface{}{
		"list_any":       []interface{}{"foo", 1, 1.23, true},
		"list_any_two":   []interface{}{"foo", "bar", "baz"},
		"list_any_three": []interface{}{1, 2, 3},
	}

	errs = fpparse.ValidateParams(validateMyParam, listAny, nil)
	assert.Equal(0, len(errs))

	setString := map[string]interface{}{
		"set_string": []string{"foo", "bar", "baz"},
	}
	errs = fpparse.ValidateParams(validateMyParam, setString, nil)
	assert.Equal(0, len(errs))

	setNumber := map[string]interface{}{
		"set_number": []int{1, 2, 3},
	}
	errs = fpparse.ValidateParams(validateMyParam, setNumber, nil)
	assert.Equal(0, len(errs))

	setBool := map[string]interface{}{
		"set_bool": []bool{false, true, true},
	}
	errs = fpparse.ValidateParams(validateMyParam, setBool, nil)
	assert.Equal(0, len(errs))

	stringMap := map[string]interface{}{
		"map_of_string": map[string]string{
			"foo": "bar",
			"baz": "qux",
		},
	}

	errs = fpparse.ValidateParams(validateMyParam, stringMap, nil)
	assert.Equal(0, len(errs))

	stringMapGeneric := map[string]interface{}{
		"map_of_string": map[string]interface{}{
			"foo": "bar",
			"baz": "qux",
		},
	}
	errs = fpparse.ValidateParams(validateMyParam, stringMapGeneric, nil)
	assert.Equal(0, len(errs))

	stringMapGenericInvalid := map[string]interface{}{
		"map_of_string": map[string]interface{}{
			"foo": "bar",
			"baz": 123,
		},
	}
	errs = fpparse.ValidateParams(validateMyParam, stringMapGenericInvalid, nil)
	assert.Equal(1, len(errs))
	assert.Equal("Bad Request: invalid data type for parameter 'map_of_string' wanted map of string but received map[string]interface {}", errs[0].Error())

	numberMap := map[string]interface{}{
		"map_of_number": map[string]float64{
			"foo": 1.23,
			"baz": 4.56,
		},
	}
	errs = fpparse.ValidateParams(validateMyParam, numberMap, nil)
	assert.Equal(0, len(errs))

	numberMapInvalid := map[string]interface{}{
		"map_of_number": map[string]interface{}{
			"foo": "1.23",
			"baz": "4.56",
		},
	}
	errs = fpparse.ValidateParams(validateMyParam, numberMapInvalid, nil)
	assert.Equal(1, len(errs))

	numberMapInvalid = map[string]interface{}{
		"map_of_number": map[string]string{
			"foo": "1.23",
			"baz": "4.56",
		},
		"map_of_number_two": 4,
	}
	errs = fpparse.ValidateParams(validateMyParam, numberMapInvalid, nil)
	assert.Equal(2, len(errs))

	numberMap = map[string]interface{}{
		"map_of_number": map[string]float64{
			"foo": 1.23,
			"baz": 4.56,
		},
		"map_of_number_two": map[string]int{
			"foo": 1,
			"baz": 4,
		},
	}
	errs = fpparse.ValidateParams(validateMyParam, numberMap, nil)
	assert.Equal(0, len(errs))

	numberMap = map[string]interface{}{
		"map_of_number": map[string]int16{
			"foo": 1,
			"baz": 4,
		},
		"map_of_number_two": map[string]uint32{
			"foo": 1,
			"baz": 4,
		},
	}
	errs = fpparse.ValidateParams(validateMyParam, numberMap, nil)
	assert.Equal(0, len(errs))

	anyMap := map[string]interface{}{
		"map_of_string": map[string]interface{}{
			"foo": "bar",
			"baz": "123",
		},
		"map_of_any": map[string]int16{
			"foo": 1,
			"baz": 4,
		},
		"map_of_any_two": map[string]string{
			"foo": "1",
			"baz": "4",
		},
		"map_of_any_three": map[string]interface{}{
			"foo": 1,
			"baz": "4",
		},
	}
	errs = fpparse.ValidateParams(validateMyParam, anyMap, nil)
	assert.Equal(0, len(errs))

	anyMapInvalid := map[string]interface{}{
		"map_of_any":       []interface{}{1, 2, 3},
		"map_of_any_two":   []interface{}{"foo", 2, 3},
		"map_of_any_three": 23,
	}
	errs = fpparse.ValidateParams(validateMyParam, anyMapInvalid, nil)
	assert.Equal(3, len(errs))

}

func TestParamCoerce(t *testing.T) {
	assert := assert.New(t)

	pipelines, _, err := fpparse.LoadPipelines(context.TODO(), "./pipelines/param_validation.fp")
	assert.Nil(err, "error found")

	validateMyParam := pipelines["local.pipeline.validate_my_param"]
	if validateMyParam == nil {
		assert.Fail("validate_my_param pipeline not found")
		return
	}

	stringParam := map[string]string{
		"my_token": "abc",
	}

	res, errs := fpparse.CoerceParams(validateMyParam, stringParam, nil)
	if len(errs) > 0 {
		assert.Fail("error found")
		return
	}
	assert.NotNil(res)

	stringParamNotFound := map[string]string{
		"my_token_s": "abc",
	}
	_, errs = fpparse.CoerceParams(validateMyParam, stringParamNotFound, nil)
	assert.Equal(1, len(errs))
	assert.Equal("Bad Request: unknown parameter specified 'my_token_s'", errs[0].Error())

	stringParamNumberButValid := map[string]string{
		"my_token": "23",
	}
	res, errs = fpparse.CoerceParams(validateMyParam, stringParamNumberButValid, nil)
	if len(errs) > 0 {
		assert.Fail("error found")
		return
	}

	assert.NotNil(res)

	numParam := map[string]string{
		"my_number": "23",
	}
	res, errs = fpparse.CoerceParams(validateMyParam, numParam, nil)
	if len(errs) > 0 {
		assert.Fail("error found")
		return
	}

	assert.NotNil(res)
	assert.Equal(23, res["my_number"])

	numParamInvalid := map[string]string{
		"my_number": "foo",
	}
	_, errs = fpparse.CoerceParams(validateMyParam, numParamInvalid, nil)
	assert.Equal(1, len(errs))

	assert.Equal("Bad Request: unable to convert 'foo' to a number", errs[0].Error())

	boolParam := map[string]string{
		"my_bool": "true",
	}
	res, errs = fpparse.CoerceParams(validateMyParam, boolParam, nil)
	if len(errs) > 0 {
		assert.Fail("error found")
		return
	}

	assert.NotNil(res)
	assert.Equal(true, res["my_bool"])

	listSameTypes := map[string]string{
		"list_string":     `["foo", "bar", "3"]`,
		"list_number":     `[1, 2, 3]`,
		"list_number_two": `[1.1, 2.2, 3.4]`,
	}
	res, errs = fpparse.CoerceParams(validateMyParam, listSameTypes, nil)
	if len(errs) > 0 {
		assert.Fail("error found")
		return
	}

	assert.NotNil(res)
	assert.Equal(3, len(res["list_string"].([]string)))
	assert.Equal("foo", res["list_string"].([]string)[0])
	assert.Equal("bar", res["list_string"].([]string)[1])
	assert.Equal("3", res["list_string"].([]string)[2])

	assert.Equal(3, len(res["list_number"].([]float64)))
	assert.Equal(float64(1), res["list_number"].([]float64)[0])
	assert.Equal(float64(2), res["list_number"].([]float64)[1])
	assert.Equal(float64(3), res["list_number"].([]float64)[2])

	assert.Equal(3, len(res["list_number_two"].([]float64)))
	assert.Equal(1.1, res["list_number_two"].([]float64)[0])
	assert.Equal(2.2, res["list_number_two"].([]float64)[1])
	assert.Equal(3.4, res["list_number_two"].([]float64)[2])

	listStringInvalid := map[string]string{
		"list_string": `["foo", "bar", 3]`,
	}
	_, errs = fpparse.CoerceParams(validateMyParam, listStringInvalid, nil)
	assert.Equal(1, len(errs))
	assert.Equal("Bad Request: expected string type, but got number", errs[0].Error())

	moreInvalidList := map[string]string{
		"list_string": `["foo", "bar", 3]`,
		"list_number": `[1, "bar", 3]`,
	}
	_, errs = fpparse.CoerceParams(validateMyParam, moreInvalidList, nil)
	assert.Equal(2, len(errs))

	expectedErrors := []string{
		"Bad Request: expected string type, but got number",
		"Bad Request: expected number type, but got string",
	}

	actualErrors := []string{}
	for _, err := range errs {
		actualErrors = append(actualErrors, err.Error())
	}

	less := func(a, b string) bool { return a < b }
	equalIgnoreOrder := cmp.Equal(expectedErrors, actualErrors, cmpopts.SortSlices(less))
	assert.True(equalIgnoreOrder, "expected errors do not match")

	listAny := map[string]string{
		"list_any":       `["foo", 1, 1.23, true]`,
		"list_any_two":   `["foo", "bar", "baz"]`,
		"list_any_three": `[1, 2.3, 4]`,
	}
	res, errs = fpparse.CoerceParams(validateMyParam, listAny, nil)
	if len(errs) > 0 {
		assert.Fail("error found")
		return
	}
	assert.NotNil(res)

	assert.Equal(4, len(res["list_any"].([]interface{})))
	assert.Equal("foo", res["list_any"].([]interface{})[0])
	assert.Equal(1, res["list_any"].([]interface{})[1])
	assert.Equal(1.23, res["list_any"].([]interface{})[2])
	assert.Equal(true, res["list_any"].([]interface{})[3])

	assert.Equal(3, len(res["list_any_two"].([]interface{})))
	assert.Equal("foo", res["list_any_two"].([]interface{})[0])
	assert.Equal("bar", res["list_any_two"].([]interface{})[1])
	assert.Equal("baz", res["list_any_two"].([]interface{})[2])

	assert.Equal(3, len(res["list_any_three"].([]interface{})))
	assert.Equal(1, res["list_any_three"].([]interface{})[0])
	assert.Equal(2.3, res["list_any_three"].([]interface{})[1])
	assert.Equal(4, res["list_any_three"].([]interface{})[2])

	setSameTypes := map[string]string{
		"set_string": `["foo", "bar", "3"]`,
		"set_number": `[1, 2, 3]`,
	}
	res, errs = fpparse.CoerceParams(validateMyParam, setSameTypes, nil)
	if len(errs) > 0 {
		assert.Fail("error found")
		return
	}
	equalIgnoreOrder = cmp.Equal(expectedErrors, actualErrors, cmpopts.SortSlices(less))
	assert.True(equalIgnoreOrder, "expected errors do not match")

	assert.Equal(3, len(res["set_string"].([]string)))
	assert.Equal("foo", res["set_string"].([]string)[0])
	assert.Equal("bar", res["set_string"].([]string)[1])
	assert.Equal("3", res["set_string"].([]string)[2])

	assert.Equal(3, len(res["set_number"].([]float64)))
	assert.Equal(float64(1), res["set_number"].([]float64)[0])
	assert.Equal(float64(2), res["set_number"].([]float64)[1])
	assert.Equal(float64(3), res["set_number"].([]float64)[2])

	setFailures := map[string]string{
		"set_string": `["foo", "bar", "bar"]`,
	}
	_, errs = fpparse.CoerceParams(validateMyParam, setFailures, nil)
	expectedErrors = []string{
		"Bad Request: duplicate value found in set",
	}

	actualErrors = []string{}
	for _, err := range errs {
		actualErrors = append(actualErrors, err.Error())
	}

	equalIgnoreOrder = cmp.Equal(expectedErrors, actualErrors, cmpopts.SortSlices(less))
	assert.True(equalIgnoreOrder, "expected errors do not match")

	validMap := map[string]string{
		"map_of_string":     `{"foo": "bar", "baz": "qux"}`,
		"map_of_number":     `{"foo": 1.23, "baz": 4.56}`,
		"map_of_number_two": `{"foo": 1, "bar": 2}`,
		"map_of_any":        `{"foo": 1, "bar": 2.3, "baz": "qux", "bam": true}`,
		"map_of_bool":       `{"foo": true, "bar": false}`,
	}

	res, errs = fpparse.CoerceParams(validateMyParam, validMap, nil)
	if len(errs) > 0 {
		assert.Fail("error found")
		return
	}
	assert.NotNil(res)
	assert.Equal(2, len(res["map_of_string"].(map[string]string)))
	assert.Equal("bar", res["map_of_string"].(map[string]string)["foo"])
	assert.Equal("qux", res["map_of_string"].(map[string]string)["baz"])

	assert.Equal(2, len(res["map_of_number"].(map[string]float64)))
	assert.Equal(float64(1.23), res["map_of_number"].(map[string]float64)["foo"])
	assert.Equal(float64(4.56), res["map_of_number"].(map[string]float64)["baz"])

	assert.Equal(2, len(res["map_of_number_two"].(map[string]float64)))
	assert.Equal(float64(1), res["map_of_number_two"].(map[string]float64)["foo"])
	assert.Equal(float64(2), res["map_of_number_two"].(map[string]float64)["bar"])

	assert.Equal(4, len(res["map_of_any"].(map[string]interface{})))
	assert.Equal(1, res["map_of_any"].(map[string]interface{})["foo"])
	assert.Equal(2.3, res["map_of_any"].(map[string]interface{})["bar"])
	assert.Equal("qux", res["map_of_any"].(map[string]interface{})["baz"])
	assert.Equal(true, res["map_of_any"].(map[string]interface{})["bam"])

	assert.Equal(2, len(res["map_of_bool"].(map[string]bool)))
	assert.Equal(true, res["map_of_bool"].(map[string]bool)["foo"])
	assert.Equal(false, res["map_of_bool"].(map[string]bool)["bar"])

	invalidStringMap := map[string]string{
		"map_of_string": `{"foo": 1, "baz": "qux"}`,
	}

	_, errs = fpparse.CoerceParams(validateMyParam, invalidStringMap, nil)
	assert.Equal(1, len(errs))
	assert.Equal("Bad Request: expected string type, but got number", errs[0].Error())

	invalidNumberMap := map[string]string{
		"map_of_number": `{"foo": 1, "baz": "qux"}`,
	}
	_, errs = fpparse.CoerceParams(validateMyParam, invalidNumberMap, nil)
	assert.Equal(1, len(errs))
	assert.Equal("Bad Request: expected number type, but got string", errs[0].Error())
}

func TestParamCoerceWithAnyType(t *testing.T) {
	assert := assert.New(t)

	pipelines, _, err := fpparse.LoadPipelines(context.TODO(), "./pipelines/param_validation.fp")
	assert.Nil(err, "error found")

	validateMyParam := pipelines["local.pipeline.validate_my_param"]
	if validateMyParam == nil {
		assert.Fail("validate_my_param pipeline not found")
		return
	}

	stringParam := map[string]string{
		"param_any": "abc",
	}
	res, errs := fpparse.CoerceParams(validateMyParam, stringParam, nil)
	if len(errs) > 0 {
		assert.Fail("error found")
		return
	}

	assert.NotNil(res)
	assert.Equal("abc", res["param_any"])

	intParam := map[string]string{
		"param_any": "23",
	}

	res, errs = fpparse.CoerceParams(validateMyParam, intParam, nil)
	if len(errs) > 0 {
		assert.Fail("error found")
		return
	}

	assert.NotNil(res)
	assert.Equal(23, res["param_any"])

	complexParam := map[string]string{
		"param_any": `["foo", "bar", "baz"]`,
	}

	res, errs = fpparse.CoerceParams(validateMyParam, complexParam, nil)
	if len(errs) > 0 {
		assert.Fail("error found")
		return
	}

	assert.NotNil(res)
	assert.Equal(3, len(res["param_any"].([]interface{})))
	assert.Equal("foo", res["param_any"].([]interface{})[0])
	assert.Equal("bar", res["param_any"].([]interface{})[1])
	assert.Equal("baz", res["param_any"].([]interface{})[2])

	complexParam = map[string]string{
		"param_any": `["foo", 42, "baz"]`,
	}

	res, errs = fpparse.CoerceParams(validateMyParam, complexParam, nil)
	if len(errs) > 0 {
		assert.Fail("error found")
		return
	}

	assert.NotNil(res)
	assert.Equal(3, len(res["param_any"].([]interface{})))
	assert.Equal("foo", res["param_any"].([]interface{})[0])
	assert.Equal(42, res["param_any"].([]interface{})[1])
	assert.Equal("baz", res["param_any"].([]interface{})[2])

	complexParam = map[string]string{
		"param_any": `[{"foo": "bar"}, {"baz": "qux"}]`,
	}

	res, errs = fpparse.CoerceParams(validateMyParam, complexParam, nil)
	if len(errs) > 0 {
		assert.Fail("error found")
		return
	}

	assert.NotNil(res)
	assert.Equal(2, len(res["param_any"].([]interface{})))
	assert.Equal("bar", res["param_any"].([]interface{})[0].(map[string]interface{})["foo"])
	assert.Equal("qux", res["param_any"].([]interface{})[1].(map[string]interface{})["baz"])
}
