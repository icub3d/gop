#!/usr/bin/fish

find -name "*.go" | xargs fgrep -L "Copyright (c)" | \
		while read file
					cat .top.license $file > $file.new
					mv $file.new $file
		end
