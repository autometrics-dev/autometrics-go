package doc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCommentDirective calls GenerateDocumentation on a
// decorated function, making sure that the autometrics
// directive adds a new comment section about autometrics.
func TestCommentDirective(t *testing.T) {
	sourceCode := `// This is the package comment.
package main

// This comment is associated with the main function.
//
//autometrics:doc
func main() {
	fmt.Println(hello) // line comment 3
}
`

	want := "// This is the package comment.\n" +
"package main\n" +
"\n" +
"// This comment is associated with the main function.\n" +
"//\n" +
"//\n" +
"//   autometrics:doc-start DO NOT EDIT\n" +
"//\n" +
"// # Autometrics\n" +
"//\n" +
"// ## Prometheus\n" +
"//\n" +
"// View the live metrics for the `main` function:\n" +
"//   - [Request Rate](http://localhost:9090/graph?g0.expr=%23+Rate+of+calls+to+the+%60main%60+function+per+second%2C+averaged+over+5+minute+windows%0A%0Asum+by+%28function%2C+module%29+%28rate%28function_calls_counter%7Bfunction%3D%22main%22%7D%5B5m%5D%29%29&g0.tab=0)\n" +
"//   - [Error Ratio](http://localhost:9090/graph?g0.expr=%23+Percentage+of+calls+to+the+%60main%60+function+that+return+errors%2C+averaged+over+5+minute+windows%0A%0Asum+by+%28function%2C+module%29+%28rate%28function_calls_counter%7Bfunction%3D%22main%22%2Cresult%3D%22error%22%7D%5B5m%5D%29%29&g0.tab=0)\n" +
"//   - [Latency (95th and 99th percentiles)](http://localhost:9090/graph?g0.expr=%23+95th+and+99th+percentile+latencies+%28in+seconds%29+for+the+%60main%60+function%0A%0Ahistogram_quantile%280.99%2C+sum+by+%28le%2C+function%2C+module%29+%28rate%28function_calls_duration_bucket%7Bfunction%3D%22main%22%7D%5B5m%5D%29%29%29+or+histogram_quantile%280.95%2C+sum+by+%28le%2C+function%2C+module%29+%28rate%28function_calls_duration_bucket%7Bfunction%3D%22main%22%7D%5B5m%5D%29%29%29&g0.tab=0)\n" +
"//   - [Concurrent Calls](http://localhost:9090/graph?g0.expr=%23+Concurrent+calls+to+the+%60main%60+function%0A%0Asum+by+%28function%2C+module%29+function_calls_concurrent%7Bfunction%3D%22main%22%7D&g0.tab=0)\n" +
"//\n" +
"// Or, dig into the metrics of *functions called by* `main`\n" +
"//   - [Request Rate](http://localhost:9090/graph?g0.expr=%23+Rate+of+function+calls+emanating+from+%60main%60+function+per+second%2C+averaged+over+5+minute+windows%0A%0Asum+by+%28function%2C+module%29+%28rate%28function_calls_counter%7Bcaller%3D%22main%22%7D%5B5m%5D%29%29&g0.tab=0)\n" +
"//   - [Error Ratio](http://localhost:9090/graph?g0.expr=%23+Percentage+of+function+emanating+from+%60main%60+function+that+return+errors%2C+averaged+over+5+minute+windows%0A%0Asum+by+%28function%2C+module%29+%28rate%28function_calls_counter%7Bcaller%3D%22main%22%2Cresult%3D%22error%22%7D%5B5m%5D%29%29&g0.tab=0)\n" +
"//\n" +
"//\n" +
"//   autometrics:doc-end DO NOT EDIT\n" +
"//\n" +
"//autometrics:doc\n" +
"func main() {\n" +
"	fmt.Println(hello) // line comment 3\n" +
"}\n"

	actual, err := GenerateDocumentation(sourceCode, NewPrometheusDoc())
	if err != nil {
		t.Fatalf("error generating the documentation: %s", err)
	}

	assert.Equal(t, want, actual, "The generated source code is not as expected.")
}

// TestCommentRefresh calls GenerateDocumentation on a
// decorated function that already has a comment, making sure that the autometrics
// directive only updates the comment section about autometrics.
func TestCommentRefresh(t *testing.T) {
	sourceCode := `// This is the package comment.
package main

// This comment is associated with the main function.
//
//   autometrics:doc-start
//
// Obviously not a good comment
//
//   autometrics:doc-end DO NOT EDIT
//
//autometrics:doc
func main() {
	fmt.Println(hello) // line comment 3
}
`

	want := "// This is the package comment.\n" +
"package main\n" +
"\n" +
"// This comment is associated with the main function.\n" +
"//\n" +
"//\n" +
"//\n" +
"//   autometrics:doc-start DO NOT EDIT\n" +
"//\n" +
"// # Autometrics\n" +
"//\n" +
"// ## Prometheus\n" +
"//\n" +
"// View the live metrics for the `main` function:\n" +
"//   - [Request Rate](http://localhost:9090/graph?g0.expr=%23+Rate+of+calls+to+the+%60main%60+function+per+second%2C+averaged+over+5+minute+windows%0A%0Asum+by+%28function%2C+module%29+%28rate%28function_calls_counter%7Bfunction%3D%22main%22%7D%5B5m%5D%29%29&g0.tab=0)\n" +
"//   - [Error Ratio](http://localhost:9090/graph?g0.expr=%23+Percentage+of+calls+to+the+%60main%60+function+that+return+errors%2C+averaged+over+5+minute+windows%0A%0Asum+by+%28function%2C+module%29+%28rate%28function_calls_counter%7Bfunction%3D%22main%22%2Cresult%3D%22error%22%7D%5B5m%5D%29%29&g0.tab=0)\n" +
"//   - [Latency (95th and 99th percentiles)](http://localhost:9090/graph?g0.expr=%23+95th+and+99th+percentile+latencies+%28in+seconds%29+for+the+%60main%60+function%0A%0Ahistogram_quantile%280.99%2C+sum+by+%28le%2C+function%2C+module%29+%28rate%28function_calls_duration_bucket%7Bfunction%3D%22main%22%7D%5B5m%5D%29%29%29+or+histogram_quantile%280.95%2C+sum+by+%28le%2C+function%2C+module%29+%28rate%28function_calls_duration_bucket%7Bfunction%3D%22main%22%7D%5B5m%5D%29%29%29&g0.tab=0)\n" +
"//   - [Concurrent Calls](http://localhost:9090/graph?g0.expr=%23+Concurrent+calls+to+the+%60main%60+function%0A%0Asum+by+%28function%2C+module%29+function_calls_concurrent%7Bfunction%3D%22main%22%7D&g0.tab=0)\n" +
"//\n" +
"// Or, dig into the metrics of *functions called by* `main`\n" +
"//   - [Request Rate](http://localhost:9090/graph?g0.expr=%23+Rate+of+function+calls+emanating+from+%60main%60+function+per+second%2C+averaged+over+5+minute+windows%0A%0Asum+by+%28function%2C+module%29+%28rate%28function_calls_counter%7Bcaller%3D%22main%22%7D%5B5m%5D%29%29&g0.tab=0)\n" +
"//   - [Error Ratio](http://localhost:9090/graph?g0.expr=%23+Percentage+of+function+emanating+from+%60main%60+function+that+return+errors%2C+averaged+over+5+minute+windows%0A%0Asum+by+%28function%2C+module%29+%28rate%28function_calls_counter%7Bcaller%3D%22main%22%2Cresult%3D%22error%22%7D%5B5m%5D%29%29&g0.tab=0)\n" +
"//\n" +
"//\n" +
"//   autometrics:doc-end DO NOT EDIT\n" +
"//\n" +
"//autometrics:doc\n" +
"func main() {\n" +
"	fmt.Println(hello) // line comment 3\n" +
"}\n"


	actual, err := GenerateDocumentation(sourceCode, NewPrometheusDoc())
	if err != nil {
		t.Fatalf("error generating the documentation: %s", err)
	}

	assert.Equal(t, want, actual, "The generated source code is not as expected.")
}
