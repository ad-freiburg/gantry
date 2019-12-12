package gantry_test

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/ad-freiburg/gantry"
	"github.com/ad-freiburg/gantry/types"
)

// https://docs.docker.com/compose/wordpress/
const wordpressExample = `version: '3.3'

services:
   db:
     image: mysql:5.7
     volumes:
       - /tmp/data:/var/lib/mysql
     restart: always
     environment:
       MYSQL_ROOT_PASSWORD: somewordpress
       MYSQL_DATABASE: wordpress
       MYSQL_USER: wordpress
       MYSQL_PASSWORD: wordpress

   wordpress:
     depends_on:
       - db
     image: wordpress:latest
     ports:
       - "8000:80"
     restart: always
     environment:
       WORDPRESS_DB_HOST: db:3306
       WORDPRESS_DB_USER: wordpress
       WORDPRESS_DB_PASSWORD: wordpress
       WORDPRESS_DB_NAME: wordpress
`

func TestPipelineCompatibilityWordpress(t *testing.T) {
	tmpDef, err := ioutil.TempFile("", "def")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tmpDef.Name())
	err = ioutil.WriteFile(tmpDef.Name(), []byte(wordpressExample), 0644)
	if err != nil {
		log.Fatal(err)
	}

	cases := []struct {
		selected types.StringSet
		ignored  types.StringSet
	}{
		{types.StringSet{}, types.StringSet{"db": false, "wordpress": false}},
		{types.StringSet{"wordpress": true}, types.StringSet{"db": false, "wordpress": false}},
		{types.StringSet{"db": true}, types.StringSet{"db": false, "wordpress": true}},
	}

	// Perform parse and tests
	for i, c := range cases {
		p, err := gantry.NewPipeline(tmpDef.Name(), "", types.StringMap{}, types.StringSet{}, c.selected)
		if err != nil {
			t.Error(err)
		}
		if err := p.Check(); err != nil {
			t.Error(err)
		}
		for k, v := range c.ignored {
			if p.Definition.Steps[k].Meta.Ignore != v {
				t.Errorf("incorrect ignored state@%d for '%s', got: %t, wanted: %t", i, k, v, p.Definition.Steps[k].Meta.Ignore)
				t.Errorf("%#v", p.Definition.Steps)
			}
		}
	}
}

func TestPipelineCompatibilityUnknownService(t *testing.T) {
	tmpDef, err := ioutil.TempFile("", "def")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tmpDef.Name())
	err = ioutil.WriteFile(tmpDef.Name(), []byte(wordpressExample), 0644)
	if err != nil {
		log.Fatal(err)
	}

	cases := []struct {
		selected types.StringSet
		err      string
	}{
		{types.StringSet{}, ""},
		{types.StringSet{"foo": true}, "no such service or step: foo"},
	}

	// Perform parse and tests
	for i, c := range cases {
		_, err := gantry.NewPipeline(tmpDef.Name(), "", types.StringMap{}, types.StringSet{}, c.selected)
		if err != nil {
			if c.err == "" {
				t.Errorf("unexpected error @%d, got: %s, wanted: nil", i, err)
				continue
			}
			if err.Error() != c.err {
				t.Errorf("incorrect error @%d, got: %s, wanted: %s", i, err, c.err)
			}
		}
		if err == nil && c.err != "" {
			t.Errorf("failed to error @%d, got: nil, wanted: %s", i, c.err)
			continue
		}
	}
}
