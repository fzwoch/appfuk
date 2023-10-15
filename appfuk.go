//
// appfuk. Make macOS application bundles deployable.
// Copyright (C) 2023 Florian Zwoch <fzwoch@gmail.com>
//
// This file is part of appfuk.
//
// appfuk is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 2 of the License, or
// (at your option) any later version.
//
// appfuk is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with appfuk. If not, see <http://www.gnu.org/licenses/>.
//

package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type paths struct {
	searched string
	absolute string
}

var (
	otool             string
	install_name_tool string
	frameworks        string
)

func deps(exe string, indent string) {
	var (
		libs []paths
		next []paths
	)

	f := filepath.Base(exe)
	exe, err := filepath.EvalSymlinks(exe)
	if err != nil {
		panic(err)
	}
	dir := filepath.Dir(exe)
	exe = filepath.Join(dir, f)

	out, err := exec.Command(otool, "-L", exe).Output()
	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(bytes.NewReader(out))
	scanner.Scan()
	for scanner.Scan() {
		searched := strings.TrimSpace(scanner.Text())
		searched = strings.Fields(searched)[0]
		if searched == "" || strings.HasPrefix(searched, "/usr/lib/") || strings.HasPrefix(searched, "/System/Library/") {
			continue
		}
		if filepath.Base(exe) == filepath.Base(searched) {
			continue
		}

		absolute := searched
		if strings.HasPrefix(searched, "@") {
			absolute = strings.Replace(absolute, absolute[:strings.Index(absolute, "/")], dir, 1)
		}

		absolute, err = filepath.Abs(absolute)
		if err != nil {
			panic(err)
		}

		libs = append(libs, paths{searched: searched, absolute: absolute})
	}

	if len(libs) > 0 {
		fmt.Println(indent + exe + ":")
	}

	for _, paths := range libs {
		file := filepath.Base(paths.absolute)

		_, err = os.Stat(filepath.Join(frameworks, file))
		if err == nil {
			fmt.Println(indent + "  [skip] " + file)
			continue
		}

		fmt.Println(indent + "  [copy] " + file)

		i, err := os.Open(paths.absolute)
		if err != nil {
			panic(err)
		}
		o, err := os.Create(filepath.Join(frameworks, file))
		if err != nil {
			panic(err)
		}

		_, err = io.Copy(o, i)
		if err != nil {
			panic(err)
		}
		i.Close()
		o.Close()

		dst := filepath.Join(frameworks, f)
		if strings.Contains(exe, "/Contents/MacOS") {
			dst = exe
		}

		err = exec.Command(install_name_tool, "-change", paths.searched, "@executable_path/../Frameworks/"+file, dst).Run()
		if err != nil {
			panic(err)
		}

		next = append(next, paths)
	}

	for _, paths := range next {
		deps(paths.absolute, indent+"  ")
	}
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [options] <path/to/some.app/Contents/MacOS/exe>\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.StringVar(&otool, "otool", "otool", "otool executable")
	flag.StringVar(&install_name_tool, "install_name_tool", "install_name_tool", "install_name_tool executable")
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	exe, err := filepath.Abs(flag.Args()[0])
	if err != nil {
		panic(err)
	}

	info, err := os.Stat(exe)
	if err != nil {
		panic(err)
	}
	if info.IsDir() {
		panic(exe + " is a directory")
	}

	dir := filepath.Dir(exe)
	if !strings.HasSuffix(dir, "/Contents/MacOS") {
		panic("no bundle structure")
	}

	frameworks, err = filepath.Abs(filepath.Join(dir, "../Frameworks"))
	if err != nil {
		panic(nil)
	}
	_, err = os.Stat(frameworks)
	if err != nil {
		err = os.Mkdir(frameworks, 0755)
		if err != nil {
			panic(err)
		}
	}

	deps(exe, "")
}
