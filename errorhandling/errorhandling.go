package errorhandling

import "log"

func ErrorResolver(errChan chan error, errLog *log.Logger, errLimit int) {
	lmt := 0
	for e := range errChan {
		lmt++
		if lmt >= errLimit {
			errLog.Println(e)
			log.Fatal("too many errors...check log:", e)
		}
		errLog.Println()
	}
}
