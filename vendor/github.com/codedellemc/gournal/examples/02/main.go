package main

import (
	"github.com/codedellemc/gournal"
	"github.com/codedellemc/gournal/logrus"
	"github.com/codedellemc/gournal/zap"
)

func main() {
	// Make a Zap-based Appender the default appender for when one is not
	// present in a Context, or when a nill Context is provided to a logging
	// function.
	gournal.DefaultAppender = zap.New()

	// The following call fails to provide a valid Context argument. In this
	// case the DefaultAppender is used.
	gournal.WithFields(map[string]interface{}{
		"size":     2,
		"location": "Boston",
	}).Error(nil, "Hello %s", "Bob")

	ctx := gournal.Background()
	ctx = gournal.WithValue(ctx, gournal.LevelKey(), gournal.InfoLevel)

	// Even though this next call provides a valid Context, there is no
	// Appender present in the Context so the DefaultAppender will be used.
	gournal.Info(ctx, "Hello %s", "Mary")

	ctx = gournal.WithValue(ctx, gournal.AppenderKey(), logrus.New())

	// This last log function uses a Context that has been created with a
	// Logrus Appender. Even though the DefaultAppender is assigned and is a
	// Zap-based logger, this call will utilize the Context Appender instance,
	// a Logrus Appender.
	gournal.WithFields(map[string]interface{}{
		"size":     1,
		"location": "Austin",
	}).Warn(ctx, "Hello %s", "Alice")
}
