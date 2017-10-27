package main

import (
	"context"
	"flag"
	"fmt"
	"html/template"
	"os"

	"github.com/thecodeteam/gocsi"
	"github.com/thecodeteam/gocsi/csi"
	"google.golang.org/grpc"
)

var identityCmds = []*cmd{
	&cmd{
		Name:    "getsupportedversions",
		Aliases: []string{"gets"},
		Action:  getSupportedVersions,
		Flags:   flagsGetSupportedVersions,
	},
	&cmd{
		Name:    "getplugininfo",
		Aliases: []string{"getp"},
		Action:  getPluginInfo,
		Flags:   flagsGetPluginInfo,
	},
}

///////////////////////////////////////////////////////////////////////////////
//                          GetSupportedVersions                             //
///////////////////////////////////////////////////////////////////////////////
func flagsGetSupportedVersions(
	ctx context.Context, rpc string) *flag.FlagSet {

	fs := flag.NewFlagSet(rpc, flag.ExitOnError)
	flagsGlobal(fs, versionFormat, "*csi.Version")

	fs.Usage = func() {
		fmt.Fprintf(
			os.Stderr,
			"usage: %s %s [ARGS...]\n",
			appName, rpc)
		fs.PrintDefaults()
	}
	return fs
}

func getSupportedVersions(
	ctx context.Context,
	fs *flag.FlagSet,
	cc *grpc.ClientConn) error {

	// initialize the csi client
	client := csi.NewIdentityClient(cc)

	// execute the rpc
	versions, err := gocsi.GetSupportedVersions(ctx, client)
	if err != nil {
		return err
	}

	// create a template for emitting the output
	tpl := template.New("template")
	if tpl, err = tpl.Parse(args.format); err != nil {
		return err
	}
	// emit the result
	for _, v := range versions {
		if err = tpl.Execute(
			os.Stdout, v); err != nil {
			return err
		}
	}

	return nil
}

///////////////////////////////////////////////////////////////////////////////
//                          GetPluginInfo                                    //
///////////////////////////////////////////////////////////////////////////////
func flagsGetPluginInfo(
	ctx context.Context, rpc string) *flag.FlagSet {

	fs := flag.NewFlagSet(rpc, flag.ExitOnError)
	flagsGlobal(fs, pluginInfoFormat, "*csi.GetPluginInfoResponse_Result")

	fs.Usage = func() {
		fmt.Fprintf(
			os.Stderr,
			"usage: %s %s [ARGS...]\n",
			appName, rpc)
		fs.PrintDefaults()
	}
	return fs
}

func getPluginInfo(
	ctx context.Context,
	fs *flag.FlagSet,
	cc *grpc.ClientConn) error {

	// initialize the csi client
	client := csi.NewIdentityClient(cc)

	// execute the rpc
	info, err := gocsi.GetPluginInfo(ctx, client, args.version)
	if err != nil {
		return err
	}

	// create a template for emitting the output
	tpl := template.New("template")
	if tpl, err = tpl.Parse(args.format); err != nil {
		return err
	}
	// emit the result
	if err = tpl.Execute(
		os.Stdout, info); err != nil {
		return err
	}

	return nil
}
