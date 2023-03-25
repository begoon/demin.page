package panic

import (
	"fmt"
	"path"
	"runtime"

	"github.com/rs/zerolog/log"
)

func caller() string {
	pc, file, line, ok := runtime.Caller(3)
	if ok {
		funcName := runtime.FuncForPC(pc).Name()
		fileName := path.Base(file)
		return fmt.Sprintf("%s: %s:%d", fileName, funcName, line)
	}
	return "?"
}

func Go(err error) {
	log.Panic().Err(err).Str("where", caller()).Send()
}
