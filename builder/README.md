```shell

#!/usr/bin/env bash

ProgramName=test
ProgramVersion=1.0.0
ProgramBranch=`git rev-parse --abbrev-ref HEAD`
ProgramRevision=`git rev-parse HEAD`
CompilerVersion="`go version`"
BuildTime=`date -u '+%Y-%m-%d %H:%M:%S'`
Author=`whoami`@`hostname`

go build -ldflags "-X 'github.com/go-trellis/common.v2/builder.ProgramName=$ProgramName' \
-X 'github.com/go-trellis/common.v2/builder.ProgramVersion=$ProgramVersion' \
-X 'github.com/go-trellis/common.v2/builder.ProgramBranch=$ProgramBranch' \
-X 'github.com/go-trellis/common.v2/builder.ProgramRevision=$ProgramRevision' \
-X 'github.com/go-trellis/common.v2/builder.CompilerVersion=${CompilerVersion}' \
-X 'github.com/go-trellis/common.v2/builder.BuildTime=$BuildTime' \
-X 'github.com/go-trellis/common.v2/builder.Author=$Author' \
" -o ${ProgramName} main.go

./${ProgramName}

rm ./${ProgramName}
```

```go
package main

import (
	"github.com/go-trellis/common.v2/builder"
)

func main() {
	builder.Show()

	builder.Show(builder.OnShow(), builder.Color("{{ .AnsiColor.BrightRed }}"))
}
```