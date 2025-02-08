package testing

import "github.com/siyul-park/uniflow/pkg/process"

// TestCase represents a single test case function.
type TestCase func(t *Tester)

// TestCases is a map of test case names to their implementations.
type TestCases map[string]TestCase

// RunTestCases runs a collection of test cases within a suite.
func RunTestCases(cases TestCases) func(t *Tester) {
	return func(t *Tester) {
		for name, testCase := range cases {
			name := name
			testCase := testCase

			subTest := NewTester(t.Name() + "/" + name)

			// Copy parent's exit hook
			subTest.AddExitHook(process.ExitFunc(func(err error) {
				t.Close(err)
			}))

			testCase(subTest)
		}
	}
}

// Skip marks a test as skipped with an optional message.
func Skip(t *Tester, message string) {
	if message == "" {
		message = "test skipped"
	}
	t.Close(nil)
}

// Parallel marks a test to be run in parallel with other parallel tests.
func Parallel(t *Tester) {
	// TODO: Implement parallel test execution
}

// Cleanup registers a function to be called when the test completes.
func Cleanup(t *Tester, fn func()) {
	t.AddExitHook(process.ExitFunc(func(error) {
		fn()
	}))
}
