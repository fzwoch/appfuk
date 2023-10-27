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
	executable        string
	frameworks        string
)

func deps(exe string, indent string) {
	var (
		libs []paths
		next []paths
	)

	out, err := exec.Command(otool, "-L", exe).Output()
	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(bytes.NewReader(out))
	scanner.Scan()
	for scanner.Scan() {
		searched := strings.Fields(strings.TrimSpace(scanner.Text()))[0]
		if filepath.Base(exe) == filepath.Base(searched) || strings.HasPrefix(searched, "/usr/lib/") || strings.HasPrefix(searched, "/System/Library/") {
			continue
		}

		absolute := searched
		if strings.HasPrefix(searched, "@") {
			tmp, err := filepath.EvalSymlinks(exe)
			if err != nil {
				panic(err)
			}

			absolute = strings.Replace(absolute, absolute[:strings.Index(absolute, "/")], filepath.Dir(tmp), 1)
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
		} else {
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

			next = append(next, paths)
		}

		dst := filepath.Join(frameworks, filepath.Base(exe))
		if exe == executable {
			dst = exe
		}

		rel, err := filepath.Rel(filepath.Dir(executable), frameworks)
		if err != nil {
			panic(err)
		}

		err = exec.Command(install_name_tool, "-change", paths.searched, "@executable_path/"+filepath.Join(rel, file), dst).Run()
		if err != nil {
			panic(err)
		}
	}

	for _, paths := range next {
		deps(paths.absolute, indent+"  ")
	}
}

func main() {
	var err error

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

	executable, err = filepath.Abs(flag.Args()[0])
	if err != nil {
		panic(err)
	}

	info, err := os.Stat(executable)
	if err != nil {
		panic(err)
	}
	if info.IsDir() {
		panic(executable + " is a directory")
	}

	dir := strings.SplitAfter(filepath.Dir(executable), "/Contents/MacOS")[0]
	if !strings.HasSuffix(dir, "/Contents/MacOS") {
		panic("unexpected bundle structure")
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

	deps(executable, "")
}
