package utils

import (
	p "path"
	"strings"
)

func MatchPath(path string, pattern string) bool {
	match, err := p.Match(pattern, path)
	if err != nil || match != true {
		return false
	}
	return true
}

func MatchPathPrefix(path string, pattern string) bool {
	folders := strings.Split(path, "/")
	patternFolders := strings.Split(pattern, "/")
	for d := range folders {
		if len(patternFolders) <= d {
			return true
		}
		match, err := p.Match(patternFolders[d], folders[d])
		if err != nil || match != true {
			return false
		}
	}
	return true
}

func PrefixForPattern(pattern string) string {
	minLength := len(pattern)
	l := strings.Index(pattern, "*")
	if l >= 0 && l < minLength {
		minLength = l
	}
	l = strings.Index(pattern, "?")
	if l >= 0 && l < minLength {
		minLength = l
	}
	if minLength == len(pattern) {
		return pattern
	}
	pathI := strings.LastIndex(pattern[:minLength], "/")
	return pattern[:pathI+1]
}

func CommonPrefixForPatterns(pattern1 string, pattern2 string) string {
	shortPattern := pattern1
	if len(pattern1) > len(pattern2) {
		shortPattern = pattern2
	}
	for c := range shortPattern {
		if pattern1[c] != pattern2[c] {
			return pattern1[:strings.LastIndex(pattern1[:c], "/")+1]
		}
		if pattern1[c] == '*' || pattern1[c] == '?' {
			return pattern1[:strings.LastIndex(pattern1[:c], "/")+1]
		}
		if pattern2[c] == '*' || pattern2[c] == '?' {
			return pattern1[:strings.LastIndex(pattern1[:c], "/")+1]
		}
	}
	return pattern1
}
