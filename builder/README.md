```shell

#!/usr/bin/env bash

ProgramName=test
ProgramVersion=1.0.0
ProgramBranch=`git rev-parse --abbrev-ref HEAD`
ProgramRevision=`git rev-parse HEAD`
CompilerVersion="`go version`"
BuildTime=`date -u '+%Y-%m-%d %H:%M:%S'`
Author=`whoami`@`hostname`

go build -ldflags "-X 'github.com/go-trellis/common/builder.ProgramName=$ProgramName' \
-X 'github.com/go-trellis/common/builder.ProgramVersion=$ProgramVersion' \
-X 'github.com/go-trellis/common/builder.ProgramBranch=$ProgramBranch' \
-X 'github.com/go-trellis/common/builder.ProgramRevision=$ProgramRevision' \
-X 'github.com/go-trellis/common/builder.CompilerVersion=${CompilerVersion}' \
-X 'github.com/go-trellis/common/builder.BuildTime=$BuildTime' \
-X 'github.com/go-trellis/common/builder.Author=$Author' \
" -o ${ProgramName} main.go

./${ProgramName}

rm ./${ProgramName}
```

```go
package main

import (
	"github.com/go-trellis/common/builder"
)

func main() {
	builder.Show()

	builder.Show(builder.OnShow(), builder.Color("{{ .AnsiColor.BrightRed }}"))
}
```