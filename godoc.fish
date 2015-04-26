#!/usr/bin/fish

find -mindepth 2 -name README.md | xargs fgrep -L "[GoDoc]" | \
		while read file
					set pkg (echo $file | awk -F/ '{print $2;}')
					sed -i "3i [![GoDoc](https://godoc.org/github.com/icub3d/gop/$pkg?status.svg)](https://godoc.org/github.com/icub3d/gop/$pkg)\n" $file
		end
