#!/bin/bash
/home/jake/code/go/workspace/bin/./go-callvis -nostd -group type -focus porttools github.com/jakeschurch/porttools/example/ | dot -Tpng -o ../docs/porttoolsCallgraph.png

/home/jake/code/go/workspace/bin/./go-callvis -group type -focus porttools github.com/jakeschurch/porttools/example/ | dot -Tpng -o ../docs/porttoolsErrgraph.png
