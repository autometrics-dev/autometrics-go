package generate

import (
	"testing"

	"github.com/stretchr/testify/assert"

	internal "github.com/autometrics-dev/autometrics-go/internal/autometrics"
	"github.com/autometrics-dev/autometrics-go/pkg/autometrics"
)

// TestVanillaContext tests that autometrics correctly detects a context.Context in
// a function signature.
func TestVanillaContext(t *testing.T) {
	sourceCode := `// This is the package comment.
package main

import (
	"context"

	prom "github.com/autometrics-dev/autometrics-go/prometheus/autometrics"
)

// This comment is associated with the main function.
//
//autometrics:inst --no-doc --slo "Service Test" --success-target 99
func main(thisIsAContext context.Context) {
	fmt.Println(hello) // line comment 3
}
`

	want := "// This is the package comment.\n" +
		"package main\n" +
		"\n" +
		"import (\n" +
		"\t\"context\"\n" +
		"\n" +
		"\tprom \"github.com/autometrics-dev/autometrics-go/prometheus/autometrics\"\n" +
		")\n" +
		"\n" +
		"// This comment is associated with the main function.\n" +
		"//\n" +
		"//autometrics:inst --no-doc --slo \"Service Test\" --success-target 99\n" +
		"func main(thisIsAContext context.Context) {\n" +
		"\tthisIsAContext = prom.PreInstrument(prom.NewContext(\n" +
		"\t\tthisIsAContext,\n" +
		"\t\tprom.WithConcurrentCalls(true),\n" +
		"\t\tprom.WithCallerName(true),\n" +
		"\t\tprom.WithSloName(\"Service Test\"),\n" +
		"\t\tprom.WithAlertSuccess(99),\n" +
		"\t)) //autometrics:shadow-ctx\n" +
		"\tdefer prom.Instrument(thisIsAContext, nil) //autometrics:defer\n" +
		"\n" +
		"	fmt.Println(hello) // line comment 3\n" +
		"}\n"

	ctx, err := internal.NewGeneratorContext(autometrics.PROMETHEUS, defaultPrometheusInstanceUrl, false, true, false, false)
	if err != nil {
		t.Fatalf("error creating the generation context: %s", err)
	}

	actual, err := GenerateDocumentationAndInstrumentation(ctx, sourceCode, "main")
	if err != nil {
		t.Fatalf("error generating the documentation: %s", err)
	}

	assert.Equal(t, want, actual, "The generated source code is not as expected.")
}

// TestVanillaContextRenamed tests that autometrics correctly detects a context.Context in
// a function signature when the import is renamed.
func TestVanillaContextRenamed(t *testing.T) {
	sourceCode := `// This is the package comment.
package main

import (
	vanilla "context"

	prom "github.com/autometrics-dev/autometrics-go/prometheus/autometrics"
)

// This comment is associated with the main function.
//
//autometrics:inst --no-doc --slo "Service Test" --success-target 99
func main(thisIsAContext vanilla.Context) {
	fmt.Println(hello) // line comment 3
}
`

	want := "// This is the package comment.\n" +
		"package main\n" +
		"\n" +
		"import (\n" +
		"\tvanilla \"context\"\n" +
		"\n" +
		"\tprom \"github.com/autometrics-dev/autometrics-go/prometheus/autometrics\"\n" +
		")\n" +
		"\n" +
		"// This comment is associated with the main function.\n" +
		"//\n" +
		"//autometrics:inst --no-doc --slo \"Service Test\" --success-target 99\n" +
		"func main(thisIsAContext vanilla.Context) {\n" +
		"\tthisIsAContext = prom.PreInstrument(prom.NewContext(\n" +
		"\t\tthisIsAContext,\n" +
		"\t\tprom.WithConcurrentCalls(true),\n" +
		"\t\tprom.WithCallerName(true),\n" +
		"\t\tprom.WithSloName(\"Service Test\"),\n" +
		"\t\tprom.WithAlertSuccess(99),\n" +
		"\t)) //autometrics:shadow-ctx\n" +
		"\tdefer prom.Instrument(thisIsAContext, nil) //autometrics:defer\n" +
		"\n" +
		"	fmt.Println(hello) // line comment 3\n" +
		"}\n"

	ctx, err := internal.NewGeneratorContext(autometrics.PROMETHEUS, defaultPrometheusInstanceUrl, false, true, false, false)
	if err != nil {
		t.Fatalf("error creating the generation context: %s", err)
	}

	actual, err := GenerateDocumentationAndInstrumentation(ctx, sourceCode, "main")
	if err != nil {
		t.Fatalf("error generating the documentation: %s", err)
	}

	assert.Equal(t, want, actual, "The generated source code is not as expected.")
}

// TestVanillaContextAnonymous tests that autometrics correctly detects a context.Context in
// a function signature when the import is anon.
func TestVanillaContextAnonymous(t *testing.T) {
	sourceCode := `// This is the package comment.
package main

import (
	. "context"

	prom "github.com/autometrics-dev/autometrics-go/prometheus/autometrics"
)

// This comment is associated with the main function.
//
//autometrics:inst --no-doc --slo "Service Test" --success-target 99
func main(thisIsAContext Context) {
	fmt.Println(hello) // line comment 3
}
`

	want := "// This is the package comment.\n" +
		"package main\n" +
		"\n" +
		"import (\n" +
		"\t. \"context\"\n" +
		"\n" +
		"\tprom \"github.com/autometrics-dev/autometrics-go/prometheus/autometrics\"\n" +
		")\n" +
		"\n" +
		"// This comment is associated with the main function.\n" +
		"//\n" +
		"//autometrics:inst --no-doc --slo \"Service Test\" --success-target 99\n" +
		"func main(thisIsAContext Context) {\n" +
		"\tthisIsAContext = prom.PreInstrument(prom.NewContext(\n" +
		"\t\tthisIsAContext,\n" +
		"\t\tprom.WithConcurrentCalls(true),\n" +
		"\t\tprom.WithCallerName(true),\n" +
		"\t\tprom.WithSloName(\"Service Test\"),\n" +
		"\t\tprom.WithAlertSuccess(99),\n" +
		"\t)) //autometrics:shadow-ctx\n" +
		"\tdefer prom.Instrument(thisIsAContext, nil) //autometrics:defer\n" +
		"\n" +
		"	fmt.Println(hello) // line comment 3\n" +
		"}\n"

	ctx, err := internal.NewGeneratorContext(autometrics.PROMETHEUS, defaultPrometheusInstanceUrl, false, true, false, false)
	if err != nil {
		t.Fatalf("error creating the generation context: %s", err)
	}

	actual, err := GenerateDocumentationAndInstrumentation(ctx, sourceCode, "main")
	if err != nil {
		t.Fatalf("error generating the documentation: %s", err)
	}

	assert.Equal(t, want, actual, "The generated source code is not as expected.")
}

// TestHttpRequestContext tests that autometrics correctly detects a http.Request in
// a function signature.
func TestHttpRequestContext(t *testing.T) {
	sourceCode := `// This is the package comment.
package main

import (
	"net/http"

	prom "github.com/autometrics-dev/autometrics-go/prometheus/autometrics"
)

// This comment is associated with the main function.
//
//autometrics:inst --no-doc --slo "Service Test" --success-target 99
func main(w http.ResponseWriter, req *http.Request) {
	fmt.Println(hello) // line comment 3
}
`

	want := "// This is the package comment.\n" +
		"package main\n" +
		"\n" +
		"import (\n" +
		"\t\"net/http\"\n" +
		"\n" +
		"\tprom \"github.com/autometrics-dev/autometrics-go/prometheus/autometrics\"\n" +
		")\n" +
		"\n" +
		"// This comment is associated with the main function.\n" +
		"//\n" +
		"//autometrics:inst --no-doc --slo \"Service Test\" --success-target 99\n" +
		"func main(w http.ResponseWriter, req *http.Request) {\n" +
		"\tamCtx := prom.PreInstrument(prom.NewContext(\n" +
		"\t\treq.Context(),\n" +
		"\t\tprom.WithConcurrentCalls(true),\n" +
		"\t\tprom.WithCallerName(true),\n" +
		"\t\tprom.WithSloName(\"Service Test\"),\n" +
		"\t\tprom.WithAlertSuccess(99),\n" +
		"\t)) //autometrics:shadow-ctx\n" +
		"\tdefer prom.Instrument(amCtx, nil) //autometrics:defer\n" +
		"\n" +
		"	fmt.Println(hello) // line comment 3\n" +
		"}\n"

	ctx, err := internal.NewGeneratorContext(autometrics.PROMETHEUS, defaultPrometheusInstanceUrl, false, true, false, false)
	if err != nil {
		t.Fatalf("error creating the generation context: %s", err)
	}

	actual, err := GenerateDocumentationAndInstrumentation(ctx, sourceCode, "main")
	if err != nil {
		t.Fatalf("error generating the documentation: %s", err)
	}

	assert.Equal(t, want, actual, "The generated source code is not as expected.")
}

// TestHttpRequestContextRenamed tests that autometrics correctly detects a http.Request in
// a function signature when the import is renamed.
func TestHttpRequestContextRenamed(t *testing.T) {
	sourceCode := `// This is the package comment.
package main

import (
	vanilla "net/http"

	prom "github.com/autometrics-dev/autometrics-go/prometheus/autometrics"
)

// This comment is associated with the main function.
//
//autometrics:inst --no-doc --slo "Service Test" --success-target 99
func main(w vanilla.ResponseWriter, req *vanilla.Request) {
	fmt.Println(hello) // line comment 3
}
`

	want := "// This is the package comment.\n" +
		"package main\n" +
		"\n" +
		"import (\n" +
		"\tvanilla \"net/http\"\n" +
		"\n" +
		"\tprom \"github.com/autometrics-dev/autometrics-go/prometheus/autometrics\"\n" +
		")\n" +
		"\n" +
		"// This comment is associated with the main function.\n" +
		"//\n" +
		"//autometrics:inst --no-doc --slo \"Service Test\" --success-target 99\n" +
		"func main(w vanilla.ResponseWriter, req *vanilla.Request) {\n" +
		"\tamCtx := prom.PreInstrument(prom.NewContext(\n" +
		"\t\treq.Context(),\n" +
		"\t\tprom.WithConcurrentCalls(true),\n" +
		"\t\tprom.WithCallerName(true),\n" +
		"\t\tprom.WithSloName(\"Service Test\"),\n" +
		"\t\tprom.WithAlertSuccess(99),\n" +
		"\t)) //autometrics:shadow-ctx\n" +
		"\tdefer prom.Instrument(amCtx, nil) //autometrics:defer\n" +
		"\n" +
		"	fmt.Println(hello) // line comment 3\n" +
		"}\n"

	ctx, err := internal.NewGeneratorContext(autometrics.PROMETHEUS, defaultPrometheusInstanceUrl, false, true, false, false)
	if err != nil {
		t.Fatalf("error creating the generation context: %s", err)
	}

	actual, err := GenerateDocumentationAndInstrumentation(ctx, sourceCode, "main")
	if err != nil {
		t.Fatalf("error generating the documentation: %s", err)
	}

	assert.Equal(t, want, actual, "The generated source code is not as expected.")
}

// TestHttpRequestContextAnonymous tests that autometrics correctly detects a http.Request in
// a function signature when the import is anon.
func TestHttpRequestContextAnonymous(t *testing.T) {
	sourceCode := `// This is the package comment.
package main

import (
	. "net/http"

	prom "github.com/autometrics-dev/autometrics-go/prometheus/autometrics"
)

// This comment is associated with the main function.
//
//autometrics:inst --no-doc --slo "Service Test" --success-target 99
func main(w ResponseWriter, req *Request) {
	fmt.Println(hello) // line comment 3
}
`

	want := "// This is the package comment.\n" +
		"package main\n" +
		"\n" +
		"import (\n" +
		"\t. \"net/http\"\n" +
		"\n" +
		"\tprom \"github.com/autometrics-dev/autometrics-go/prometheus/autometrics\"\n" +
		")\n" +
		"\n" +
		"// This comment is associated with the main function.\n" +
		"//\n" +
		"//autometrics:inst --no-doc --slo \"Service Test\" --success-target 99\n" +
		"func main(w ResponseWriter, req *Request) {\n" +
		"\tamCtx := prom.PreInstrument(prom.NewContext(\n" +
		"\t\treq.Context(),\n" +
		"\t\tprom.WithConcurrentCalls(true),\n" +
		"\t\tprom.WithCallerName(true),\n" +
		"\t\tprom.WithSloName(\"Service Test\"),\n" +
		"\t\tprom.WithAlertSuccess(99),\n" +
		"\t)) //autometrics:shadow-ctx\n" +
		"\tdefer prom.Instrument(amCtx, nil) //autometrics:defer\n" +
		"\n" +
		"	fmt.Println(hello) // line comment 3\n" +
		"}\n"

	ctx, err := internal.NewGeneratorContext(autometrics.PROMETHEUS, defaultPrometheusInstanceUrl, false, true, false, false)
	if err != nil {
		t.Fatalf("error creating the generation context: %s", err)
	}

	actual, err := GenerateDocumentationAndInstrumentation(ctx, sourceCode, "main")
	if err != nil {
		t.Fatalf("error generating the documentation: %s", err)
	}

	assert.Equal(t, want, actual, "The generated source code is not as expected.")
}

// TestBuffaloContext tests that autometrics correctly detects a buffalo.Context in
// a function signature.
func TestBuffaloContext(t *testing.T) {
	sourceCode := `// This is the package comment.
package main

import (
	"github.com/gobuffalo/buffalo"

	prom "github.com/autometrics-dev/autometrics-go/prometheus/autometrics"
)

// This comment is associated with the main function.
//
//autometrics:inst --no-doc --slo "Service Test" --success-target 99
func main(thisIsAContext buffalo.Context) {
	fmt.Println(hello) // line comment 3
}
`

	want := "// This is the package comment.\n" +
		"package main\n" +
		"\n" +
		"import (\n" +
		"\t\"github.com/gobuffalo/buffalo\"\n" +
		"\n" +
		"\tprom \"github.com/autometrics-dev/autometrics-go/prometheus/autometrics\"\n" +
		")\n" +
		"\n" +
		"// This comment is associated with the main function.\n" +
		"//\n" +
		"//autometrics:inst --no-doc --slo \"Service Test\" --success-target 99\n" +
		"func main(thisIsAContext buffalo.Context) {\n" +
		"\tthisIsAContext = prom.PreInstrument(prom.NewContext(\n" +
		"\t\tthisIsAContext,\n" +
		"\t\tprom.WithConcurrentCalls(true),\n" +
		"\t\tprom.WithCallerName(true),\n" +
		"\t\tprom.WithSloName(\"Service Test\"),\n" +
		"\t\tprom.WithAlertSuccess(99),\n" +
		"\t)) //autometrics:shadow-ctx\n" +
		"\tdefer prom.Instrument(thisIsAContext, nil) //autometrics:defer\n" +
		"\n" +
		"	fmt.Println(hello) // line comment 3\n" +
		"}\n"

	ctx, err := internal.NewGeneratorContext(autometrics.PROMETHEUS, defaultPrometheusInstanceUrl, false, true, false, false)
	if err != nil {
		t.Fatalf("error creating the generation context: %s", err)
	}

	actual, err := GenerateDocumentationAndInstrumentation(ctx, sourceCode, "main")
	if err != nil {
		t.Fatalf("error generating the documentation: %s", err)
	}

	assert.Equal(t, want, actual, "The generated source code is not as expected.")
}

// TestBuffaloContextRenamed tests that autometrics correctly detects a buffalo.Context in
// a function signature when the import is renamed.
func TestBuffaloContextRenamed(t *testing.T) {
	sourceCode := `// This is the package comment.
package main

import (
	vanilla "github.com/gobuffalo/buffalo"

	prom "github.com/autometrics-dev/autometrics-go/prometheus/autometrics"
)

// This comment is associated with the main function.
//
//autometrics:inst --no-doc --slo "Service Test" --success-target 99
func main(thisIsAContext vanilla.Context) {
	fmt.Println(hello) // line comment 3
}
`

	want := "// This is the package comment.\n" +
		"package main\n" +
		"\n" +
		"import (\n" +
		"\tvanilla \"github.com/gobuffalo/buffalo\"\n" +
		"\n" +
		"\tprom \"github.com/autometrics-dev/autometrics-go/prometheus/autometrics\"\n" +
		")\n" +
		"\n" +
		"// This comment is associated with the main function.\n" +
		"//\n" +
		"//autometrics:inst --no-doc --slo \"Service Test\" --success-target 99\n" +
		"func main(thisIsAContext vanilla.Context) {\n" +
		"\tthisIsAContext = prom.PreInstrument(prom.NewContext(\n" +
		"\t\tthisIsAContext,\n" +
		"\t\tprom.WithConcurrentCalls(true),\n" +
		"\t\tprom.WithCallerName(true),\n" +
		"\t\tprom.WithSloName(\"Service Test\"),\n" +
		"\t\tprom.WithAlertSuccess(99),\n" +
		"\t)) //autometrics:shadow-ctx\n" +
		"\tdefer prom.Instrument(thisIsAContext, nil) //autometrics:defer\n" +
		"\n" +
		"	fmt.Println(hello) // line comment 3\n" +
		"}\n"

	ctx, err := internal.NewGeneratorContext(autometrics.PROMETHEUS, defaultPrometheusInstanceUrl, false, true, false, false)
	if err != nil {
		t.Fatalf("error creating the generation context: %s", err)
	}

	actual, err := GenerateDocumentationAndInstrumentation(ctx, sourceCode, "main")
	if err != nil {
		t.Fatalf("error generating the documentation: %s", err)
	}

	assert.Equal(t, want, actual, "The generated source code is not as expected.")
}

// TestBuffaloContextAnonymous tests that autometrics correctly detects a buffalo.Context in
// a function signature when the import is anon.
func TestBuffaloContextAnonymous(t *testing.T) {
	sourceCode := `// This is the package comment.
package main

import (
	. "github.com/gobuffalo/buffalo"

	prom "github.com/autometrics-dev/autometrics-go/prometheus/autometrics"
)

// This comment is associated with the main function.
//
//autometrics:inst --no-doc --slo "Service Test" --success-target 99
func main(thisIsAContext Context) {
	fmt.Println(hello) // line comment 3
}
`

	want := "// This is the package comment.\n" +
		"package main\n" +
		"\n" +
		"import (\n" +
		"\t. \"github.com/gobuffalo/buffalo\"\n" +
		"\n" +
		"\tprom \"github.com/autometrics-dev/autometrics-go/prometheus/autometrics\"\n" +
		")\n" +
		"\n" +
		"// This comment is associated with the main function.\n" +
		"//\n" +
		"//autometrics:inst --no-doc --slo \"Service Test\" --success-target 99\n" +
		"func main(thisIsAContext Context) {\n" +
		"\tthisIsAContext = prom.PreInstrument(prom.NewContext(\n" +
		"\t\tthisIsAContext,\n" +
		"\t\tprom.WithConcurrentCalls(true),\n" +
		"\t\tprom.WithCallerName(true),\n" +
		"\t\tprom.WithSloName(\"Service Test\"),\n" +
		"\t\tprom.WithAlertSuccess(99),\n" +
		"\t)) //autometrics:shadow-ctx\n" +
		"\tdefer prom.Instrument(thisIsAContext, nil) //autometrics:defer\n" +
		"\n" +
		"	fmt.Println(hello) // line comment 3\n" +
		"}\n"

	ctx, err := internal.NewGeneratorContext(autometrics.PROMETHEUS, defaultPrometheusInstanceUrl, false, true, false, false)
	if err != nil {
		t.Fatalf("error creating the generation context: %s", err)
	}

	actual, err := GenerateDocumentationAndInstrumentation(ctx, sourceCode, "main")
	if err != nil {
		t.Fatalf("error generating the documentation: %s", err)
	}

	assert.Equal(t, want, actual, "The generated source code is not as expected.")
}

// TestEchoContext tests that autometrics correctly detects an echo.Context in
// a function signature.
func TestEchoContext(t *testing.T) {
	sourceCode := `// This is the package comment.
package main

import (
	"github.com/labstack/echo/v4"

	prom "github.com/autometrics-dev/autometrics-go/prometheus/autometrics"
)

// This comment is associated with the main function.
//
//autometrics:inst --no-doc --slo "Service Test" --success-target 99
func main(thisIsAContext echo.Context) {
	fmt.Println(hello) // line comment 3
}
`

	want := "// This is the package comment.\n" +
		"package main\n" +
		"\n" +
		"import (\n" +
		"\t\"github.com/labstack/echo/v4\"\n" +
		"\n" +
		"\tprom \"github.com/autometrics-dev/autometrics-go/prometheus/autometrics\"\n" +
		")\n" +
		"\n" +
		"// This comment is associated with the main function.\n" +
		"//\n" +
		"//autometrics:inst --no-doc --slo \"Service Test\" --success-target 99\n" +
		"func main(thisIsAContext echo.Context) {\n" +
		"\tamCtx := prom.PreInstrument(prom.NewContext(\n" +
		"\t\tnil,\n" +
		"\t\tprom.WithTraceID(prom.DecodeString(thisIsAContext.Get(\"autometricsTraceID\"))),\n" +
		"\t\tprom.WithSpanID(prom.DecodeString(thisIsAContext.Get(\"autometricsSpanID\"))),\n" +
		"\t\tprom.WithConcurrentCalls(true),\n" +
		"\t\tprom.WithCallerName(true),\n" +
		"\t\tprom.WithSloName(\"Service Test\"),\n" +
		"\t\tprom.WithAlertSuccess(99),\n" +
		"\t)) //autometrics:shadow-ctx\n" +
		"\tdefer prom.Instrument(amCtx, nil) //autometrics:defer\n" +
		"\n" +
		"	fmt.Println(hello) // line comment 3\n" +
		"}\n"

	ctx, err := internal.NewGeneratorContext(autometrics.PROMETHEUS, defaultPrometheusInstanceUrl, false, true, false, false)
	if err != nil {
		t.Fatalf("error creating the generation context: %s", err)
	}

	actual, err := GenerateDocumentationAndInstrumentation(ctx, sourceCode, "main")
	if err != nil {
		t.Fatalf("error generating the documentation: %s", err)
	}

	assert.Equal(t, want, actual, "The generated source code is not as expected.")
}

// TestEchoContextRenamed tests that autometrics correctly detects an echo.Context in
// a function signature when the import is renamed.
func TestEchoContextRenamed(t *testing.T) {
	sourceCode := `// This is the package comment.
package main

import (
	vanilla "github.com/labstack/echo/v4"

	prom "github.com/autometrics-dev/autometrics-go/prometheus/autometrics"
)

// This comment is associated with the main function.
//
//autometrics:inst --no-doc --slo "Service Test" --success-target 99
func main(thisIsAContext vanilla.Context) {
	fmt.Println(hello) // line comment 3
}
`

	want := "// This is the package comment.\n" +
		"package main\n" +
		"\n" +
		"import (\n" +
		"\tvanilla \"github.com/labstack/echo/v4\"\n" +
		"\n" +
		"\tprom \"github.com/autometrics-dev/autometrics-go/prometheus/autometrics\"\n" +
		")\n" +
		"\n" +
		"// This comment is associated with the main function.\n" +
		"//\n" +
		"//autometrics:inst --no-doc --slo \"Service Test\" --success-target 99\n" +
		"func main(thisIsAContext vanilla.Context) {\n" +
		"\tamCtx := prom.PreInstrument(prom.NewContext(\n" +
		"\t\tnil,\n" +
		"\t\tprom.WithTraceID(prom.DecodeString(thisIsAContext.Get(\"autometricsTraceID\"))),\n" +
		"\t\tprom.WithSpanID(prom.DecodeString(thisIsAContext.Get(\"autometricsSpanID\"))),\n" +
		"\t\tprom.WithConcurrentCalls(true),\n" +
		"\t\tprom.WithCallerName(true),\n" +
		"\t\tprom.WithSloName(\"Service Test\"),\n" +
		"\t\tprom.WithAlertSuccess(99),\n" +
		"\t)) //autometrics:shadow-ctx\n" +
		"\tdefer prom.Instrument(amCtx, nil) //autometrics:defer\n" +
		"\n" +
		"	fmt.Println(hello) // line comment 3\n" +
		"}\n"

	ctx, err := internal.NewGeneratorContext(autometrics.PROMETHEUS, defaultPrometheusInstanceUrl, false, true, false, false)
	if err != nil {
		t.Fatalf("error creating the generation context: %s", err)
	}

	actual, err := GenerateDocumentationAndInstrumentation(ctx, sourceCode, "main")
	if err != nil {
		t.Fatalf("error generating the documentation: %s", err)
	}

	assert.Equal(t, want, actual, "The generated source code is not as expected.")
}

// TestEchoContextAnonymous tests that autometrics correctly detects an echo.Context in
// a function signature when the import is anon.
func TestEchoContextAnonymous(t *testing.T) {
	sourceCode := `// This is the package comment.
package main

import (
	. "github.com/labstack/echo/v4"

	prom "github.com/autometrics-dev/autometrics-go/prometheus/autometrics"
)

// This comment is associated with the main function.
//
//autometrics:inst --no-doc --slo "Service Test" --success-target 99
func main(thisIsAContext Context) {
	fmt.Println(hello) // line comment 3
}
`

	want := "// This is the package comment.\n" +
		"package main\n" +
		"\n" +
		"import (\n" +
		"\t. \"github.com/labstack/echo/v4\"\n" +
		"\n" +
		"\tprom \"github.com/autometrics-dev/autometrics-go/prometheus/autometrics\"\n" +
		")\n" +
		"\n" +
		"// This comment is associated with the main function.\n" +
		"//\n" +
		"//autometrics:inst --no-doc --slo \"Service Test\" --success-target 99\n" +
		"func main(thisIsAContext Context) {\n" +
		"\tamCtx := prom.PreInstrument(prom.NewContext(\n" +
		"\t\tnil,\n" +
		"\t\tprom.WithTraceID(prom.DecodeString(thisIsAContext.Get(\"autometricsTraceID\"))),\n" +
		"\t\tprom.WithSpanID(prom.DecodeString(thisIsAContext.Get(\"autometricsSpanID\"))),\n" +
		"\t\tprom.WithConcurrentCalls(true),\n" +
		"\t\tprom.WithCallerName(true),\n" +
		"\t\tprom.WithSloName(\"Service Test\"),\n" +
		"\t\tprom.WithAlertSuccess(99),\n" +
		"\t)) //autometrics:shadow-ctx\n" +
		"\tdefer prom.Instrument(amCtx, nil) //autometrics:defer\n" +
		"\n" +
		"	fmt.Println(hello) // line comment 3\n" +
		"}\n"

	ctx, err := internal.NewGeneratorContext(autometrics.PROMETHEUS, defaultPrometheusInstanceUrl, false, true, false, false)
	if err != nil {
		t.Fatalf("error creating the generation context: %s", err)
	}

	actual, err := GenerateDocumentationAndInstrumentation(ctx, sourceCode, "main")
	if err != nil {
		t.Fatalf("error generating the documentation: %s", err)
	}

	assert.Equal(t, want, actual, "The generated source code is not as expected.")
}

// TestGinContext tests that autometrics correctly detects a gin.Context in
// a function signature.
func TestGinContext(t *testing.T) {
	sourceCode := `// This is the package comment.
package main

import (
	"github.com/gin-gonic/gin"

	prom "github.com/autometrics-dev/autometrics-go/prometheus/autometrics"
)

// This comment is associated with the main function.
//
//autometrics:inst --no-doc --slo "Service Test" --success-target 99
func main(thisIsAContext *gin.Context) {
	fmt.Println(hello) // line comment 3
}
`

	want := "// This is the package comment.\n" +
		"package main\n" +
		"\n" +
		"import (\n" +
		"\t\"github.com/gin-gonic/gin\"\n" +
		"\n" +
		"\tprom \"github.com/autometrics-dev/autometrics-go/prometheus/autometrics\"\n" +
		")\n" +
		"\n" +
		"// This comment is associated with the main function.\n" +
		"//\n" +
		"//autometrics:inst --no-doc --slo \"Service Test\" --success-target 99\n" +
		"func main(thisIsAContext *gin.Context) {\n" +
		"\tamCtx := prom.PreInstrument(prom.NewContext(\n" +
		"\t\tnil,\n" +
		"\t\tprom.WithTraceID(prom.DecodeString(thisIsAContext.GetString(\"autometricsTraceID\"))),\n" +
		"\t\tprom.WithSpanID(prom.DecodeString(thisIsAContext.GetString(\"autometricsSpanID\"))),\n" +
		"\t\tprom.WithConcurrentCalls(true),\n" +
		"\t\tprom.WithCallerName(true),\n" +
		"\t\tprom.WithSloName(\"Service Test\"),\n" +
		"\t\tprom.WithAlertSuccess(99),\n" +
		"\t)) //autometrics:shadow-ctx\n" +
		"\tdefer prom.Instrument(amCtx, nil) //autometrics:defer\n" +
		"\n" +
		"	fmt.Println(hello) // line comment 3\n" +
		"}\n"

	ctx, err := internal.NewGeneratorContext(autometrics.PROMETHEUS, defaultPrometheusInstanceUrl, false, true, false, false)
	if err != nil {
		t.Fatalf("error creating the generation context: %s", err)
	}

	actual, err := GenerateDocumentationAndInstrumentation(ctx, sourceCode, "main")
	if err != nil {
		t.Fatalf("error generating the documentation: %s", err)
	}

	assert.Equal(t, want, actual, "The generated source code is not as expected.")
}

// TestGinContextRenamed tests that autometrics correctly detects a gin.Context in
// a function signature when the import is renamed.
func TestGinContextRenamed(t *testing.T) {
	sourceCode := `// This is the package comment.
package main

import (
	vanilla "github.com/gin-gonic/gin"

	prom "github.com/autometrics-dev/autometrics-go/prometheus/autometrics"
)

// This comment is associated with the main function.
//
//autometrics:inst --no-doc --slo "Service Test" --success-target 99
func main(thisIsAContext *vanilla.Context) {
	fmt.Println(hello) // line comment 3
}
`

	want := "// This is the package comment.\n" +
		"package main\n" +
		"\n" +
		"import (\n" +
		"\tvanilla \"github.com/gin-gonic/gin\"\n" +
		"\n" +
		"\tprom \"github.com/autometrics-dev/autometrics-go/prometheus/autometrics\"\n" +
		")\n" +
		"\n" +
		"// This comment is associated with the main function.\n" +
		"//\n" +
		"//autometrics:inst --no-doc --slo \"Service Test\" --success-target 99\n" +
		"func main(thisIsAContext *vanilla.Context) {\n" +
		"\tamCtx := prom.PreInstrument(prom.NewContext(\n" +
		"\t\tnil,\n" +
		"\t\tprom.WithTraceID(prom.DecodeString(thisIsAContext.GetString(\"autometricsTraceID\"))),\n" +
		"\t\tprom.WithSpanID(prom.DecodeString(thisIsAContext.GetString(\"autometricsSpanID\"))),\n" +
		"\t\tprom.WithConcurrentCalls(true),\n" +
		"\t\tprom.WithCallerName(true),\n" +
		"\t\tprom.WithSloName(\"Service Test\"),\n" +
		"\t\tprom.WithAlertSuccess(99),\n" +
		"\t)) //autometrics:shadow-ctx\n" +
		"\tdefer prom.Instrument(amCtx, nil) //autometrics:defer\n" +
		"\n" +
		"	fmt.Println(hello) // line comment 3\n" +
		"}\n"

	ctx, err := internal.NewGeneratorContext(autometrics.PROMETHEUS, defaultPrometheusInstanceUrl, false, true, false, false)
	if err != nil {
		t.Fatalf("error creating the generation context: %s", err)
	}

	actual, err := GenerateDocumentationAndInstrumentation(ctx, sourceCode, "main")
	if err != nil {
		t.Fatalf("error generating the documentation: %s", err)
	}

	assert.Equal(t, want, actual, "The generated source code is not as expected.")
}

// TestGinContextAnonymous tests that autometrics correctly detects a gin.Context in
// a function signature when the import is anon.
func TestGinContextAnonymous(t *testing.T) {
	sourceCode := `// This is the package comment.
package main

import (
	. "github.com/gin-gonic/gin"

	prom "github.com/autometrics-dev/autometrics-go/prometheus/autometrics"
)

// This comment is associated with the main function.
//
//autometrics:inst --no-doc --slo "Service Test" --success-target 99
func main(thisIsAContext *Context) {
	fmt.Println(hello) // line comment 3
}
`

	want := "// This is the package comment.\n" +
		"package main\n" +
		"\n" +
		"import (\n" +
		"\t. \"github.com/gin-gonic/gin\"\n" +
		"\n" +
		"\tprom \"github.com/autometrics-dev/autometrics-go/prometheus/autometrics\"\n" +
		")\n" +
		"\n" +
		"// This comment is associated with the main function.\n" +
		"//\n" +
		"//autometrics:inst --no-doc --slo \"Service Test\" --success-target 99\n" +
		"func main(thisIsAContext *Context) {\n" +
		"\tamCtx := prom.PreInstrument(prom.NewContext(\n" +
		"\t\tnil,\n" +
		"\t\tprom.WithTraceID(prom.DecodeString(thisIsAContext.GetString(\"autometricsTraceID\"))),\n" +
		"\t\tprom.WithSpanID(prom.DecodeString(thisIsAContext.GetString(\"autometricsSpanID\"))),\n" +
		"\t\tprom.WithConcurrentCalls(true),\n" +
		"\t\tprom.WithCallerName(true),\n" +
		"\t\tprom.WithSloName(\"Service Test\"),\n" +
		"\t\tprom.WithAlertSuccess(99),\n" +
		"\t)) //autometrics:shadow-ctx\n" +
		"\tdefer prom.Instrument(amCtx, nil) //autometrics:defer\n" +
		"\n" +
		"	fmt.Println(hello) // line comment 3\n" +
		"}\n"

	ctx, err := internal.NewGeneratorContext(autometrics.PROMETHEUS, defaultPrometheusInstanceUrl, false, true, false, false)
	if err != nil {
		t.Fatalf("error creating the generation context: %s", err)
	}

	actual, err := GenerateDocumentationAndInstrumentation(ctx, sourceCode, "main")
	if err != nil {
		t.Fatalf("error generating the documentation: %s", err)
	}

	assert.Equal(t, want, actual, "The generated source code is not as expected.")
}
