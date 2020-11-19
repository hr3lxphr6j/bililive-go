package utils

import (
	"text/template"

	"github.com/Masterminds/sprig"
)

var functionList = map[string]interface{}{
	"decodeUnicode":      ParseUnicode,
	"replaceIllegalChar": ReplaceIllegalChar,
	"unescapeHTMLEntity": UnescapeHTMLEntity,
	"filenameFilter":     NewStringFilterChain(ReplaceIllegalChar, UnescapeHTMLEntity).Do,
}

func GetFuncMap() template.FuncMap {
	funcs := sprig.TxtFuncMap()
	for key, fn := range functionList {
		funcs[key] = fn
	}
	return funcs
}
