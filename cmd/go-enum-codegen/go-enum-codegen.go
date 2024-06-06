package main

import (
	"flag"
	"fmt"
	goenumcodegen "github.com/ejfrick/go-enum-codegen"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
)

var (
	flagTypeNames    string
	flagOutput       string
	flagErrOnUnk     bool
	flagBuildTags    string
	flagPrintUsage   bool
	flagPrintVersion bool
	flagJsonOnly     bool
	flagSQLOnly      bool
	flagUseStringer  bool
)

func errExitf(format string, args ...any) {
	log.Fatalf(format, args...)
}

var noVCSVersionOverride string

func main() {
	log.SetFlags(0)
	log.SetPrefix("go-enum-codegen: ")
	flag.StringVar(&flagTypeNames, "type", "", "comma-separated list of type names; must be set")
	flag.StringVar(&flagOutput, "output", "", "output file name; default srcdir/<type>.gen.go")
	flag.BoolVar(&flagErrOnUnk, "error-on-unknown", false, "whether to return an error if scanning or unmarshalling an unknown value; automatically set to true when iota is first set to \"_\" or there is no enum equal to the empty value of its underlying type; otherwise default is false and an unknown value will be assigned to the enum with the empty value of its underlying type")
	flag.BoolVar(&flagErrOnUnk, "e", false, "same as -error-on-unknown")
	flag.StringVar(&flagBuildTags, "tags", "", "comma-separated list of build tags to apply")
	flag.BoolVar(&flagPrintUsage, "help", false, "show this help and exit")
	flag.BoolVar(&flagPrintUsage, "h", false, "same as -help.")
	flag.BoolVar(&flagPrintVersion, "version", false, "show version and exit")
	flag.BoolVar(&flagJsonOnly, "json", false, "generate only json.Marshaler and json.Unmarshaler methods; default false")
	flag.BoolVar(&flagSQLOnly, "sql", false, "generate only sql.Scanner and driver.Value methods; default false")
	flag.BoolVar(&flagUseStringer, "stringer", false, "use the String() method of the enum instead of the underlying integer value; default false")

	flag.Parse()

	if flagPrintUsage {
		flag.Usage()
		return
	}

	if flagPrintVersion {
		buildInfo, ok := debug.ReadBuildInfo()
		if !ok {
			errExitf("error reading build info")
		}
		fmt.Println(buildInfo.Main.Path + "/cmd/go-enum-codegen")
		version := buildInfo.Main.Version
		if len(noVCSVersionOverride) > 0 {
			version = noVCSVersionOverride
		}
		fmt.Println(version)
		return
	}

	typeList := strings.Split(flagTypeNames, ",")
	if len(typeList) == 0 {
		flag.Usage()
		errExitf("no types specified")
	}

	tags := strings.Split(flagBuildTags, ",")

	args := flag.Args()
	if len(args) == 0 {
		args = []string{"."}
	}

	var dir string
	if len(args) == 1 && isDirectory(args[0]) {
		dir = args[0]
	} else {
		if len(tags) != 0 {
			errExitf("-tags option applies only to directories, not individual files")
		}
		dir = filepath.Dir(args[0])
	}

	if flagJsonOnly && flagSQLOnly {
		errExitf("`-only-json` and '-only-sql' are mutually exclusive")
	}

	var opts []goenumcodegen.Opt
	switch {
	case flagJsonOnly:
		opts = append(opts, goenumcodegen.WithOnlyJsonMethods())
	case flagSQLOnly:
		opts = append(opts, goenumcodegen.WithOnlySQLMethods())
	case flagErrOnUnk:
		opts = append(opts, goenumcodegen.WithErrorOnUnknown())
	case flagUseStringer:
		opts = append(opts, goenumcodegen.WithUseStringer())
	}

	g := goenumcodegen.NewGenerator(opts...)

	err := g.ParsePackage(args, tags)
	if err != nil {
		errExitf("error parsing package: %v", err)
	}

	g.WritePreamble(os.Args[1:])

	for _, typeName := range typeList {
		err := g.Generate(typeName)
		if err != nil {
			errExitf("error generating enum code for type %s: %v", typeName, err)
		}
	}

	src, err := g.Format()
	if err != nil {
		errExitf("error formatting code: %v", err)
	}
	outputName := flagOutput
	if outputName == "" {
		baseName := fmt.Sprintf("%s.gen.go", typeList[0])
		outputName = filepath.Join(dir, strings.ToLower(baseName))
	}

	err = os.WriteFile(outputName, src, 0644)
	if err != nil {
		errExitf("failed to write output file: %v", err)
	}
}

func isDirectory(name string) bool {
	info, err := os.Stat(name)
	if err != nil {
		log.Fatal(err)
	}
	return info.IsDir()
}
