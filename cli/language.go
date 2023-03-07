// Copyright 2023 CodeMaker AI Inc. All rights reserved.

package cli

import "fmt"

var fileExtensions = map[string]string{
	".java": "JAVA",
}

func LanguageFromExtension(extension string) (string, error) {
	if lang, ok := fileExtensions[extension]; ok {
		return lang, nil
	}
	return "", fmt.Errorf("the file extension %s is not supported", extension)
}
