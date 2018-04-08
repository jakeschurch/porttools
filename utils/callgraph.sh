#!/bin/bash
/home/jake/go/bin/./go-callvis -minlen 3 -nodesep 0.2 -nostd -group type -ignore github.com/jakeschurch/porttools/utils -focus ""  github.com/jakeschurch/porttools/example/ | dot -Tpng -o ../docs/porttoolsErrgraph.png
