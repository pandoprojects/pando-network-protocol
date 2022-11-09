package main

import (
	"log"
	"strings"

	"github.com/spf13/cobra/doc"
	pando "github.com/pandoprojects/pando/cmd/pando/cmd"
	pandocli "github.com/pandoprojects/pando/cmd/pandocli/cmd"
)

func generatePandoCLIDoc(filePrepender, linkHandler func(string) string) {
	var all = pandocli.RootCmd
	err := doc.GenMarkdownTreeCustom(all, "./wallet/", filePrepender, linkHandler)
	if err != nil {
		log.Fatal(err)
	}
}

func generatePandoDoc(filePrepender, linkHandler func(string) string) {
	var all = pando.RootCmd
	err := doc.GenMarkdownTreeCustom(all, "./ledger/", filePrepender, linkHandler)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	filePrepender := func(filename string) string {
		return ""
	}

	linkHandler := func(name string) string {
		return strings.ToLower(name)
	}

	generatePandoCLIDoc(filePrepender, linkHandler)
	generatePandoDoc(filePrepender, linkHandler)
	Walk()
}
