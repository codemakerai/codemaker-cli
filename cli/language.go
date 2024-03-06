// Copyright 2023 CodeMaker AI Inc. All rights reserved.

package cli

import (
	"fmt"
	"github.com/codemakerai/codemaker-sdk-go/client"
)

var fileExtensions = map[string]string{
	".c":    client.LanguageC,
	".cpp":  client.LanguageCPP,
	".cxx":  client.LanguageCPP,
	".cc":   client.LanguageCPP,
	".js":   client.LanguageJavaScript,
	".jsx":  client.LanguageJavaScript,
	".java": client.LanguageJava,
	".php":  client.LanguagePHP,
	".cs":   client.LanguageCSharp,
	".go":   client.LanguageGo,
	".kt":   client.LanguageKotlin,
	".ts":   client.LanguageTypeScript,
	".tsx":  client.LanguageTypeScript,
	".rs":   client.LanguageRust,
}

var testFileSuffixes = map[string]string{
	client.LanguageJavaScript: "_test.js",
	client.LanguageJava:       "Test.java",
	client.LanguageGo:         "_test.go",
	client.LanguageKotlin:     "Test.kt",
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
