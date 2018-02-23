package main

import (
	"context"

	"github.com/akutz/gournal"
	"github.com/akutz/gournal/logrus"
)

func main() {
	ctx := context.Background()
	ctx = context.WithValue(ctx, gournal.LevelKey(), gournal.InfoLevel)
	ctx = context.WithValue(ctx, gournal.AppenderKey(), logrus.New())

	gournal.Info(ctx, "Hello %s", "Bob")

	gournal.WithFields(map[string]interface{}{
		"size":     1,
		"location": "Austin",
	}).Warn(ctx, "Hello %s", "Mary")
}
