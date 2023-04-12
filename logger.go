package unbroken

import "go.uber.org/zap"

var log *zap.Logger

func init() {
	log = zap.Must(zap.NewDevelopment())
}
