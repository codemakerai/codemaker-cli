// Copyright 2023 CodeMaker AI Inc. All rights reserved.

package cli

import "fmt"

var fileExtensions = map[string]string{
	".java": "JAVA",
	".js":   "JAVASCRIPT",
}

var testFileSuffixes = map[string]string{
	"JAVA":       "Test.java",
	"JAVASCRIPT": "_test.js",
}

func languageFromExtension(extension string) (string, error) {
	if lang, ok := fileExtensions[extension]; ok {
		return lang, nil
	}
	return "", fmt.Errorf("the file extension %s is not supported", extension)
}

func testFileSuffix(language string) (string, error) {
	if lang, ok := testFileSuffixes[language]; ok {
		return lang, nil
	}
	return "", fmt.Errorf("the language %s is not supported", language)
}
