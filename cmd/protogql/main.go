package main

import (
	"flag"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/vektah/gqlparser/v2/formatter"

	"github.com/fizx/go-proto-gql/pkg/generator"
)

type arrayFlags []string

func (i *arrayFlags) String() string {
	return "str list"
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

var (
	importPath = arrayFlags{}
	fileNames  = arrayFlags{}
	svc        = flag.Bool("svc", false, "")
	merge      = flag.Bool("merge", false, "")
)

func main() {
	flag.Var(&importPath, "I", "path")
	flag.Var(&fileNames, "f", "path")
	flag.Parse()

	newFileNames, err := protoparse.ResolveFilenames(importPath, fileNames...)
	if err != nil {
		log.Fatal(err)
	}
	descs, err := protoparse.Parser{ImportPaths: importPath}.ParseFiles(newFileNames...)
	if err != nil {
		log.Fatal(err)
	}
	gqlDesc, err := generator.NewSchemas(descs, *merge, *svc)
	if err != nil {
		log.Fatal(err)
	}
	for _, schema := range gqlDesc {
		if len(schema.FileDescriptors) < 1 {
			log.Fatalf("unexpected number of proto descriptors: %d for gql schema", len(schema.FileDescriptors))
		}
		if len(schema.FileDescriptors) > 1 {
			if err := generateFile(schema, true); err != nil {
				log.Fatal(err)
			}
			break
		}
		if err := generateFile(schema, *merge); err != nil {
			log.Fatal(err)
		}
	}
}

func generateFile(schema *generator.SchemaDescriptor, merge bool) error {
	sc, err := os.Create(resolveGraphqlFilename(schema.FileDescriptors[0].GetName(), merge))
	if err != nil {
		return err
	}
	defer sc.Close()

	formatter.NewFormatter(sc).FormatSchema(schema.AsGraphql())
	return nil
}

func resolveGraphqlFilename(protoFileName string, merge bool) string {
	if merge {
		gqlFileName := "schema.graphqls"
		absProtoFileName, err := filepath.Abs(protoFileName)
		if err == nil {
			protoDirSlice := strings.Split(filepath.Dir(absProtoFileName), string(filepath.Separator))
			if len(protoDirSlice) > 0 {
				gqlFileName = protoDirSlice[len(protoDirSlice)-1] + ".graphqls"
			}
		}
		protoDir, _ := path.Split(protoFileName)
		return path.Join(protoDir, gqlFileName)
	}

	return strings.TrimSuffix(protoFileName, path.Ext(protoFileName)) + ".graphqls"
}
