package utils

import "testing"

func TestCorrectPathMatching(t *testing.T) {
	assert := func(value bool, log string, args ...interface{}) {
		if value != true {
			t.Fatalf(log, args...)
		}
	}
	assert(MatchPath("/aaa/bbb/ccc", "/aaa/bbb/ccc"), "/aaa/bbb/ccc")
	assert(MatchPath("/aaa/bbb/ccc", "/aaa/bbb/*"), "/aaa/bbb/*")
	assert(MatchPath("/aaa/bbb/ccc", "/aaa/*/ccc"), "/aaa/*/ccc")
	assert(MatchPath("/aaa/bbb/ccc", "/aaa/*/*"), "/aaa/*/*")
	assert(MatchPath("/aaa/bbb/ccc", "/aaa/b*/ccc"), "/aaa/b*/ccc")
	assert(!MatchPath("/aaa/bbb/ccc", "/aaa/*"), "/aaa/*")
	assert(!MatchPath("/aaa/bbb/ccc", "/aaa/c*/ccc"), "/aaa/c*/ccc")
	assert(!MatchPath("/aaa/bbb/ccc", "/aaa/ccc/*"), "/aaa/ccc/ccc")
}

func TestCorrectPathPrefixMatching(t *testing.T) {
	assert := func(value bool, log string, args ...interface{}) {
		if value != true {
			t.Fatalf(log, args...)
		}
	}
	assert(MatchPathPrefix("/aaa/bbb/ccc/ddd/eee", "/aaa/bbb/ccc"), "/aaa/bbb/ccc")
	assert(MatchPathPrefix("/aaa/bbb/ccc/ddd/eee", "/aaa/bbb/*"), "/aaa/bbb/*")
	assert(MatchPathPrefix("/aaa/bbb/ccc/ddd/eee", "/aaa/*/ccc"), "/aaa/*/ccc")
	assert(MatchPathPrefix("/aaa/bbb/ccc/ddd/eee", "/aaa/*/*"), "/aaa/*/*")
	assert(MatchPathPrefix("/aaa/bbb/ccc/ddd/eee", "/aaa/b*/ccc"), "/aaa/b*/ccc")
	assert(MatchPathPrefix("/aaa/bbb/ccc/ddd/eee", "/aaa/*"), "/aaa/*")
	assert(!MatchPathPrefix("/aaa/bbb/ccc/ddd/eee", "/aaa/c*/ccc"), "/aaa/c*/ccc")
	assert(!MatchPathPrefix("/aaa/bbb/ccc/ddd/eee", "/aaa/ccc/*"), "/aaa/ccc/ccc")
}

func TestPrefixForPattern(t *testing.T) {
	assert := func(value string, required string) {
		if value != required {
			t.Fatalf("%s != %s", value, required)
		}
	}
	assert(PrefixForPattern("/aaa/bbb/ccc/*"), "/aaa/bbb/ccc/")
	assert(PrefixForPattern("/aaa/bbb/*"), "/aaa/bbb/")
	assert(PrefixForPattern("/aaa/bbb/c*"), "/aaa/bbb/")
	assert(PrefixForPattern("/aaa/bbb*"), "/aaa/")
	assert(PrefixForPattern("/aaa/*/ccc"), "/aaa/")
	assert(PrefixForPattern("/aaa/bbb/ccc"), "/aaa/bbb/ccc")
	assert(PrefixForPattern("/aaa"), "/aaa")
	assert(PrefixForPattern("/"), "/")
	assert(PrefixForPattern("/*"), "/")
}

func TestCommonPrefixForPatterns(t *testing.T) {
	assert := func(value string, required string) {
		if value != required {
			t.Fatalf("%s != %s", value, required)
		}
	}
	assert(CommonPrefixForPatterns("/aaa/bbb/*", "/aaa/bbb/ccc/*"), "/aaa/bbb/")
	assert(CommonPrefixForPatterns("/aaa/bbb/*/ccc", "/aaa/bbb/*"), "/aaa/bbb/")
	assert(CommonPrefixForPatterns("/aaa/bbb/*", "/ccc/ddd/*"), "/")
}
