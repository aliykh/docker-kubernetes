package runtime

import "log"

func Require(err error, desc string) {
	if err != nil {
		log.Fatalf("%s: %s", desc, err)
	}
}
