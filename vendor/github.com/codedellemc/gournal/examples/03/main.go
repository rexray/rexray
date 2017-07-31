package main

import (
	"github.com/codedellemc/gournal"
	"github.com/codedellemc/gournal/logrus"
)

func main() {
	ctx := gournal.Background()
	ctx = gournal.WithValue(ctx, gournal.LevelKey(), gournal.InfoLevel)
	ctx = gournal.WithValue(ctx, gournal.AppenderKey(), logrus.New())

	ctx = gournal.WithValue(
		ctx,
		gournal.FieldsKey(),
		map[string]interface{}{
			"name":  "Venus",
			"color": 0x00ff00,
		})

	// The following log entry will print the message and the name and color
	// of the planet.
	gournal.Info(ctx, "Discovered planet")

	ctx = gournal.WithValue(
		ctx,
		gournal.FieldsKey(),
		func() map[string]interface{} {
			return map[string]interface{}{
				"galaxy":   "Milky Way",
				"distance": 42,
			}
		})

	// The following log entry will print the message and the galactic location
	// and distance of the planet.
	gournal.Info(ctx, "Discovered planet")

	// Create a Context with the FieldsKey that points to a function which
	// returns a Context's derived fields based upon what data was provided
	// to a the log function.
	ctx = gournal.WithValue(
		ctx,
		gournal.FieldsKey(),
		func(ctx gournal.Context,
			lvl gournal.Level,
			fields map[string]interface{},
			args ...interface{}) map[string]interface{} {

			if v, ok := fields["z-value"].(int); ok {
				delete(fields, "z-value")
				return map[string]interface{}{
					"point": struct {
						x int
						y int
						z int
					}{1, -1, v},
				}
			}

			return map[string]interface{}{
				"point": struct {
					x int
					y int
				}{1, -1},
			}
		})

	// The following log entry will print the message and two-dimensional
	// location information about the planet.
	gournal.Info(ctx, "Discovered planet")

	// This log entry, however, will print the message and the same location
	// information, however, because the function used to derive the Context's
	// fields inspects the field's "z-value" key, it will add that data to the
	// location information, making it three-dimensional.
	gournal.WithField("z-value", 3).Info(ctx, "Discovered planet")
}
