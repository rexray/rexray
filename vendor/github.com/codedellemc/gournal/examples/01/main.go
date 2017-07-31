package main

import (
	"github.com/codedellemc/gournal"
	"github.com/codedellemc/gournal/logrus"
)

func main() {
	ctx := gournal.Background()
	ctx = gournal.WithValue(ctx, gournal.LevelKey(), gournal.InfoLevel)
	ctx = gournal.WithValue(ctx, gournal.AppenderKey(), logrus.New())

	gournal.Info(ctx, "Hello %s", "Bob")

	gournal.WithFields(map[string]interface{}{
		"size":     1,
		"location": "Austin",
	}).Warn(ctx, "Hello %s", "Mary")
}
