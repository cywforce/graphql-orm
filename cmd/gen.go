package cmd

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"text/template"

	"github.com/inloop/goclitools"

	"github.com/novacloudcz/graphql-orm/model"
	"github.com/novacloudcz/graphql-orm/templates"
	"github.com/urfave/cli"
)

var genCmd = cli.Command{
	Name:  "generate",
	Usage: "generate contents",
	Action: func(ctx *cli.Context) error {
		if err := generate("model.graphql"); err != nil {
			return cli.NewExitError(err, 1)
		}
		return nil
	},
}

func generate(filename string) error {
	modelSource, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	m, err := model.Parse(string(modelSource))
	if err != nil {
		return err
	}
	// plainModel, err := model.Parse(string(modelSource))
	// if err != nil {
	// 	return err
	// }

	if _, err := os.Stat("./gen"); os.IsNotExist(err) {
		os.Mkdir("./gen", 0777)
	}

	err = generateFiles(m)
	if err != nil {
		return err
	}

	err = model.EnrichModel(&m)
	if err != nil {
		return err
	}

	schema, err := model.PrintSchema(m)
	if err != nil {
		return err
	}

	schema = "# This schema is generated, please don't update it manually\n\n" + schema

	if err := ioutil.WriteFile("gen/schema.graphql", []byte(schema), 0644); err != nil {
		return err
	}

	if err := goclitools.RunInteractiveInDir("go run github.com/99designs/gqlgen", "./gen"); err != nil {
		return err
	}

	// for _, obj := range plainModel.Objects() {
	// 	s1 := fmt.Sprintf("type %s struct {", obj.Name())
	// 	s2 := fmt.Sprintf("type %s struct {\n\t%sExtensions", obj.Name(), obj.Name())
	// 	if err := replaceStringInFile("gen/models_gen.go", s1, s2); err != nil {
	// 		return err
	// 	}
	// }

	return nil
}

func generateFiles(m model.Model) error {
	if err := writeTemplate(templates.Database, "gen/database.go", &m); err != nil {
		return err
	}
	if err := writeTemplate(templates.Resolver, "gen/resolver.go", &m); err != nil {
		return err
	}
	if err := writeTemplate(templates.GQLGen, "gen/gqlgen.yml", &m); err != nil {
		return err
	}
	if err := writeTemplate(templates.Model, "gen/models.go", &m); err != nil {
		return err
	}

	return nil
}

func writeTemplate(t, filename string, data interface{}) error {
	temp, err := template.New(filename).Parse(t)
	if err != nil {
		return err
	}
	var content bytes.Buffer
	writer := io.Writer(&content)

	err = temp.Execute(writer, data)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filename, content.Bytes(), 0777)
	if err != nil {
		return err
	}
	if path.Ext(filename) == ".go" {
		return goclitools.RunInteractive(fmt.Sprintf("gofmt -w %s", filename))
	}
	return nil
}

// func replaceStringInFile(filename, old, new string) error {
// 	content, err := ioutil.ReadFile(filename)
// 	if err != nil {
// 		return err
// 	}
// 	newContent := []byte(strings.ReplaceAll(string(content), old, new))

// 	return ioutil.WriteFile(filename, newContent, 0644)
// }