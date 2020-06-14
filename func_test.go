package xpath

import "testing"

type testQuery string

func (t testQuery) Select(_ iterator) NodeNavigator {
	panic("implement me")
}

func (t testQuery) Clone() query {
	return t
}

func (t testQuery) Evaluate(_ iterator) interface{} {
	return string(t)
}

const strForNormalization = "\t    \rloooooooonnnnnnngggggggg  \r \n tes  \u00a0 t strin \n\n \r g "
const expectedStrAfterNormalization = `loooooooonnnnnnngggggggg tes t strin g`

func Test_NormalizeSpaceFunc(t *testing.T) {
	result := normalizespaceFunc(testQuery(strForNormalization), nil).(string)
	if expectedStrAfterNormalization != result {
		t.Fatalf("unexpected result '%s'", result)
	}
}

func Test_ConcatFunc(t *testing.T) {
	result := concatFunc(testQuery("a"), testQuery("b"))(nil, nil).(string)
	if "ab" != result {
		t.Fatalf("unexpected result '%s'", result)
	}
}

func Benchmark_NormalizeSpaceFunc(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = normalizespaceFunc(testQuery(strForNormalization), nil)
	}
}

func Benchmark_ConcatFunc(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = concatFunc(testQuery("a"), testQuery("b"))(nil, nil)
	}
}
