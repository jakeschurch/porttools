#!/bin/bash
/home/jake/go/bin/./go-callvis -nostd -group type -focus porttools github.com/jakeschurch/porttools/example/ | dot -Tpng -o ../docs/porttoolsCallgraph.png

/home/jake/go/bin/./go-callvis -minlen 3 -nodesep 0.2 -nostd -focus porttools github.com/jakeschurch/porttools/example/ | dot -Tpng -o ../docs/porttoolsCallgraph1.png

/home/jake/go/bin/./go-callvis -group type -focus porttools github.com/jakeschurch/porttools/example/ | dot -Tpng -o ../docs/porttoolsErrgraph.png
